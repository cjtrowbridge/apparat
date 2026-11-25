// Defines immutable execution inputs and the pinned worker identities used for claims.
// Every node in a roster validates the same digest before admitting an attempt.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/netip"
	"slices"
	"sort"
	"time"
)

type InferencePeer struct{ ID, Address string }
type InferenceRoster struct {
	Cluster string
	Peers   []InferencePeer
}

func (r InferenceRoster) IDs() []string {
	ids := make([]string, 0, len(r.Peers))
	for _, p := range r.Peers {
		ids = append(ids, p.ID)
	}
	return ids
}
func (r InferenceRoster) Revision() string {
	ids := r.IDs()
	slices.Sort(ids)
	return InferenceDigest(struct {
		Cluster string
		Members []string
	}{r.Cluster, ids})
}
func (r InferenceRoster) Validate(self string) error {
	if r.Cluster == "" || len(r.Peers) == 0 || !slices.Contains(r.IDs(), self) {
		return errors.New("incomplete inference roster")
	}
	seen := map[string]bool{}
	for _, p := range r.Peers {
		addr, err := netip.ParseAddrPort(p.Address)
		if err != nil || addr.Port() == 0 || len(p.ID) != 43 || seen[p.ID] {
			return errors.New("invalid inference peer")
		}
		seen[p.ID] = true
	}
	return nil
}

func InferenceDigest(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

type InferenceJob struct {
	Revision, Owner, User, Pool string
	Packet                      ChatRequest
}

func (j InferenceJob) Valid() bool {
	return ValidID(j.Packet.Job) && ValidID(j.Packet.Session) && ValidID(j.Packet.Turn) && j.Owner != "" && j.User != "" && j.Revision != "" && j.Packet.Model.Valid() && j.Packet.MaxOutputTokens > 0 && j.Packet.MaxOutputTokens <= 4096 && slices.ContainsFunc(DefaultPools(), func(p PoolDefault) bool {
		return p.Name == j.Pool
	})
}

type InferenceRecord struct {
	Job    InferenceJob
	Claim  InferenceClaim
	State  string
	Output *ChatOutput
}

func OrderInferenceJobs(jobs []InferenceJob) {
	ranks := map[string]int{}
	for i, p := range DefaultPools() {
		ranks[p.Name] = i
	}
	sort.Slice(jobs, func(i, j int) bool {
		a, b := jobs[i], jobs[j]
		if a.Pool != b.Pool {
			return ranks[a.Pool] < ranks[b.Pool]
		}
		return a.Packet.Job < b.Packet.Job
	})
}

// InferenceCommand carries either a user-authorized submission or a pinned worker transition.
type InferenceCommand struct {
	Chunk     GlobalChunk
	Pool      string
	Global    GlobalRequest
	Artifact  *GlobalArtifact
	Operation string
	Roster    InferenceRoster
	Job       InferenceJob
	Claim     InferenceClaim
	Output    *ChatOutput
	ID        string
}
type PeerObservation struct {
	Received bool
	Elapsed  time.Duration
	Observed int64
}
type InferenceStatus struct {
	Received                             int
	OldestQueuedMS                       int64
	Updating                             string
	Awaiting                             []string
	OfflineSince                         int64
	Busy                                 bool
	BlockedSince                         int64
	Peers                                map[string]PeerObservation
	Validators                           int
	Queued, Owned, Dispatching, Finished int
	Timing                               AgreementTiming
}

type InferenceReply struct {
	Executor InferencePeer `json:",omitzero"`
	Protocol int           `json:",omitempty"`
	Projects []ChatProject
	Cursor   string
	Chunk    *GlobalChunk
	Tasks    []PoolTaskRow
	Task     *InferenceRecord
	Headers  []ChatHeader
	Artifact *GlobalArtifact
	Roster   InferenceRoster
	Status   InferenceStatus
	State    string
	Answer   ClaimAnswer
	Output   *ChatOutput
}

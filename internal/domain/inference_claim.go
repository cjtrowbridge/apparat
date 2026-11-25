// Owns one inference validator's in-memory reservations for an agreed roster.
// The executor participates in every round, including its own, before dispatch.
package domain

import (
	"cmp"
	"errors"
	"slices"
	"sync"
)

type InferenceClaim struct {
	Revision, Job, Digest, Executor, Attempt string
}
type ClaimAnswer struct {
	State  string
	Holder InferenceClaim
}
type reservation struct {
	claim     InferenceClaim
	state     string
	withdrawn map[string]string
	digest    string
}
type ClaimTable struct {
	mu       sync.Mutex
	revision string
	members  map[string]bool
	ready    bool
	jobs     map[string]*reservation
}

func NewClaimTable(revision string, members []string) (*ClaimTable, error) {
	roster := slices.Clone(members)
	slices.Sort(roster)
	if revision == "" || len(roster) == 0 || roster[0] == "" || len(slices.Compact(slices.Clone(roster))) != len(roster) {
		return nil, errors.New("invalid inference roster")
	}
	membersByID := make(map[string]bool, len(roster))
	for _, id := range roster {
		membersByID[id] = true
	}
	return &ClaimTable{revision: revision, members: membersByID, jobs: make(map[string]*reservation)}, nil
}

// Admit binds the immutable digest before any executor can reserve this job.
func (t *ClaimTable) Admit(job, digest string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !ValidID(job) || digest == "" {
		return errors.New("invalid inference job")
	}
	r := t.jobs[job]
	if r != nil {
		if r.digest != digest {
			return errors.New("conflicting job digest")
		}
		return nil
	}
	t.jobs[job] = &reservation{digest: digest, withdrawn: make(map[string]string)}
	return nil
}

// Recovered is called after the durable dispatch ledger and peer reservations are reconciled.
func (t *ClaimTable) Recovered() {
	t.mu.Lock()
	t.ready = true
	t.mu.Unlock()
}

// Transition takes an already authenticated caller identity. The caller must be the executor.
// Withdrawals tombstone an abandoned attempt even when they arrive before its request.
func (t *ClaimTable) Transition(caller, operation string, c InferenceClaim) ClaimAnswer {
	t.mu.Lock()
	defer t.mu.Unlock()
	member := t.Member(caller)
	r := t.jobs[c.Job]
	if !member || caller != c.Executor || c.Revision != t.revision || !ValidID(c.Attempt) {
		return ClaimAnswer{State: "invalid"}
	}
	if r == nil {
		return ClaimAnswer{State: "missing"}
	}
	if r.digest != c.Digest {
		return ClaimAnswer{State: "invalid"}
	}
	if !t.ready {
		return ClaimAnswer{State: "recovering"}
	}
	if r.withdrawn[caller] >= c.Attempt && r.claim != c {
		return ClaimAnswer{State: "withdrawn"}
	}
	same := r.claim == c
	switch operation {
	case "confirm":
		if r.state == "" {
			r.claim = c
			r.state = "reserved"
		}
	case "withdraw":
		if same && r.state != "reserved" {
			return ClaimAnswer{State: r.state, Holder: r.claim}
		}
		r.withdrawn[caller] = c.Attempt
		if same {
			r.claim = InferenceClaim{}
			r.state = ""
		}
		return ClaimAnswer{State: "withdrawn"}
	case "own":
		if same && r.state == "reserved" {
			r.state = "owned"
		}
	case "finish":
		if same && r.state == "owned" {
			r.state = "finished"
		}
	case "inspect":
	default:
		return ClaimAnswer{State: "invalid"}
	}
	return ClaimAnswer{State: r.state, Holder: r.claim}
}

// CompareClaim orders contenders independently of request arrival order.
func CompareClaim(a, b InferenceClaim) int {
	order := cmp.Compare(a.Attempt, b.Attempt)
	if order != 0 {
		return order
	}
	return cmp.Compare(a.Executor, b.Executor)
}

// Member checks the immutable identity set without acquiring the reservation mutex.
func (t *ClaimTable) Member(id string) bool {
	return t.members[id]
}

func (t *ClaimTable) Reserved() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, r := range t.jobs {
		if r.state == "reserved" {
			return true
		}
	}
	return false
}

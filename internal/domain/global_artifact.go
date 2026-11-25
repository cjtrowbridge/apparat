// Identifies replicated global conversations and their single ordered transcript writer.
// The session writer is independent of the worker that wins an inference step.
package domain

import (
	"crypto/sha256"
	"encoding/binary"
	"slices"
)

type GlobalArtifact struct {
	User    string
	Version int64
	Session ChatSession
}
type GlobalRequest struct {
	Session, Turn, Input, Title, Base, Cursor string
	Model                                     ModelRef
}

func (r InferenceRoster) GlobalWriter(user, session string) string {
	ids := r.IDs()
	if len(ids) == 0 {
		return ""
	}
	slices.Sort(ids)
	h := sha256.Sum256([]byte(user + "/" + session))
	return ids[binary.BigEndian.Uint64(h[:8])%uint64(len(ids))]
}

type PlanningView struct {
	SyncReason   string
	Headers      []ChatHeader
	Session      *ChatSession
	Acknowledged bool
}

const GlobalChunkBytes = 128 << 10

type GlobalChunk struct {
	User          string
	Revision      int64
	Total, Offset int
	Bytes         []byte
}

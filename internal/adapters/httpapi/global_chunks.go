// Transfers global transcripts in bounded, revision-pinned reads over authenticated APIs.
// A concurrent edit makes the reader retry the new revision from its first byte.
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

type chunkReader func(context.Context, domain.InferenceCommand) (domain.InferenceReply, error)

func ReadGlobal(ctx context.Context, id, user string, read chunkReader) (domain.InferenceReply, error) {
	q := domain.InferenceCommand{Operation: "global-chunk", Global: domain.GlobalRequest{Session: id}, Chunk: domain.GlobalChunk{User: user}}
	var raw bytes.Buffer
	for ctx.Err() == nil {
		v, e := read(ctx, q)
		if e != nil {
			return domain.InferenceReply{}, e
		}
		part := v.Chunk
		if part == nil || part.Revision < 1 || (q.Chunk.Revision != 0 && part.Revision != q.Chunk.Revision) || part.Offset != raw.Len() || len(part.Bytes) == 0 || len(part.Bytes) > domain.GlobalChunkBytes || part.Total < raw.Len()+len(part.Bytes) || part.User == "" || (user != "" && user != part.User) {
			return domain.InferenceReply{}, errors.New("invalid global transcript chunk")
		}
		raw.Write(part.Bytes)
		if raw.Len() == part.Total {
			a := domain.GlobalArtifact{User: part.User, Version: part.Revision}
			e = json.Unmarshal(raw.Bytes(), &a.Session)
			if e != nil || a.Session.ID != id || a.Session.Scope != domain.GlobalChat || a.Session.Validate() != nil {
				return domain.InferenceReply{}, errors.New("invalid global transcript")
			}
			return domain.InferenceReply{Artifact: &a}, nil
		}
		q.Chunk.Offset = raw.Len()
		q.Chunk.Revision = part.Revision
	}
	return domain.InferenceReply{}, ctx.Err()
}

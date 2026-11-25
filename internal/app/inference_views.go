// Reads the user's model jobs through a pinned worker and links them from the pool browser.
// Prompt and result contents appear only on the explicitly opened task detail.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (n *Runtime) inferenceRows(ctx context.Context, w domain.ClusterDevice, pool string, seen map[string]bool) []domain.PoolTaskRow {
	if w.Inference != nil && w.Inference.OfflineSince > 0 {
		return nil
	}
	c, err := n.workspaceClient(ctx, "worker-"+w.Fingerprint)
	if err != nil {
		return nil
	}
	var out []domain.PoolTaskRow
	cursor := ""
	for ctx.Err() == nil {
		v, e := c.Inference(ctx, domain.InferenceCommand{Operation: "tasks", ID: cursor, Pool: pool})
		if e != nil {
			return out
		}
		for _, t := range v.Tasks {
			if seen[t.ID] {
				continue
			}
			seen[t.ID] = true
			t.Host = w.DisplayName()
			t.Worker = w.Fingerprint
			t.Inference = true
			out = append(out, t)
		}
		if len(v.Tasks) < 64 {
			break
		}
		cursor = v.Tasks[len(v.Tasks)-1].ID
	}
	return out
}
func (n *Runtime) inferenceTaskPage(ctx context.Context, r domain.ClusterRoute) (p domain.ClusterPage, err error) {
	c, err := n.workspaceClient(ctx, "worker-"+r.Worker)
	if err != nil {
		return p, err
	}
	v, err := c.Inference(ctx, domain.InferenceCommand{Operation: "task", ID: r.Target})
	if err != nil || v.Task == nil {
		return p, err
	}
	row := v.Task
	j := row.Job
	p.Title = "Inference task"
	input, _ := json.MarshalIndent(j.Packet, "", "  ")
	output, _ := json.MarshalIndent(row.Output, "", "  ")
	p.Detail = fmt.Sprintf("%s · %s\nModel %s\nJob %s\nSession %s\nExecutor %s\n\nInput\n%s\n\nOutput\n%s", j.Pool, row.State, j.Packet.Model.Name, j.Packet.Job, j.Packet.Session, row.Claim.Executor, input, output)
	return p, nil
}

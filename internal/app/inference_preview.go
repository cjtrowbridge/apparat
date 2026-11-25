// Owns volatile output for each local provider invocation. Viewers read by byte offset.
// The dispatch context belongs to the worker and survives subscriber disconnection.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/chatprovider"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
	"sync"
)

type liveGeneration struct {
	mu          sync.Mutex
	user, stage string
	text        strings.Builder
	finished    bool
}

func (p *inferencePool) dispatchPreview(ctx context.Context, r domain.InferenceRecord, provider domain.ChatProvider) (*domain.ChatOutput, error) {
	native, ok := provider.(*chatprovider.Completions)
	if !ok {
		return provider.Dispatch(ctx, r.Job.Packet)
	}
	g := &liveGeneration{user: r.Job.User, stage: "Evaluating prompt"}
	job := r.Job.Packet.Job
	p.previews.Store(job, g)
	defer func() {
		g.mu.Lock()
		g.finished = true
		g.mu.Unlock()
		p.previews.Delete(job)
	}()
	native.Preview = func(stage, text string) {
		g.mu.Lock()
		g.stage = stage
		g.text.WriteString(text)
		g.mu.Unlock()
	}
	return native.Dispatch(ctx, r.Job.Packet)
}
func (p *inferencePool) OpenPreview(user, job string) (func(int) domain.PreviewFrame, error) {
	value, ok := p.previews.Load(job)
	if !ok {
		return nil, httpapi.ErrDenied
	}
	g := value.(*liveGeneration)
	if user != g.user {
		return nil, httpapi.ErrDenied
	}
	return func(offset int) domain.PreviewFrame {
		g.mu.Lock()
		defer g.mu.Unlock()
		return domain.PreviewFrame{Stage: g.stage, Text: g.text.String()[offset:], Done: g.finished}
	}, nil
}

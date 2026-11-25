//go:build gui

// Watches the open conversation's pending job. Deltas update two widgets in place
// so incoming tokens leave keyboard focus, the composer and scroll ownership intact.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"strings"
	"time"
)

type previewReader interface {
	PlanningUser() string
	WatchChatPreview(context.Context, string, string, func(domain.PreviewFrame)) error
}
type chatPreviewUI struct {
	user, job, stage, text string
	updates                chan domain.PreviewFrame
	cancel                 context.CancelFunc
	widgetJob              string
	body                   *widget.Text
	status                 *widget.Button
	fallback               string
	parent                 *widget.Container
}

func (p *chatPreviewUI) stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.cancel = nil
	p.job = ""
	p.user = ""
	p.text = ""
	p.stage = ""
	p.updates = nil
}
func (g *Game) visiblePreviewJob() string {
	if g.actions.mode != "" {
		return ""
	}
	if g.shell.Snapshot().ActiveTab().ID() == hud.TabGlobalSession {
		s := g.planning.current
		if g.planning.page == "" && s != nil && len(s.Turns) > 0 {
			return s.Turns[len(s.Turns)-1].PendingJob()
		}
	} else if g.inWorkspace() {
		w := &g.projects.workspace
		if w.mode == "conversation" && w.agentThread && len(w.choices) > 0 {
			last := w.choices[len(w.choices)-1]
			if last.State == "pending" {
				return last.Job
			}
		}
	}
	return ""
}
func (g *Game) pollChatPreview() {
	p := &g.preview
	reader, ok := g.runtimeInfo.Queries.(previewReader)
	job := g.visiblePreviewJob()
	if !ok || job == "" {
		if p.cancel != nil {
			p.stop()
			g.paintChatPreview()
			p.body = nil
			p.status = nil
			p.parent = nil
		}
		return
	}
	user := reader.PlanningUser()
	if user != p.user || job != p.job {
		p.stop()
		g.paintChatPreview()
		if user == "" || g.projects.ctx == nil {
			return
		}
		ctx, cancel := context.WithCancel(g.projects.ctx)
		updates := make(chan domain.PreviewFrame, 1)
		p.cancel, p.job, p.user, p.updates = cancel, job, user, updates
		g.projects.workers.Go(func() {
			for ctx.Err() == nil {
				var text strings.Builder
				err := reader.WatchChatPreview(ctx, user, job, func(frame domain.PreviewFrame) {
					text.WriteString(frame.Text)
					frame.Text = text.String()
					select {
					case <-updates:
					default:
					}
					updates <- frame
				})
				if err == nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
				}
			}
		})
	}
	select {
	case f := <-p.updates:
		p.text, p.stage = f.Text, f.Stage
		if f.Done {
			p.stage = "Finishing model step"
		}
		g.paintChatPreview()
	default:
	}
}
func (g *Game) attachChatPreview(b *widget.Container, job string, face *widget.Button, fallback string) {
	p := &g.preview
	width := g.hudPreferredWidth()
	if g.inWorkspace() {
		width = g.detailPanePreferredWidth()
	}
	p.body = widget.NewText(widget.TextOpts.Text("", g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor.Idle), widget.TextOpts.MaxWidth(float64(max(60, min(540, width*4/5-24)))))
	p.widgetJob, p.parent, p.status, p.fallback = job, b, face, fallback
	b.AddChild(p.body)
	g.paintChatPreview()
}
func (g *Game) paintChatPreview() {
	p := &g.preview
	if p.body == nil {
		return
	}
	text, stage := "", p.fallback
	if p.widgetJob == p.job && p.stage != "" {
		text, stage = p.text, p.stage
	}
	p.body.Label = text
	vis := widget.Visibility_Show
	if text == "" {
		vis = widget.Visibility_Hide
	}
	p.body.GetWidget().SetVisibility(vis)
	p.status.SetText(stage)
	p.parent.RequestRelayout()
	if g.nav.body != nil {
		layout, ok := g.nav.body.(widget.Relayoutable)
		if ok {
			layout.RequestRelayout()
		}
	}
}

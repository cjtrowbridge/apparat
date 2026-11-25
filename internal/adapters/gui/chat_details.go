//go:build gui

// Reveals actions for a single conversation bubble and presents its per-call counters.
// Copy uses the original message, independently of wrapping and display typography.
package gui

import (
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"maps"
	"slices"
	"strings"
	"time"
)

type bubbleDetail struct {
	id    string
	user  bool
	steps []domain.ChatMeasurement
}
type bubbleTools struct {
	row        *widget.Container
	info       bubbleDetail
	copy, open func()
}
type bubbleSelection struct {
	current *bubbleTools
	pinned  bool
}

func (s *bubbleSelection) show(next *bubbleTools, pin bool) {
	if s.current != nil && s.current != next {
		s.current.row.GetWidget().SetVisibility(widget.Visibility_Hide)
	}
	s.current = next
	s.pinned = pin
	if next != nil {
		next.row.GetWidget().SetVisibility(widget.Visibility_Show)
	}
}

func (g *Game) bubbleControls(b *widget.Container, source, label string, user, waiting bool, detail bubbleDetail) {
	detail.user = user
	if user {
		detail.id += "/user"
	} else {
		detail.id += "/reply"
	}
	tools := &bubbleTools{info: detail, copy: func() {
		g.copyDocument(source)
	}}
	tools.open = func() {
		g.planning.info = detail
		if g.inWorkspace() {
			g.projects.workspace.mode = "bubble-info"
			g.layoutDirty = true
			return
		}
		g.planningPage("bubble-info")
	}
	selectBubble := func() {
		if g.bodySelectionSuppressed() {
			return
		}
		s := &g.planning.bubbles
		if s.current == tools && s.pinned {
			s.show(nil, false)
		} else {
			s.show(tools, true)
		}
	}
	color := panelColor
	if user {
		color = activeTabColor
	}
	skin := tabButtonImage()
	skin.Idle = uiimage.NewNineSliceColor(color)
	face := widget.NewButton(widget.ButtonOpts.Text(label, g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor), widget.ButtonOpts.TextPadding(&widget.Insets{}), widget.ButtonOpts.Image(skin), widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
		selectBubble()
	}))
	content := controlRow(b)
	if waiting {
		content.AddChild(planningGear())
	}
	content.AddChild(face)
	tools.row = controlRow(b)
	tools.row.AddChild(g.planningIcon("1f4cb", "Copy message", tools.copy), g.planningIcon("2139", "Message information", tools.open))
	tools.row.GetWidget().SetVisibility(widget.Visibility_Hide)
	b.GetWidget().CursorEnterEvent.AddHandler(func(any) {
		s := &g.planning.bubbles
		if s.current != tools {
			s.show(tools, false)
		}
	})
	b.GetWidget().CursorExitEvent.AddHandler(func(any) {
		s := &g.planning.bubbles
		if s.current == tools && !s.pinned {
			s.show(nil, false)
		}
	})
}

func (g *Game) bubbleActions(add registerAction) {
	t := g.planning.bubbles.current
	if t == nil {
		return
	}
	global := g.shell.Snapshot().ActiveTab().ID() == hud.TabGlobalSession && g.planning.page == "" && g.planning.current != nil && strings.HasPrefix(t.info.id, g.planning.current.ID+"/")
	w := &g.projects.workspace
	project := g.inWorkspace() && w.mode == "conversation" && strings.HasPrefix(t.info.id, w.id+"/"+w.thread+"/")
	if global || project {
		add("chat-message", t.info.id, "copy", "Copy message", "Chat / Selected message / Copy", t.copy)
		add("chat-message", t.info.id, "info", "Message information", "Chat / Selected message / Information", t.open)
	}
}

func (g *Game) bubbleTable(c *widget.Container, info bubbleDetail) {
	c.AddChild(g.detailText("Message information"))
	if info.user && len(info.steps) == 0 {
		c.AddChild(g.detailText("No inference measurements have been recorded for this prompt."))
		return
	}
	if len(info.steps) == 0 {
		c.AddChild(g.detailText("No inference measurements were recorded for this reply."))
		return
	}
	for _, s := range info.steps {
		c.AddChild(g.detailText(fmt.Sprintf("Model call %d · %s", s.Step, s.Finish)))
		m := s.Metrics
		if len(m) == 0 {
			c.AddChild(g.detailText("No provider measurements recorded."))
			continue
		}
		grid := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Spacing(12, 4))))
		c.AddChild(grid)
		row := func(k, v string) {
			txt := widget.NewText(widget.TextOpts.Text(k, g.formLabelFace, mutedTextColor), widget.TextOpts.MaxWidth(float64(max(86, min(320, g.detailPanePreferredWidth()/2-24)))))
			grid.AddChild(txt, g.detailText(v))
		}
		for _, k := range []string{"prompt_eval", "eval"} {
			count, ok := m[k+"_count"]
			if k == "prompt_eval" {
				count = max(0, count-m["prompt_eval_cached_count"])
			}
			ns := m[k+"_duration"]
			if ok && ns > 0 {
				title := "Eval"
				if k == "prompt_eval" {
					title = "Prefill"
				}
				row(title+" tok/s", fmt.Sprintf("%.2f", count*1e9/ns))
			}
		}
		for _, k := range slices.Sorted(maps.Keys(m)) {
			v := fmt.Sprintf("%g", m[k])
			if strings.HasSuffix(k, "_duration") {
				v = time.Duration(m[k]).String()
			}
			row(strings.ReplaceAll(k, "_", " "), v)
		}
	}
}

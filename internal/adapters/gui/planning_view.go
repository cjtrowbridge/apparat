//go:build gui

// Builds the planning history, session toolbar and fixed prompt around a scrollable transcript.
// The rotating indicator distinguishes a locally staged turn from an acknowledged submission.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/domain"
	"sort"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/hud"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func planningColumn() *widget.Container {
	return widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(bodyGap))))
}

func (g *Game) planningScroll(content *widget.Container) *widget.ScrollContainer {
	pane := widget.NewScrollContainer(widget.ScrollContainerOpts.Content(content), widget.ScrollContainerOpts.Image(createScrollContainerImage()), widget.ScrollContainerOpts.StretchContentWidth(), widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: 0})))
	g.registerVerticalScroll(pane)
	return pane
}

func (g *Game) planningIcon(icon, label string, fn func()) *widget.Button {
	b := g.hudIcon(icon, fn)
	b.GetWidget().CustomData = label
	return b
}

func (g *Game) planningTitle(title string, width int) string {
	face := *g.theme.ButtonTheme.TextFace
	size, _ := text.Measure(title, face, 0)
	if size <= float64(width) {
		return title
	}
	runes := []rune(title)
	end := sort.Search(len(runes), func(i int) bool {
		w, _ := text.Measure(string(runes[:i+1])+"...", face, 0)
		return w > float64(width)
	})
	return string(runes[:end]) + "..."
}

func (g *Game) planningHeader(s *planningSession) *widget.Container {
	bar := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)))
	left := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(8))))
	left.AddChild(g.planningIcon("1f519", "Back to sessions", g.planningBack), g.planningIcon("2795", "New session", g.newPlanning))
	title := g.planningTitle(s.Description, max(1, g.hudPreferredWidth()-2*(48+8+48+8)))
	label := widget.NewText(widget.TextOpts.Text(title, g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor.Idle), widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{HorizontalPosition: widget.AnchorLayoutPositionCenter, VerticalPosition: widget.AnchorLayoutPositionCenter})))
	edit := g.planningIcon("270f", "Edit session title", func() {
		g.planningPage("rename")
	})
	edit.GetWidget().LayoutData = widget.AnchorLayoutData{HorizontalPosition: widget.AnchorLayoutPositionEnd}
	bar.AddChild(left, label, edit)
	return bar
}

func (g *Game) globalSessionBody(_ hud.Tab) *widget.Container {
	p := &g.planning
	if p.current == nil {
		g.newPlanning()
	}
	s := p.current
	if s == nil {
		return planningColumn()
	}
	if p.page == "bubble-info" {
		c := planningColumn()
		back := g.planningIcon("1f519", "Back to conversation", g.planningBack)
		g.bubbleTable(c, p.info)
		page := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}))))
		page.AddChild(back, g.planningScroll(c))
		return g.globalChatDetail(s, page)
	}
	if p.page == "model-picker" {
		return g.globalModelPage(s)
	}
	if p.page == "history" {
		body := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}))))
		head := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(bodyGap))))
		head.AddChild(g.planningIcon("2795", "New session", g.newPlanning), g.detailText("Session history"))
		list := planningColumn()
		for i := len(p.sessions) - 1; i >= 0; i-- {
			session := p.sessions[i]
			name := session.Description + " · " + session.Started.Local().Format("Jan 2 15:04:05")
			if len(session.Turns) == 0 {
				name += " · Draft"
			}
			g.projectButton(list, wrapSelectorLabel(name, *g.theme.ButtonTheme.TextFace, g.hudPreferredWidth()-32), func() {
				g.openPlanning(session)
			})
		}
		body.AddChild(head, g.planningScroll(list))
		return body
	}
	if p.page == "rename" {
		body := planningColumn()
		body.AddChild(g.planningIcon("1f519", "Back to conversation", g.planningBack), g.detailText("Edit session title"))
		field := g.projectField(body, "Title", &p.title)
		p.saveTitle = g.projectButton(body, "Save title", func() {
			title := strings.TrimSpace(p.title)
			if title == "" || len(title) > 512 || strings.ContainsAny(title, "\r\n") {
				return
			}
			if len(s.Turns) > 0 && !g.savePlanningOperation("global-rename", domain.GlobalRequest{Session: s.ID, Title: title, Base: s.Description}) {
				return
			}
			s.Description = title
			g.openPlanning(s)
		})
		p.titleField = field
		return body
	}
	if p.pane != nil {
		s.scroll = p.pane.ScrollTop
	}
	compact := g.height < 590
	body := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}), widget.GridLayoutOpts.Spacing(0, bodyGap))))
	conversation := planningColumn()
	if p.sync.notice != "" {
		conversation.AddChild(g.hudText(p.sync.notice))
	}
	if s.loading {
		conversation.AddChild(g.hudText("Loading conversation…"))
	}
	p.bubbles = bubbleSelection{}
	for _, turn := range s.Turns {
		info := bubbleDetail{id: s.ID + "/" + turn.ID, steps: turn.Measurements()}
		g.chatBubble(conversation, turn.Input, true, false, info)
		switch {
		case turn.Reply != "":
			g.chatBubble(conversation, turn.Reply, false, false, info)
		case turn.Stopped != "":
			g.chatBubble(conversation, turn.Stopped, false, false, info)
		case s.acknowledged:
			label := "Waiting for reply"
			if s.syncReason != "" {
				label = s.syncReason
			}
			g.chatBubble(conversation, label, false, true, info)
		default:
			status := "Waiting to sync · saved on this device"
			if s.syncReason != "" {
				status += "\n" + s.syncReason
			}
			g.chatBubble(conversation, status, false, true, info)
		}
	}
	p.pane = g.planningScroll(conversation)
	p.pane.ScrollTop = s.scroll
	form := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(bodyGap), widget.RowLayoutOpts.Padding(&widget.Insets{Right: 2*pttRadius + pttMargin + pttTravel}))))
	if compact {
		form = widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{false}), widget.GridLayoutOpts.Spacing(bodyGap, 0), widget.GridLayoutOpts.Padding(&widget.Insets{Right: 2*pttRadius + pttMargin + pttTravel}))))
	}
	surface := g.multilineInput(form, "", &g.globalDraft)
	surface.submit = g.submitPlanning
	if p.picker.restore {
		surface.caret = min(p.picker.caret, len(surface.body))
		surface.anchor = min(p.picker.anchor, len(surface.body))
		p.picker.restore = false
	}
	if compact {
		surface.GetWidget().MinHeight = 2*surface.height + 8
	} else {
		surface.GetWidget().MinHeight = 3*surface.height + 8
	}
	g.globalField = surface.TextInput
	p.send = g.projectButton(form, "Send", g.submitPlanning)
	if compact {
		p.send.GetWidget().LayoutData = widget.GridLayoutData{MaxHeight: 44, VerticalPosition: widget.GridLayoutPositionCenter}
	}
	g.updatePlanningControls()
	body.AddChild(g.planningHeader(s), p.pane, form)
	return body
}

func (g *Game) chatBubble(parent *widget.Container, body string, user, waiting bool, details ...bubbleDetail) {
	row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})))
	position := widget.AnchorLayoutPositionStart
	color := panelColor
	if user {
		position = widget.AnchorLayoutPositionEnd
		color = activeTabColor
	}
	b := widget.NewContainer(widget.ContainerOpts.BackgroundImage(uiimage.NewNineSliceColor(color)), widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(4), widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8}))), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{HorizontalPosition: position})))

	lines := strings.Split(body, "\n")
	paneWidth := g.hudPreferredWidth()
	if g.inWorkspace() {
		paneWidth = g.detailPanePreferredWidth()
	}
	width := max(60, min(540, (paneWidth*4/5)-24))
	for i, line := range lines {
		lines[i] = wrapSelectorLabel(line, *g.theme.ButtonTheme.TextFace, width)
	}
	label := strings.Join(lines, "\n")
	if len(details) > 0 {
		g.bubbleControls(b, body, label, user, waiting, details[0])
	} else {
		content := controlRow(b)
		if waiting {
			content.AddChild(planningGear())
		}
		content.AddChild(widget.NewText(widget.TextOpts.Text(label, g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor.Idle)))
	}
	row.AddChild(b)
	parent.AddChild(row)
}

type spinningGear struct{ *widget.Graphic }

var gearTexture *ebiten.Image

func planningGear() *spinningGear {
	if gearTexture == nil {
		source := ebiten.NewImageFromImage(appIconImages()[0])
		gearTexture = ebiten.NewImage(24, 24)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Scale(24/float64(source.Bounds().Dx()), 24/float64(source.Bounds().Dy()))
		gearTexture.DrawImage(source, op)
		source.Deallocate()
	}
	return &spinningGear{widget.NewGraphic(widget.GraphicOpts.Image(gearTexture))}
}
func (s *spinningGear) Render(dst *ebiten.Image) {
	r := s.GetWidget().Rect
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-12, -12)
	op.GeoM.Rotate(float64(time.Now().UnixMilli()%6000) * 6.283185307 / 6000)
	op.GeoM.Translate(float64(r.Min.X+12), float64(r.Min.Y+r.Dy()/2))
	dst.DrawImage(gearTexture, op)
}

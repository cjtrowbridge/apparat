//go:build gui

// Builds the planning history, session toolbar and fixed prompt around a scrollable transcript.
// Local submissions remain visibly unsynchronized until a future queue supplies acknowledgement.
package gui

import (
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
			s.Description = title
			g.openPlanning(s)
		})
		p.titleField = field
		return body
	}
	if p.pane != nil {
		s.scroll = p.pane.ScrollTop
	}
	bottom := pttMargin + 2*pttRadius + pttTravel
	compact := g.height < 590
	if compact {
		bottom = 0
	}
	body := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}), widget.GridLayoutOpts.Spacing(0, bodyGap), widget.GridLayoutOpts.Padding(&widget.Insets{Bottom: bottom}))))
	conversation := planningColumn()
	for _, turn := range s.Turns {
		g.chatBubble(conversation, turn.Input, true, false)
		switch {
		case turn.Reply != "":
			g.chatBubble(conversation, turn.Reply, false, false)
		case turn.Stopped != "":
			g.chatBubble(conversation, turn.Stopped, false, false)
		case s.acknowledged:
			g.chatBubble(conversation, "Waiting for reply", false, true)
		default:
			g.chatBubble(conversation, "Waiting to sync · local only", false, true)
		}
	}
	p.pane = g.planningScroll(conversation)
	p.pane.ScrollTop = s.scroll
	form := planningColumn()
	if compact {
		form = widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Padding(&widget.Insets{Right: pttMargin + 2*pttRadius + pttTravel}))))
	}
	surface := g.multilineInput(form, "Prompt", &g.globalDraft)
	if compact {
		surface.GetWidget().MinHeight = 2*surface.height + 8
	} else {
		surface.GetWidget().MinHeight = 3*surface.height + 8
	}
	g.globalField = surface.TextInput
	p.send = g.projectButton(form, "Send", g.submitPlanning)
	g.updatePlanningControls()
	body.AddChild(g.planningHeader(s), p.pane, form)
	return body
}

func (g *Game) chatBubble(parent *widget.Container, body string, user, waiting bool) {
	row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})))
	position := widget.AnchorLayoutPositionStart
	color := panelColor
	if user {
		position = widget.AnchorLayoutPositionEnd
		color = activeTabColor
	}
	b := widget.NewContainer(widget.ContainerOpts.BackgroundImage(uiimage.NewNineSliceColor(color)), widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(8), widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8}))), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{HorizontalPosition: position})))
	if waiting {
		b.AddChild(planningGear())
	}
	lines := strings.Split(body, "\n")
	width := max(60, (g.hudPreferredWidth()*4/5)-24)
	for i, line := range lines {
		lines[i] = wrapSelectorLabel(line, *g.theme.ButtonTheme.TextFace, width)
	}
	b.AddChild(widget.NewText(widget.TextOpts.Text(strings.Join(lines, "\n"), g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor.Idle)))
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

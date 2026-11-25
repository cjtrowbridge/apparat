//go:build gui

// Resizes the chat split pane using mouse or touch gesture ownership.
// A chat remembers its divider position independently of the outer project selector.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"time"
)

func (g *Game) closeChatPermissionPage() bool {
	_, t := g.currentChatTerminal()
	if t == nil || !t.permissionPage {
		return false
	}
	t.permissionPage = false
	if t.modal != nil {
		t.modal.Close()
	}
	t.modal = nil
	if t.permissionButton != nil {
		g.ui.SetFocusedWidget(t.permissionButton)
	}
	return true
}
func (g *Game) chatSplitAvailable(a modelAddress) int {
	if a.project == "" {
		return g.hudPreferredWidth() - 24
	}
	return g.detailPanePreferredWidth() - 32
}
func (g *Game) chatSplitWidth(width int, t *chatTerminalPane) int {
	left := t.split
	if left == 0 {
		left = (width - 48) / 2
	}
	return max(44, min(left, width-96))
}
func (g *Game) chatDivider() *widget.Button {
	b := widget.NewButton(widget.ButtonOpts.Text("↔", g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor), widget.ButtonOpts.Image(g.theme.ButtonTheme.Image), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(44, 44)))
	g.planning.divider = b.GetWidget()
	return b
}
func (g *Game) beginChatDivider(x, y int, touch bool, id ebiten.TouchID) bool {
	d := g.planning.divider
	if d == nil || !imagePoint(x, y).In(d.Rect) || !g.claimGesture(gestureChatDivider, touch, id) {
		return false
	}
	a, t := g.currentChatTerminal()
	if t == nil {
		g.releaseGesture(gestureChatDivider)
		return false
	}
	g.gesture.startX = x - g.chatSplitWidth(g.chatSplitAvailable(a), t)
	return true
}
func (g *Game) moveChatDivider(x int) {
	a, t := g.currentChatTerminal()
	if t == nil {
		return
	}
	old := g.chatSplitWidth(g.chatSplitAvailable(a), t)
	t.split = x - g.gesture.startX
	if old != g.chatSplitWidth(g.chatSplitAvailable(a), t) {
		g.rebuildUI(g.shell.Snapshot())
	}
}
func (g *Game) chatDividerMouse(x, y int) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.beginChatDivider(x, y, false, 0)
	}
	if g.gesture.Owner != gestureChatDivider {
		return
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.releaseGesture(gestureChatDivider)
		return
	}
	g.moveChatDivider(x)
}

func (game *Game) chatLayoutInsets(snapshot hud.Snapshot, title string) (string, int) {
	bottomInset := diagnosticsHeight
	if snapshot.ActiveTab().ID() == hud.TabGlobalSession && game.height >= 590 {
		bottomInset += pttTravel + pttRadius
	}
	if game.inWorkspace() {
		i := game.selectedSections[hud.TabProjects] - len(projectEntrySections)
		p := game.projects.rows[i]
		title = game.projectTitle(p) + " · " + p.Branch
		if !game.debugOverlayOpen {
			bottomInset = windowMargin
		}
	}
	return title, bottomInset
}

func (g *Game) terminalActions(add registerAction) {
	a, t := g.currentChatTerminal()
	if t == nil || g.modelPageOpen() {
		return
	}
	owner := a.project + "/" + a.session
	if t.permissionPage {
		add("chat-terminal", owner, "back", "Back to chat", "Chat / Permissions / Back", func() {
			g.closeChatPermissionPage()
		})
		return
	}
	add("chat-terminal", owner, "permissions", "Chat permissions", "Chat / Permissions", func() {
		g.openChatPermissionModal(a, t)
	})
	for _, o := range []struct{ view, name string }{{"", "Show chat bubbles"}, {"split", "Show chat and terminal"}, {"terminal", "Show chat terminal"}} {
		add("chat-terminal", owner, "view-"+o.view, o.name, "Chat / "+o.name, func() {
			t.view = o.view
			g.layoutDirty = true
		})
	}
	if t.view == "" || t.pending != nil || a.project == "" && !g.ownsPlanningTerminal() {
		return
	}
	add("chat-terminal", owner, "sessions", "Terminal sessions", "Chat / Terminal sessions", func() {
		t.chooser = true
		g.requestChatTerminal(a, t, "list", "", "")
		g.layoutDirty = true
	})
	if t.chooser {
		if localHostTerminalAvailable(a, t) {
			add("chat-terminal", owner, "new-host-account", "New local host-account terminal", "Chat / Terminal sessions / New local host-account terminal", func() {
				id, err := domain.NewID(time.Now(), nil)
				if err == nil {
					g.requestChatTerminal(a, t, "start", id, "host-account")
				}
			})
		}
		for _, item := range t.choices {
			add("chat-terminal-session", owner, item.ID, item.Label, "Chat / Terminal sessions / "+item.Label, func() {
				t.chooser = false
				g.requestChatTerminal(a, t, "screen", item.ID, "select")
				g.layoutDirty = true
			})
		}
	}
}

func chatComposerButtons(parent *widget.Container) *widget.Container {
	row := widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})), widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Stretch([]bool{true, false, false}, []bool{false}), widget.GridLayoutOpts.Spacing(4, 0))))
	parent.AddChild(row)
	return row
}

func (g *Game) chatWorkspace(bubbles *widget.ScrollContainer, a modelAddress, t *chatTerminalPane) widget.PreferredSizeLocateableWidget {
	approval := planningColumn()
	g.approvalControls(approval, a, t)
	body := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}))))
	body.AddChild(approval)
	if t.view == "" {
		body.AddChild(bubbles)
		return body
	}
	head := planningColumn()
	g.chatTerminalContent(head, a, t)
	var term widget.PreferredSizeLocateableWidget = g.planningScroll(head)
	if t.state == "Running" && !t.hosts && !t.controls && !t.chooser && !t.readingHistory && (a.project != "" || g.ownsPlanningTerminal()) {
		live := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}))))
		if t.surface == nil || t.surface.id != t.selected {
			t.surface = g.newChatTerminalSurface(a, t)
		} else {
			reattachEditor(t.surface.TextInput)
		}
		live.AddChild(head, t.surface)
		term = live
	}
	if t.view == "terminal" {
		body.AddChild(term)
		return body
	}
	width := g.chatSplitAvailable(a)
	left := g.chatSplitWidth(width, t)
	bubbles.GetWidget().MinWidth = left
	split := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Stretch([]bool{false, false, true}, []bool{true}))))
	split.AddChild(&boundedPreferredWidget{child: bubbles, fixed: true, maxWidth: func() int {
		return left
	}}, g.chatDivider(), boundPreferredWidth(term, func() int {
		return max(44, width-left-44)
	}))
	body.AddChild(split)
	return body
}

func (g *Game) openChatPermissionModal(a modelAddress, t *chatTerminalPane) {
	t.permissionPage = true
	contents := widget.NewContainer(widget.ContainerOpts.BackgroundImage(uiimage.NewNineSliceColor(panelColor)), widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(bodyGap), widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)))))
	g.chatPermissionPage(contents, a, t)
	width := min(g.chatSplitAvailable(a), 620)
	x := g.width - windowMargin - width
	if g.sideRail() == "right_rail" {
		x -= railGutter
	}
	y := max(tabTop, (g.height-440)/2)
	modal := widget.NewWindow(widget.WindowOpts.Modal(), widget.WindowOpts.Contents(contents), widget.WindowOpts.Location(image.Rect(x, y, x+width, min(g.height-windowMargin, y+440))), widget.WindowOpts.CloseMode(widget.CLICK_OUT), widget.WindowOpts.ClosedHandler(func(*widget.WindowClosedEventArgs) {
		t.modal = nil
		t.permissionPage = false
		if t.permissionButton != nil {
			g.ui.SetFocusedWidget(t.permissionButton)
		}
	}))
	t.modal = modal
	g.ui.AddWindow(modal)
	focus := contents.GetFocusers()
	if len(focus) > 0 {
		g.ui.SetFocusedWidget(focus[0])
	}
}

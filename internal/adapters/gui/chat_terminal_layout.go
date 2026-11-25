//go:build gui

// Resizes the chat split pane using mouse or touch gesture ownership.
// A chat remembers its divider position independently of the outer project selector.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) closeChatPermissionPage() bool {
	_, t := g.currentChatTerminal()
	if t == nil || !t.permissionPage {
		return false
	}
	t.permissionPage = false
	g.layoutDirty = true
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
		t.permissionPage = true
		g.layoutDirty = true
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
		for _, item := range t.choices {
			add("chat-terminal-session", owner, item.ID, item.Label, "Chat / Terminal sessions / "+item.Label, func() {
				t.chooser = false
				g.requestChatTerminal(a, t, "screen", item.ID, "")
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

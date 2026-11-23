//go:build gui

// Presents transient discovery results in the Cluster detail pane.
// Linux uses Avahi on the selected interface; mobile supplies native callbacks.
package gui

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/adapters/discovery"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"net"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"time"
)

type lanView struct {
	device, message, selected string
	names                     []string
	cancel                    context.CancelFunc
	updates                   chan string
	next                      time.Time
	active                    bool
}

func (g *Game) discoveryContent(c *widget.Container) {
	if g.lan.selected != "" {
		c.AddChild(widget.NewText(widget.TextOpts.Text(wrapSelectorLabel(g.discoveryLabel(g.lan.selected), *g.pageTitleFace, max(1, g.detailPanePreferredWidth()-24)), g.pageTitleFace, g.theme.ButtonTheme.TextColor.Idle), widget.TextOpts.MaxWidth(float64(g.detailPanePreferredWidth()-24))))
		if g.runtimeInfo.ClaimHost != nil {
			host := g.lan.selected
			g.projectButton(c, "Claim this host for your cluster", func() {
				g.beginClaim(host)
			})
			c.AddChild(g.detailText("Connect using your saved Apparat identity."))
		}
		if g.runtimeInfo.PairPersonal != nil {
			host := g.lan.selected
			g.projectButton(c, "Receive my identity from this device", func() {
				g.pairFrom(host)
			})
		}
		g.projectButton(c, "Back to discovery", g.leavePeer)
		return
	}
	c.AddChild(g.detailText("Local network discovery"))
	if runtime.GOOS != "android" {
		devices, _ := net.Interfaces()
		for _, dev := range devices {
			if dev.Flags&net.FlagUp == 0 || dev.Flags&net.FlagMulticast == 0 {
				continue
			}
			g.projectButton(c, "Browse "+dev.Name, func() {
				g.startLAN(dev.Name)
			})
		}
	}
	if g.canOfferPair() {
		g.projectButton(c, "Pair another personal device", g.offerPair)
	}
	msg := g.lan.message
	if msg == "" {
		msg = "Select a network to browse."
	}
	c.AddChild(g.detailText(msg))
	for _, name := range g.lan.names {
		width := max(1, g.detailPanePreferredWidth()-tabIconSize-48)
		label := wrapSelectorLabel(g.discoveryLabel(name), *g.theme.ButtonTheme.TextFace, width)
		btn := widget.NewButton(widget.ButtonOpts.Text(label, g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor), widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter), widget.ButtonOpts.TextPadding(&widget.Insets{Left: 12 + tabIconSize + 8, Right: 12, Top: 8, Bottom: 8}),
			widget.ButtonOpts.Image(g.theme.ButtonTheme.Image), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44), widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})),
			widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
				if !g.bodySelectionSuppressed() {
					g.openPeer(name)
				}
			}))
		c.AddChild(plusRow(btn))
	}
}

func (g *Game) startLAN(device string) {
	g.stopLAN()
	ctx, cancel := context.WithCancel(g.projects.ctx)
	updates := make(chan string, 1)
	g.lan = lanView{device: device, cancel: cancel, updates: updates, active: true, message: "Searching on " + device}
	g.layoutDirty = true
	go func() {
		send := func(msg string) {
			select {
			case updates <- msg:
			case <-ctx.Done():
			}
		}
		if runtime.GOOS != "linux" {
			err := discovery.Browse(ctx, device, func(names []string) {
				send(lanNames(names))
			})
			if err != nil && ctx.Err() == nil {
				send("Discovery unavailable on " + device)
			}
			return
		}
		cmd := discovery.Command(ctx, "browse", device)
		pipe, err := cmd.StdoutPipe()
		if err == nil {
			err = cmd.Start()
		}
		if err == nil {
			scanner := bufio.NewScanner(pipe)
			for scanner.Scan() {
				var names []string
				if json.Unmarshal(scanner.Bytes(), &names) == nil {
					send(lanNames(names))
				}
			}
			_ = cmd.Wait()
		}
		if ctx.Err() == nil {
			send("Discovery unavailable on " + device)
		}
	}()
}

func lanNames(names []string) string {
	if len(names) == 0 {
		return "Searching · no nearby hosts"
	}
	encoded, _ := json.Marshal(names)
	return string(encoded)
}

var hostLabel = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

func (g *Game) receiveLAN(msg string) {
	var names []string
	if json.Unmarshal([]byte(msg), &names) == nil {
		valid := names[:0]
		for _, name := range names {
			if hostLabel.MatchString(name) {
				valid = append(valid, name)
			}
		}
		names = valid
		msg = "Searching · no nearby hosts"
		if len(names) > 0 {
			msg = "Nearby hosts"
		}
	}
	if msg == g.lan.message && slices.Equal(names, g.lan.names) {
		return
	}
	g.lan.message = msg
	g.lan.names = names
	if !slices.Contains(names, g.lan.selected) {
		g.lan.selected = ""
	}
	g.layoutDirty = true
}

func (g *Game) openPeer(name string) {
	if slices.Contains(g.lan.names, name) {
		g.lan.selected = name
		g.layoutDirty = true
	}
}

func (g *Game) leavePeer() {
	g.lan.selected = ""
	g.layoutDirty = true
}

func (g *Game) stopLAN() {
	if g.lan.cancel != nil {
		g.lan.cancel()
	}
	if g.lan.active && g.runtimeInfo.BrowseLAN != nil {
		g.runtimeInfo.BrowseLAN(false)
	}
	g.lan = lanView{}
}

func (g *Game) pollLAN() {
	tab := g.shell.Snapshot().ActiveTab()
	if tab.ID() == hud.TabCluster {
		tab = g.presentedTab(tab)
	}
	i := g.selectedSectionIndex(tab)
	if tab.ID() != hud.TabCluster || g.collapsed() && !g.detailOpen[hud.TabCluster] || i < 0 || i >= len(tab.Sections) || tab.Sections[i].ActionID != "add-new-device" {
		if g.lan.active {
			g.stopLAN()
		}
		return
	}
	if g.runtimeInfo.BrowseLAN != nil && time.Now().After(g.lan.next) {
		g.lan.active = true
		g.lan.next = time.Now().Add(time.Second)
		msg := g.runtimeInfo.BrowseLAN(true)
		var result struct {
			Network string
			Hosts   []string
		}
		if json.Unmarshal([]byte(msg), &result) == nil && result.Network != "" {
			g.lan.device = result.Network
			msg = lanNames(result.Hosts)
		}
		g.receiveLAN(msg)
	}
	select {
	case msg := <-g.lan.updates:
		g.receiveLAN(msg)
	default:
	}
}

func (g *Game) configureAPI() {
	if runtime.GOOS == "linux" {
		cmd := discovery.Command(g.projects.ctx, "configure", filepath.Join(g.runtimeInfo.RuntimePath, "data", "apparat.db"))
		go func() {
			_ = cmd.Run()
		}()
	}
}

func (g *Game) beginClaim(host string) {
	if g.runtimeInfo.ClaimHost != nil && slices.Contains(g.lan.names, host) {
		go g.runtimeInfo.ClaimHost(g.lan.device, host)
	}
}

func (g *Game) canOfferPair() bool {
	return runtime.GOOS == "linux" && g.runtimeInfo.PairPersonal != nil && g.lan.active && g.lan.device != ""
}
func (g *Game) offerPair() {
	if g.canOfferPair() {
		go g.runtimeInfo.PairPersonal(g.lan.device, "")
	}
}
func (g *Game) pairFrom(host string) {
	if g.runtimeInfo.PairPersonal != nil && slices.Contains(g.lan.names, host) {
		go g.runtimeInfo.PairPersonal(g.lan.device, host)
	}
}

// Enter the normal receiver workflow after the missing-user setup dialog closes.
func (g *Game) openRequestedDiscovery() {
	source, ok := g.runtimeInfo.Queries.(interface{ TakePairingDiscoveryRequest() bool })
	if !ok || !source.TakePairingDiscoveryRequest() {
		return
	}
	for index, tab := range g.shell.Snapshot().Tabs {
		if tab.ID() != hud.TabCluster {
			continue
		}
		g.selectTab(index, nil)
		for i, section := range g.presentedTab(tab).Sections {
			if section.ActionID == "add-new-device" {
				g.selectSection(tab.ID(), i)
				return
			}
		}
	}
}

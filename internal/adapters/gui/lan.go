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
	"runtime"
	"strings"
	"time"
)

type lanView struct {
	device, message string
	cancel          context.CancelFunc
	updates         chan string
	next            time.Time
	active          bool
}

func (g *Game) discoveryContent(c *widget.Container) {
	c.AddChild(g.detailText("Local network discovery"))
	if runtime.GOOS == "linux" {
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
	msg := g.lan.message
	if msg == "" {
		msg = "Select a network to browse."
	}
	c.AddChild(g.detailText(msg))
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
	return "Nearby hosts\n" + strings.Join(names, "\n")
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
	if tab.ID() != hud.TabCluster || i < 0 || i >= len(tab.Sections) || tab.Sections[i].ActionID != "add-new-device" {
		if g.lan.active {
			g.stopLAN()
		}
		return
	}
	if g.runtimeInfo.BrowseLAN != nil && time.Now().After(g.lan.next) {
		g.lan.active = true
		g.lan.next = time.Now().Add(time.Second)
		msg := g.runtimeInfo.BrowseLAN(true)
		if msg != g.lan.message {
			g.lan.message = msg
			g.layoutDirty = true
		}
	}
	select {
	case msg := <-g.lan.updates:
		g.lan.message = msg
		g.layoutDirty = true
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

//go:build gui

// Routes terminal lifecycle, controller handoff and bounded input via workspace operations.
// Session output is separate from telemetry and stays on its owning node.
package gui

import (
	"bytes"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/ebitenui/ebitenui/widget"
	et "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gomono"
	"image/color"
	"time"
)

var terminalFont, _ = et.NewGoTextFaceSource(bytes.NewReader(gomono.TTF))

func (g *Game) terminalScreen(c *widget.Container, w *workspacePanel) {
	if w.report != "" {
		if w.mode == "terminal-session" {
			var face et.Face = &et.GoTextFace{Source: terminalFont, Size: max(8, min(15, g.detailTextMaxWidth()/48))}
			c.AddChild(widget.NewText(widget.TextOpts.Text(w.report, &face, color.White)))
		} else {
			c.AddChild(g.detailText(w.report))
		}
	}
	if w.mode == "terminal" {
		for _, choice := range w.choices {
			g.workspaceRow(c, choice.Label, func() {
				w.runID = choice.ID
				g.workspaceGo("terminal-session", "terminal-screen")
			})
		}
		if w.cursor != "" {
			g.projectButton(c, "More sessions", func() {
				w.argument = w.cursor
				g.workspaceRequest("terminal-list")
			})
		}
		if w.report == "" {
			g.projectButton(c, "Start isolated terminal", func() {
				w.runID, _ = domain.NewID(time.Now(), nil)
				g.workspaceRequest("terminal-start")
			})
		}
	} else if w.mode == "terminal-history" {
		if w.cursor != "" {
			g.projectButton(c, "Older output", func() {
				w.argument = w.cursor
				g.workspaceRequest("terminal-history")
			})
		}
	} else {
		if w.selected.State != "Ended" {
			g.multilineProjectField(c, "Terminal input", &w.message)
			keys := []struct{ label, input string }{{"Send line", ""}, {"Enter", "\r"}, {"Tab", "\t"}, {"Escape", "\x1b"}, {"Interrupt", "\x03"}, {"Up", "\x1b[A"}, {"Down", "\x1b[B"}, {"Left", "\x1b[D"}, {"Right", "\x1b[C"}}
			for _, key := range keys {
				g.projectButton(c, key.label, func() {
					w.argument = key.input
					if w.argument == "" {
						w.argument = w.message + "\r"
					}
					g.workspaceRequest("terminal-input")
				})
			}
			g.projectButton(c, "Send text without Enter", func() {
				w.argument = w.message
				g.workspaceRequest("terminal-input")
			})
			for _, size := range []string{"40x20", "80x24", "120x40"} {
				g.projectButton(c, size, func() {
					w.argument = size
					g.workspaceRequest("terminal-resize")
				})
			}
			g.workspaceActions(c, "Take control", "terminal-take", "End session", "terminal-stop")
		}
		g.projectButton(c, "Retained output", func() {
			g.workspaceGo("terminal-history", "terminal-history")
		})
		g.projectButton(c, "Refresh screen", func() {
			g.workspaceRequest("terminal-screen")
		})
	}
}

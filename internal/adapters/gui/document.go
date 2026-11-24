//go:build gui

// Opens a platform multiline editor over the same project-scoped draft and save commands.
// Only the selected document crosses the bridge. The editor never receives filesystem access.
package gui

import (
	"context"
	"github.com/ebitenui/ebitenui/widget"
	"strings"
)

func (g *Game) projectField(c *widget.Container, label string, value *string) widget.Focuser {
	if g.runtimeInfo.PromptValue == nil {
		return g.projectInput(c, label, *value, func(text string) {
			*value = text
		})
	}
	return g.nativeProjectField(c, label, *value, true, func(text string) {
		*value = text
	})
}
func (g *Game) nativeProjectField(c *widget.Container, label string, value string, singleLine bool, changed func(string)) *widget.Button {
	prompt := g.runtimeInfo.PromptValue
	var row *widget.Container
	if singleLine {
		row = g.formFieldRow(c, label)
	} else {
		row = g.multilineRow(c, label)
	}
	caption := "Tap to enter…"
	if value != "" {
		caption = value
	}
	width := max(1, g.detailPanePreferredWidth()-56-g.formLabelWidth()-4-32)
	if !singleLine {
		width += g.formLabelWidth() + 4
	}
	caption = wrapSelectorLabel(caption, *g.theme.ButtonTheme.TextFace, width)
	if singleLine {
		first, _, more := strings.Cut(caption, "\n")
		if more {
			caption = first + "…"
		}
	}
	button := g.projectButton(row, caption, func() {
		g.projectWork(func(ctx context.Context) projectResult {
			text, ok := prompt(label, value, singleLine)
			return projectResult{applyField: func() {
				if ok {
					changed(text)
				}
			}}
		})
	})
	if !singleLine {
		button.GetWidget().LayoutData = widget.RowLayoutData{Stretch: true}
		button.GetWidget().MinHeight = 3*(g.prefs.Font+5) + 8
	}
	return button
}

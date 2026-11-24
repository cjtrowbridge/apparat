//go:build gui

// Stacks the form caption and a full-width multiline editor.
// Reuses document keyboard, IME, clipboard, selection and undo handling.
package gui

import "github.com/ebitenui/ebitenui/widget"

func (g *Game) multilineRow(c *widget.Container, label string) *widget.Container {
	row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		widget.RowLayoutOpts.Spacing(4),
		widget.RowLayoutOpts.Padding(&widget.Insets{Top: g.prefs.Font / 2, Bottom: g.prefs.Font / 2}),
	)), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})))
	c.AddChild(row)
	if label == "" {
		g.addModelButton(row)
	} else {
		row.AddChild(g.formLabel(label))
	}
	return row
}

func (g *Game) multilineInput(c *widget.Container, label string, value *string) *documentSurface {
	row := g.multilineRow(c, label)
	if g.ui != nil {
		field, ok := g.ui.GetFocusedWidget().(*widget.TextInput)
		if ok {
			old, _ := field.GetWidget().CustomData.(*documentSurface)
			if old != nil && old.draft == value {
				reattachEditor(field)
				row.AddChild(old)
				return old
			}
		}
	}
	e := &documentSurface{TextInput: g.newProjectInput("", ignoreDocumentTextChange), game: g, draft: value, height: max(20, g.prefs.Font+5)}
	e.GetWidget().CustomData = e
	e.GetWidget().MinHeight = 3*e.height + 8
	e.load(*value)
	e.anchor = len(*value)
	e.caret = e.anchor
	e.setupIME()
	row.AddChild(e)
	return e
}

func (g *Game) multilineProjectField(c *widget.Container, label string, value *string) {
	if g.runtimeInfo.PromptValue != nil {
		g.nativeProjectField(c, label, *value, false, func(s string) {
			*value = s
		})
		return
	}
	g.multilineInput(c, label, value)
}

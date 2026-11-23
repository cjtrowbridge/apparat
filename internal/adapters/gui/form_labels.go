//go:build gui

// Renders project form labels and descriptions with distinct font styles.
// Field rows reserve a bounded label column and give remaining width to the editor.
package gui

import "github.com/ebitenui/ebitenui/widget"

func (g *Game) formLabel(label string) *widget.Text {
	return widget.NewText(widget.TextOpts.Text(label, g.formLabelFace, g.theme.LabelTheme.Color.Idle), widget.TextOpts.MaxWidth(g.detailTextMaxWidth()))
}
func (g *Game) formDescription(message string) *widget.Text {
	return widget.NewText(widget.TextOpts.Text(message, g.formDescriptionFace, g.theme.LabelTheme.Color.Idle), widget.TextOpts.MaxWidth(g.detailTextMaxWidth()))
}
func (g *Game) formLabelWidth() int {
	return min(155, max(24, (g.detailPanePreferredWidth()-56)/3))
}
func (g *Game) formFieldRow(c *widget.Container, label string) *widget.Container {
	width := g.formLabelWidth()
	row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{false}), widget.GridLayoutOpts.Spacing(8, 0))), widget.ContainerOpts.AutoDisableChildren(), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})))
	caption := widget.NewText(widget.TextOpts.Text(wrapSelectorLabel(label, *g.formLabelFace, width), g.formLabelFace, g.theme.LabelTheme.Color.Idle), widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter), widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(width, 0)))
	row.AddChild(caption)
	c.AddChild(row)
	return row
}

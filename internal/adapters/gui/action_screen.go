//go:build gui

// Presents the searchable action directory and the voice clarification dialog.
// Both use the same current catalog and the shell's keyboard and controller navigation.
package gui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

type actionScreen struct {
	mode, query, prompt string
	choices             []hud.Destination
	returnToBrowser     bool
	focusIndex          int
	returnStrip         bool
	pane                *widget.ScrollContainer
	stopSpeech          context.CancelFunc
	background          *ebitenui.UI
}

func (g *Game) stopChoiceSpeech() {
	if g.runtimeInfo.SpeakChoice != nil {
		g.runtimeInfo.SpeakChoice("")
	}
	if g.actions.stopSpeech != nil {
		g.actions.stopSpeech()
		g.actions.stopSpeech = nil
	}
}
func (g *Game) speakChoices(text string) {
	g.stopChoiceSpeech()
	if !g.prefs.SpokenChoices {
		return
	}
	if g.runtimeInfo.SpeakChoice != nil {
		g.runtimeInfo.SpeakChoice(text)
		return
	}
	for _, name := range []string{"espeak-ng", "espeak"} {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(g.projects.ctx, 45*time.Second)
		g.actions.stopSpeech = cancel
		go func() {
			defer cancel()
			cmd := exec.CommandContext(ctx, path, "--stdin")
			cmd.Stdin = strings.NewReader(text)
			_ = cmd.Run()
		}()
		return
	}
}
func (g *Game) openActionBrowser() {
	for i, t := range g.shell.Snapshot().Tabs {
		if t.ID() == hud.TabSettings {
			g.selectTab(i, nil)
			break
		}
	}
	g.actions = actionScreen{mode: "browse", returnStrip: g.nav.strip}
	g.resetActionFocus()
}
func (g *Game) resetActionFocus() {
	g.ui.ClearFocus()
	g.nav.target = focusTarget{}
	g.nav.strip = false
	g.nav.gallery = false
	g.nav.active = true
	g.nav.resizing = false
	g.rebuildUI(g.shell.Snapshot())
}
func (g *Game) closeActionScreen() {
	g.stopChoiceSpeech()
	strip := g.actions.returnStrip
	g.actions = actionScreen{}
	g.nav.target = focusTarget{}
	g.nav.strip = strip
	g.layoutDirty = true
}
func (g *Game) backActionScreen() {
	browse := g.actions.returnToBrowser
	query := g.actions.query
	g.closeActionScreen()
	if browse {
		g.actions = actionScreen{mode: "browse", query: query, returnStrip: g.nav.strip}
		g.resetActionFocus()
		return
	}
	g.rebuildUI(g.shell.Snapshot())
}
func (g *Game) actionNotice(prompt string, choices []hud.Destination) {
	a := &g.actions
	if a.mode == "" {
		a.returnStrip = g.nav.strip
	}
	if a.mode != "choose" {
		a.background = g.ui
	}
	a.focusIndex = 0
	a.returnToBrowser = a.returnToBrowser || a.mode == "browse"
	a.mode = "choose"
	a.prompt = prompt
	a.choices = choices
	a.pane = nil
	g.resetActionFocus()
	spoken := prompt
	var lines []string
	for i, c := range choices {
		lines = append(lines, fmt.Sprintf("%d. %s, %s.", i+1, c.Address.Kind, c.Path))
	}
	spoken += " " + strings.Join(lines, " ")
	g.speakChoices(spoken)
}
func (g *Game) voiceCommand(text string) {
	words := hud.CommandWords(text)
	if words == "cancel" && g.actions.mode != "" {
		g.backActionScreen()
		return
	}
	if g.actions.mode == "choose" {
		words = strings.TrimPrefix(words, "the ")
		ordinal := []string{"first", "second", "third", "fourth", "fifth", "sixth", "seventh", "eighth", "ninth", "tenth"}
		numbers := strings.Fields("one two three four five six seven eight nine ten")
		var matches []hud.Destination
		for i, c := range g.actions.choices {
			if words == strconv.Itoa(i+1) || i < len(ordinal) && (words == ordinal[i] || words == ordinal[i]+" one" || words == numbers[i]) {
				g.invokeAction(c.Address)
				return
			}
			if words == c.Address.Kind || words == hud.CommandWords(c.Name) || words == hud.CommandWords(c.Path) {
				matches = append(matches, c)
			}
		}
		if len(matches) == 1 {
			g.invokeAction(matches[0].Address)
			return
		}
		if len(g.actions.choices) > 0 {
			g.actionNotice("Choose a listed destination by number or say cancel.", g.actions.choices)
			return
		}
		g.backActionScreen()
	}
	if words == "what can i say" || words == "help" {
		g.openActionBrowser()
		return
	}
	if g.awaitProjectCatalog() {
		g.actionNotice("Project destinations are loading. Try the command again or say cancel.", nil)
		return
	}
	m := g.queryActions(text, true)
	switch {
	case m.Exact && len(m.Choices) == 1:
		g.invokeAction(m.Choices[0].Address)
	case m.Exact:
		g.actionNotice("Multiple destinations match. Which would you like to open?", m.Choices)
	case len(m.Choices) > 0:
		g.actionNotice("Did you mean one of these destinations?", m.Choices)
	default:
		g.actionNotice("No destination matched “"+text+"”. Say another command or cancel.", nil)
	}
}
func (g *Game) buildActionScreen() {
	a := &g.actions
	g.tabScroll = nil
	g.divider = nil
	c := g.selectorRows()
	if a.mode == "browse" {
		c.AddChild(g.hudText("Command directory"))
		g.projectButton(c, "Back to Settings", g.backActionScreen)
		c.AddChild(g.hudText("Try “open Projects” or “switch to Settings”."))
		field := g.editorForLayout("Search actions", func() *widget.TextInput {
			input := g.newProjectInput(a.query, func(value string) {
				if a.mode != "browse" || value == a.query {
					return
				}
				a.query = value
				a.pane = nil
				g.layoutDirty = true
			})
			widget.TextInputOpts.Image(&widget.TextInputImage{Idle: uiimage.NewNineSliceColor(panelColor), Highlight: uiimage.NewNineSliceColor(activeTabColor)})(input)
			widget.TextInputOpts.Placeholder("Search actions")(input)
			return input
		})
		c.AddChild(field)
		if g.projects.store != nil && g.projects.catalogInitialized && !g.projects.catalogLoaded {
			if g.projects.pending != nil {
				c.AddChild(g.hudText("Loading project destinations…"))
			} else {
				g.projectButton(c, "Retry loading project destinations", func() {
					g.awaitProjectCatalog()
				})
			}
		}
		result := g.queryActions(a.query, false)
		if len(result.Choices) == 0 {
			c.AddChild(g.hudText("No matching actions"))
		}
		for _, v := range result.Choices {
			g.actionRow(c, v.Path, v.Address)
		}
	} else {
		c.AddChild(widget.NewText(widget.TextOpts.Text(a.prompt, g.theme.ButtonTheme.TextFace, g.theme.LabelTheme.Color.Idle), widget.TextOpts.MaxWidth(float64(g.choiceBounds().Dx()-32))))
		for i, v := range a.choices {
			g.actionRow(c, fmt.Sprintf("%d. %s · %s", i+1, v.Address.Kind, v.Path), v.Address)
		}
		g.projectButton(c, "Cancel", g.backActionScreen)
	}
	scroll := widget.NewScrollContainer(widget.ScrollContainerOpts.Content(c), widget.ScrollContainerOpts.Image(createScrollContainerImage()), widget.ScrollContainerOpts.StretchContentWidth(), widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true})))
	if a.pane != nil {
		scroll.ScrollTop = a.pane.ScrollTop
	}
	a.pane = scroll
	g.verticalScrolls = []*widget.ScrollContainer{scroll}
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 80}))))
	if a.mode == "choose" {
		root = g.choiceRoot(scroll)
	} else {
		root.AddChild(boundPreferredWidth(scroll, g.hudPreferredWidth))
	}
	old := g.ui.GetFocusedWidget()
	g.ui = &ebitenui.UI{DisableDefaultFocus: true, Container: root, PrimaryTheme: g.theme}
	g.nav.body = scroll
	g.nav.target = focusTarget{}
	g.nav.strip = false
	g.restoreEditing(old)
}

func (g *Game) settingsCardControls(c *widget.Container, id string) {
	switch id {
	case "hud-configuration", "navigation-layout":
		g.preferenceControls(c, id)
	case "speech-recognition":
		g.speechSettings(c)
	case "api-configuration":
		if runtime.GOOS == "linux" {
			g.projectButton(c, "Configure API hosting", g.configureAPI)
		} else {
			c.AddChild(g.detailText("Client only · scoped hosting unavailable on this platform"))
		}
	case "command-directory":
		g.projectButton(c, "Browse available commands", g.openActionBrowser)
		g.spokenChoiceSetting(c)
	}
}

func (g *Game) spokenChoiceSetting(c *widget.Container) {
	for _, name := range []string{"espeak-ng", "espeak"} {
		_, err := exec.LookPath(name)
		if err == nil || g.runtimeInfo.SpeakChoice != nil {
			label := "Enable spoken action choices"
			if g.prefs.SpokenChoices {
				label = "Disable spoken action choices"
			}
			g.projectButton(c, label, func() {
				p := g.prefs.Preferences
				p.SpokenChoices = !p.SpokenChoices
				g.savePreferences(p)
			})
			break
		}
	}
}

func (g *Game) actionRow(c *widget.Container, label string, id hud.ActionAddress) {
	width := g.hudPreferredWidth()
	if g.actions.mode == "choose" {
		width = g.choiceBounds().Dx()
	}
	g.projectButton(c, wrapSelectorLabel(label, *g.theme.ButtonTheme.TextFace, width-60), func() {
		g.invokeAction(id)
	})
}

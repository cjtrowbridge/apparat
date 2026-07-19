package hud

import "fmt"

type Action string

const (
	ActionPreviousTab     Action = "previous_tab"
	ActionNextTab         Action = "next_tab"
	ActionDirectTab       Action = "direct_tab"
	ActionPushToTalk      Action = "push_to_talk"
	ActionCancelRecording Action = "cancel_recording"
	ActionFocusUp         Action = "focus_up"
	ActionFocusDown       Action = "focus_down"
	ActionFocusLeft       Action = "focus_left"
	ActionFocusRight      Action = "focus_right"
	ActionActivate        Action = "activate"
	ActionBack            Action = "back"
	ActionContextMenu     Action = "context_menu"
	ActionCommandPalette  Action = "command_palette"
	ActionScroll          Action = "scroll"
	ActionScrollUp        Action = "scroll_up"
	ActionScrollDown      Action = "scroll_down"
	ActionCollectionNext  Action = "collection_next"
	ActionCollectionPrev  Action = "collection_previous"
)

func DefaultBindings() BindingConfig {
	return BindingConfig{Bindings: map[Action][]Binding{
		ActionPreviousTab:     {{Device: "controller", Input: "L1"}, {Device: "keyboard", Input: "Ctrl+PageUp"}},
		ActionNextTab:         {{Device: "controller", Input: "R1"}, {Device: "keyboard", Input: "Ctrl+PageDown"}},
		ActionDirectTab:       {{Device: "keyboard", Input: "Alt+1..Alt+6"}},
		ActionPushToTalk:      {{Device: "controller", Input: "R2"}, {Device: "keyboard", Input: "RightCtrl"}},
		ActionCancelRecording: {{Device: "keyboard", Input: "Escape"}},
		ActionFocusUp:         {{Device: "controller", Input: "DPadUp"}, {Device: "keyboard", Input: "ArrowUp"}},
		ActionFocusDown:       {{Device: "controller", Input: "DPadDown"}, {Device: "keyboard", Input: "ArrowDown"}},
		ActionFocusLeft:       {{Device: "controller", Input: "DPadLeft"}, {Device: "keyboard", Input: "ArrowLeft"}},
		ActionFocusRight:      {{Device: "controller", Input: "DPadRight"}, {Device: "keyboard", Input: "ArrowRight"}},
		ActionActivate:        {{Device: "controller", Input: "A"}, {Device: "keyboard", Input: "Enter"}, {Device: "keyboard", Input: "Space"}},
		ActionBack:            {{Device: "controller", Input: "B"}, {Device: "keyboard", Input: "Escape"}},
		ActionContextMenu:     {{Device: "controller", Input: "Menu"}, {Device: "keyboard", Input: "Shift+F10"}},
		ActionCommandPalette:  {{Device: "keyboard", Input: "Ctrl+Shift+P"}},
		ActionScroll:          {{Device: "controller", Input: "RightStick"}, {Device: "pointer", Input: "Wheel"}, {Device: "pointer", Input: "Drag"}, {Device: "touch", Input: "Drag"}},
		ActionScrollUp:        {{Device: "keyboard", Input: "PageUp"}, {Device: "controller", Input: "RightStickUp"}},
		ActionScrollDown:      {{Device: "keyboard", Input: "PageDown"}, {Device: "controller", Input: "RightStickDown"}},
		ActionCollectionNext:  {{Device: "keyboard", Input: "PageDown"}},
		ActionCollectionPrev:  {{Device: "keyboard", Input: "PageUp"}},
	}}
}

func (bindings BindingConfig) Inputs(action Action) []Binding {
	return append([]Binding(nil), bindings.Bindings[action]...)
}

func (shell *Shell) ApplyAction(action Action) error {
	switch action {
	case ActionPreviousTab:
		shell.PreviousTab()
	case ActionNextTab:
		shell.NextTab()
	case ActionFocusUp, ActionFocusLeft, ActionCollectionPrev, ActionScrollUp:
		shell.MoveFocus(-1)
	case ActionFocusDown, ActionFocusRight, ActionCollectionNext, ActionScrollDown:
		shell.MoveFocus(1)
	case ActionCancelRecording:
		shell.CancelVoiceCapture()
	case ActionPushToTalk:
		shell.StartVoiceCapture(string(action))
	default:
		return fmt.Errorf("unsupported action %q", action)
	}
	return nil
}

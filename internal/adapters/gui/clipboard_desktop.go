//go:build gui && !android

// Exchanges UTF-8 text through the native system clipboard.
// Wayland sessions prefer the compositor clipboard and fall back to XWayland.
package gui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.design/x/clipboard"
)

func systemCopy(value string) bool {
	if runtime.GOOS == "linux" && os.Getenv("WAYLAND_DISPLAY") != "" {
		cmd := exec.Command("wl-copy", "--type", "text/plain;charset=utf-8")
		cmd.Stdin = strings.NewReader(value)
		if cmd.Run() == nil {
			return true
		}
	}
	return clipboard.Init() == nil && clipboard.Write(clipboard.FmtText, []byte(value)) != nil
}

func systemPaste() string {
	if runtime.GOOS == "linux" && os.Getenv("WAYLAND_DISPLAY") != "" {
		b, err := exec.Command("wl-paste", "--no-newline", "--type", "text").Output()
		if err == nil {
			return string(b)
		}
	}
	if clipboard.Init() != nil {
		return ""
	}
	return string(clipboard.Read(clipboard.FmtText))
}

//go:build gui && linux && !android

// Reads the session clipboard without development libraries on the launcher's library path.
// GTK negotiates text targets and obtains foreground access when Wayland requires it.
package gui

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.design/x/clipboard"
)

//go:embed clipboard_linux.py
var gtkClipboard string

var errMissingGTK = errors.New("Clipboard unavailable. Install GTK 3 and Python GObject bindings for the desktop clipboard helper.")

var clipboardOwner struct {
	sync.Mutex
	process *os.Process
}

func desktopClipboardError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return errMissingGTK
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		switch exit.ExitCode() {
		case 3:
			return errors.New("Nothing to paste. The clipboard contains an item without a text format.")
		case 4:
			return errors.New("Clipboard unavailable. Apparat could not connect to the graphical session. Open Apparat from your desktop.")
		case 5:
			return errors.New("Clipboard access requires focus. Select the Apparat clipboard window and retry.")
		case 6:
			return errors.New("Paste failed. The source app advertised text but did not deliver it. Copy the text again.")
		case 7:
			return errMissingGTK
		}
	}
	return errors.New("The desktop clipboard helper could not complete the transfer. Check that /usr/bin/python3 and GTK 3 are available.")
}

func gtkCommand(ctx context.Context, action, backend string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "/usr/bin/python3", "-I", "-c", gtkClipboard, action)
	cmd.Env = append(os.Environ(), "GDK_BACKEND="+backend)
	return cmd
}

func systemPaste() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	wayland := os.Getenv("WAYLAND_DISPLAY") != ""
	if wayland {
		probe, stop := context.WithTimeout(ctx, 1500*time.Millisecond)
		bytes, err := exec.CommandContext(probe, "wl-paste", "--no-newline", "--type", "text").Output()
		stop()
		if err == nil {
			return string(bytes), nil
		}
	}
	if os.Getenv("DISPLAY") != "" {
		probe := ctx
		if wayland {
			var stop context.CancelFunc
			probe, stop = context.WithTimeout(ctx, 2*time.Second)
			defer stop()
		}
		data, err := gtkCommand(probe, "paste", "x11").Output()
		if err != nil && errors.Is(desktopClipboardError(err), errMissingGTK) && clipboard.Init() == nil {
			native := clipboard.Read(clipboard.FmtText)
			if native != nil {
				return string(native), nil
			}
		}
		if !wayland || err == nil && len(data) > 0 {
			return string(data), clipboardReadError(ctx, err)
		}
	}
	if !wayland {
		return "", errors.New("Paste unavailable. Apparat has no connected graphical display. Open it from your desktop.")
	}
	data, err := gtkCommand(ctx, "paste", "wayland").Output()
	return string(data), clipboardReadError(ctx, err)
}

func clipboardReadError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return errors.New("Paste timed out waiting for the clipboard owner. Return to the source app and copy the text again.")
	}
	if err != nil {
		return desktopClipboardError(err)
	}
	return nil
}

func systemCopy(value string) error {
	wayland := os.Getenv("WAYLAND_DISPLAY") != ""
	if wayland {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "wl-copy", "--type", "text/plain;charset=utf-8")
		cmd.Stdin = strings.NewReader(value)
		if cmd.Run() == nil {
			return nil
		}
	}
	if os.Getenv("DISPLAY") != "" {
		err := gtkCopy(value, "x11")
		if errors.Is(err, errMissingGTK) && clipboard.Init() == nil && clipboard.Write(clipboard.FmtText, []byte(value)) != nil {
			return nil
		}
		if err == nil || !wayland {
			return err
		}
	}
	if wayland {
		return gtkCopy(value, "wayland")
	}
	return errors.New("Copy unavailable. Apparat has no connected graphical display. Open it from your desktop.")
}

func gtkCopy(value, backend string) error {
	cmd := gtkCommand(context.Background(), "copy", backend)
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	cmd.Stdin = strings.NewReader(value)
	output, err := cmd.StdoutPipe()
	if err != nil {
		return desktopClipboardError(err)
	}
	err = cmd.Start()
	if err != nil {
		return desktopClipboardError(err)
	}
	timer := time.AfterFunc(6*time.Second, func() {
		_ = cmd.Process.Kill()
	})
	var ready [1]byte
	_, err = io.ReadFull(output, ready[:])
	stopped := timer.Stop()
	if err != nil || !stopped || ready[0] != '1' {
		wait := cmd.Wait()
		return desktopClipboardError(wait)
	}
	clipboardOwner.Lock()
	previous := clipboardOwner.process
	clipboardOwner.process = cmd.Process
	clipboardOwner.Unlock()
	if previous != nil {
		_ = previous.Kill()
	}
	go func() {
		_ = cmd.Wait()
		clipboardOwner.Lock()
		if clipboardOwner.process == cmd.Process {
			clipboardOwner.process = nil
		}
		clipboardOwner.Unlock()
	}()
	return nil
}

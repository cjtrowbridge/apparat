//go:build gui && !android && !linux

// Exchanges text with the system clipboard on Windows and macOS.
// A busy clipboard gets a short retry before an error reaches the editor.
package gui

import (
	"errors"
	"time"

	"golang.design/x/clipboard"
)

func systemCopy(value string) error {
	err := clipboard.Init()
	if err != nil {
		return errors.New("Copy failed. Apparat could not connect to the desktop clipboard.")
	}
	for i := 0; i < 3; i++ {
		if clipboard.Write(clipboard.FmtText, []byte(value)) != nil {
			return nil
		}
		time.Sleep(40 * time.Millisecond)
	}
	return errors.New("Copy failed. The desktop refused the clipboard write after three attempts.")
}

func systemPaste() (string, error) {
	err := clipboard.Init()
	if err != nil {
		return "", errors.New("Paste failed. Apparat could not connect to the desktop clipboard.")
	}
	for i := 0; i < 3; i++ {
		b := clipboard.Read(clipboard.FmtText)
		if b != nil {
			return string(b), nil
		}
		time.Sleep(40 * time.Millisecond)
	}
	return "", errors.New("Nothing to paste. The desktop supplied no text after three attempts. Copy plain text from the source app.")
}

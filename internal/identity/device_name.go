// Loads the installation's chosen hostname or derives an unambiguous OS name.
// Both pairing and login require the same durable choice on generically named devices.
package identity

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (s *Session) Hostname() (string, error) {
	host, err := readPrivate(filepath.Join(s.dir, "hostname"))
	if os.IsNotExist(err) {
		system, _ := os.Hostname()
		name := domain.SystemDeviceName(system)
		if name != "" {
			return name, nil
		}
	} else if err != nil {
		return "", err
	} else if domain.ValidDeviceName(string(host)) {
		return string(host), nil
	}
	return "", errors.New("choose a hostname for this installation during device setup")
}

func (s *Session) SetupHostname(dialog Dialog) bool {
	_, err := s.Hostname()
	if err == nil {
		return true
	}
	ui := setupUI{dialog}
	message := "Choose a recognizable hostname for this device, such as pixel-4a. Other devices in your cluster will see this name."
	for {
		name, ok := ui.ask("text", "Name this device", message, "")
		if !ok {
			return false
		}
		name = strings.TrimSpace(name)
		if !domain.ValidDeviceName(name) {
			message = "Use 1–63 letters, digits or hyphens. Choose a specific name, such as pixel-4a."
			continue
		}
		err = atomicWrite(filepath.Join(s.dir, "hostname"), []byte(name), 0600)
		if err != nil {
			ui.ask("notice", "Device name", "Could not save the hostname.", "")
		}
		return err == nil
	}
}

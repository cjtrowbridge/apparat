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
	name, ok := ui.ask("text", "Name this device", "Choose a recognizable hostname, such as Pixel 10 XL. Spaces become hyphens. You can also name this device later in Cluster.", "")
	if !ok {
		return false
	}
	err = s.SetHostname(name)
	if err != nil {
		ui.ask("notice", "Device name", err.Error(), "")
	}
	return err == nil
}

func (s *Session) SetHostname(value string) error {
	name := strings.Join(strings.Fields(value), "-")
	if !domain.ValidDeviceName(name) {
		return errors.New("choose a specific name with 1–63 letters, digits or hyphens, such as Pixel-10-XL")
	}
	err := atomicWrite(filepath.Join(s.dir, "hostname"), []byte(name), 0600)
	if err != nil {
		return errors.New("could not save this device’s hostname")
	}
	return nil
}

// Prompts interactive installations to create, restore or unlock a personal identity.
// Backups contain the signed user profile and password-encrypted private key.
package identity

import (
	"encoding/base64"
	"os"
	"path/filepath"
)

func (p *Personal) Setup(dialog Dialog) bool {
	p.mu.Lock()
	ready := len(p.key) > 0
	p.mu.Unlock()
	if ready {
		return false
	}
	ui := setupUI{dialog}
	_, err := os.Lstat(p.path)
	if !os.IsNotExist(err) {
		ref, readErr := readPrivate(filepath.Join(filepath.Dir(p.path), "unlock.ref"))
		if readErr == nil {
			pass, loaded := ui.ask("load", "Unlock personal identity", "", string(ref))
			if loaded && p.Unlock([]byte(pass)) == nil {
				return false
			}
		}
		password, ok := ui.ask("password", "Unlock personal identity", "Enter your personal identity password.", "")
		if ok && p.Unlock([]byte(password)) != nil {
			ui.ask("notice", "Personal identity locked", "Check your password and recovery copy.", "")
		}
		return false
	}
	const first = "This is my first Apparat device"
	const join = "Join my existing cluster"
	action, ok := ui.ask("choose", "Your Apparat identity", "Is this your first Apparat device, or are you joining your existing cluster?", "", first, join)
	if !ok {
		return false
	}
	var password string
	switch action {
	case join:
		method, selected := ui.ask("choose", "Join your existing cluster", "How would you like to bring your personal identity to this device?", "", "Pair with another device", "Restore a backup")
		if !selected {
			return false
		}
		if method == "Pair with another device" {
			return true
		}
		if method != "Restore a backup" {
			return false
		}
		raw, selected := ui.ask("file", "Restore personal identity", "Select your encrypted personal identity backup.", "")
		if !selected {
			return false
		}
		data, decodeErr := base64.StdEncoding.DecodeString(raw)
		var entered bool
		password, entered = ui.ask("password", "Restore personal identity", "Enter the password for the backup.", "")
		if !entered {
			return false
		}
		err = decodeErr
		if err == nil {
			err = p.Import(data, []byte(password))
		}
	case first:
		name, entered := ui.ask("text", "Your screen name", "Enter the name to show to other cluster users.", "")
		if !entered {
			return false
		}
		var confirmed bool
		password, confirmed = ui.password("Protect your portable personal identity")
		if !confirmed {
			return false
		}
		err = p.Create(name, []byte(password))
	default:
		return false
	}
	if err != nil {
		ui.ask("notice", "Personal identity setup", err.Error(), "")
		return false
	}
	p.offerSystemUnlock(ui, password)
	choice, ok := ui.ask("choose", "Personal identity ready", "Keep an encrypted recovery copy so you can restore your user on another device.", "", "Save recovery copy", "Finish")
	if ok && choice == "Save recovery copy" {
		data, err := readPrivate(p.path)
		if err == nil {
			_, ok = ui.ask("save", "Save personal identity", "", base64.StdEncoding.EncodeToString(data))
		}
		if err != nil || !ok {
			ui.ask("notice", "Recovery copy", "Your local identity is ready. The recovery copy was not saved.", "")
		}
	}
	return false
}

func NotifyIdentityUnavailable(dialog Dialog) {
	ui := setupUI{dialog}
	ui.ask("notice", "Personal identity locked", "Apparat could not record your device authorization. Retry after checking local storage.", "")
}

func (p *Personal) offerSystemUnlock(ui setupUI, password string) {
	useSystem, ok := ui.systemUnlockChoice("Unlock your personal identity")
	if !ok || !useSystem {
		return
	}
	ref, saved := ui.ask("store", "Protect personal identity", "", password)
	if saved && ref != "" {
		err := atomicWrite(filepath.Join(filepath.Dir(p.path), "unlock.ref"), []byte(ref), 0600)
		if err == nil {
			return
		}
		ui.ask("remove", "", "", ref)
	}
	ui.ask("notice", "System unlock unavailable", "Use your personal identity password to unlock.", "")
}

func (p *Personal) OfferPairedUnlock(dialog Dialog, password string) {
	_, err := os.Lstat(filepath.Join(filepath.Dir(p.path), "unlock.ref"))
	if os.IsNotExist(err) {
		p.offerSystemUnlock(setupUI{dialog}, password)
	}
}

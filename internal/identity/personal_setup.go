// Prompts interactive installations to create, restore or unlock a personal identity.
// Backups contain the signed user profile and password-encrypted private key.
package identity

import (
	"encoding/base64"
	"os"
	"path/filepath"
)

func (p *Personal) Setup(dialog Dialog) {
	p.mu.Lock()
	ready := len(p.key) > 0
	p.mu.Unlock()
	if ready {
		return
	}
	ui := setupUI{dialog}
	_, err := os.Lstat(p.path)
	if !os.IsNotExist(err) {
		ref, readErr := readPrivate(filepath.Join(filepath.Dir(p.path), "unlock.ref"))
		if readErr == nil {
			pass, loaded := ui.ask("load", "Unlock personal identity", "", string(ref))
			if loaded && p.Unlock([]byte(pass)) == nil {
				return
			}
		}
		password, ok := ui.ask("password", "Unlock personal identity", "Enter your personal identity password.", "")
		if ok && p.Unlock([]byte(password)) != nil {
			ui.ask("notice", "Personal identity locked", "Check your password and recovery copy.", "")
		}
		return
	}
	action, ok := ui.ask("choose", "Your Apparat identity", "Use the same personal identity on each of your devices to administer your cluster. Restore your existing identity if you have one.", "", "Restore personal identity", "Create personal identity")
	if !ok {
		return
	}
	var password string
	switch action {
	case "Restore personal identity":
		raw, selected := ui.ask("file", "Restore personal identity", "Select your encrypted personal identity backup.", "")
		if !selected {
			return
		}
		data, decodeErr := base64.StdEncoding.DecodeString(raw)
		var entered bool
		password, entered = ui.ask("password", "Restore personal identity", "Enter the password for the backup.", "")
		if !entered {
			return
		}
		err = decodeErr
		if err == nil {
			err = p.Import(data, []byte(password))
		}
	case "Create personal identity":
		name, entered := ui.ask("text", "Your screen name", "Enter the name to show to other cluster users.", "")
		if !entered {
			return
		}
		var confirmed bool
		password, confirmed = ui.password("Protect your portable personal identity")
		if !confirmed {
			return
		}
		err = p.Create(name, []byte(password))
	default:
		return
	}
	if err != nil {
		ui.ask("notice", "Personal identity setup", err.Error(), "")
		return
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
}

func NotifyIdentityUnavailable(dialog Dialog) {
	ui := setupUI{dialog}
	ui.ask("notice", "Personal identity locked", "Apparat could not record your device authorization. Retry after checking local storage.", "")
}

func (p *Personal) offerSystemUnlock(ui setupUI, password string) {
	native, available := ui.ask("available", "", "", "")
	if !available || native == "" {
		return
	}
	choice, ok := ui.ask("choose", "Unlock on this device", "Your recovery copy still needs its password. You can use system security to unlock the local copy.", "", "Use password", native)
	if !ok || choice != native {
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

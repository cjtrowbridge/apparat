// Coordinates owner-driven setup and recovery using private native dialogs.
// Prompts, chooser payloads and credential exchanges stay outside the action catalog and logs.
package identity

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
)

type Dialog func(string) string

type request struct {
	Op      string
	Title   string
	Message string
	Options []string
	Value   string
}

type setupUI struct{ dialog Dialog }

func (ui setupUI) ask(op, title, message, value string, choices ...string) (string, bool) {
	data, _ := json.Marshal(request{op, title, message, choices, value})
	response := ui.dialog(string(data))
	var answer *string
	err := json.Unmarshal([]byte(response), &answer)
	if err != nil || answer == nil {
		return "", false
	}
	return *answer, true
}

func (ui setupUI) password(title string) (string, bool) {
	first, ok := ui.ask("password", title, "Enter a password.", "")
	if !ok {
		return "", false
	}
	second, ok := ui.ask("password", title, "Confirm the password.", "")
	if !ok {
		return "", false
	}
	if first == "" || first != second {
		ui.ask("notice", title, "Enter the same nonempty password twice.", "")
		return "", false
	}
	return first, true
}

func (s *Session) Setup(dialog Dialog) {
	ui := setupUI{dialog}
	for {
		state := s.State()
		if state == "unlocked" {
			return
		}
		if state == "plain" && s.Unlock("") == nil {
			return
		}
		if state == "password" || state == "system" {
			action, ok := ui.ask("choose", "Unlock Apparat node", "Use your existing node identity.", "", "Unlock", "Recover keys")
			if !ok {
				return
			}
			if action == "Unlock" {
				var secret string
				if state == "system" {
					secret, ok = ui.ask("load", "Unlock node", "", s.Reference())
				} else {
					secret, ok = ui.ask("password", "Unlock node", "Enter the node key password.", "")
				}
				if ok && s.Unlock(secret) == nil {
					return
				}
				if ok {
					ui.ask("notice", "Key remains locked", "The credential did not unlock this key.", "")
				}
				continue
			}
		}
		action, ok := ui.ask("choose", "Apparat node keys", RecoveryWarning, "", "Find existing keys", "Generate new keys")
		if !ok {
			return
		}
		generate := action == "Generate new keys"
		var backup []byte
		var old string
		if generate {
			answer, confirmed := ui.ask("choose", "Create a new node identity", RecoveryWarning, "", "Generate new identity")
			if !confirmed || answer != "Generate new identity" {
				continue
			}
		} else if action == "Find existing keys" {
			selected, picked := ui.ask("file", "Find Apparat private key", "Choose device.key.json or a password-protected Apparat key backup.", "")
			if !picked {
				continue
			}
			var err error
			backup, err = base64.StdEncoding.DecodeString(selected)
			if err != nil || len(backup) == 0 || len(backup) > 16384 {
				ui.ask("notice", "Recovery", "Invalid key file.", "")
				continue
			}
			if len(backup) != 64 {
				old, ok = ui.ask("password", "Restore private key", "Enter the password used for this key.", "")
				if !ok {
					continue
				}
			}
		} else {
			return
		}
		err := s.protect(ui, backup, old, generate)
		clear(backup)
		if err != nil {
			ui.ask("notice", "Identity setup", err.Error(), "")
		}
	}
}

func (s *Session) protect(ui setupUI, backup []byte, old string, generate bool) error {
	options := []string{"Create a password"}
	native, available := ui.ask("available", "", "", "")
	if available && native != "" {
		options = append(options, native)
	}
	choice, ok := ui.ask("choose", "Protect your node key", "Choose how to unlock this identity on this device.", "", options...)
	if !ok {
		return nil
	}
	mode := "password"
	var reference, secret string
	if choice == "Create a password" {
		secret, ok = ui.password("Protect node key")
	} else if available && choice == native {
		mode = "system"
		var random [32]byte
		_, err := rand.Read(random[:])
		if err != nil {
			return errors.New("could not generate protection secret")
		}
		secret = base64.StdEncoding.EncodeToString(random[:])
		clear(random[:])
		reference, ok = ui.ask("store", "Protect node key", "", secret)
		if !ok || reference == "" {
			return errors.New("system credential enrollment was cancelled or unavailable")
		}
	} else {
		return errors.New("unknown protection choice")
	}
	if !ok {
		return nil
	}
	err := s.Install(backup, old, mode, secret, reference, generate)
	if err != nil {
		if mode == "system" {
			ui.ask("remove", "", "", reference)
		}
		return err
	}
	choice, ok = ui.ask("choose", "Node identity ready", "Save a portable, password-protected backup to restore your cluster connections if you lose this device or its system credentials.", "", "Save recovery copy", "Finish")
	if ok && choice == "Save recovery copy" {
		return s.Backup(ui.dialog)
	}
	return nil
}

func (s *Session) Backup(dialog Dialog) error {
	ui := setupUI{dialog}
	pwd, ok := ui.password("Protect recovery copy")
	if !ok {
		return nil
	}
	s.mu.Lock()
	data, err := encryptKey(s.key, []byte(pwd))
	s.mu.Unlock()
	if err != nil {
		return errors.New("unlock the node before exporting a backup")
	}
	_, ok = ui.ask("save", "Save Apparat key backup", "", base64.StdEncoding.EncodeToString(data))
	if !ok {
		return errors.New("recovery copy was not saved")
	}
	return nil
}

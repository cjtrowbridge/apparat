// Coordinates the local confirmation, scoped TLS request and durable worker registration.
// Untrusted discovery names are labels. The private invitation supplies the trust anchor.
package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func (node *Runtime) ClaimWithDialog(ctx context.Context, host string, dialog identity.Dialog) {
	if !node.claimMu.TryLock() {
		return
	}
	defer node.claimMu.Unlock()

	if node.Personal == nil {
		return
	}
	raw, ok := nativeQuestion(dialog, "file", "Select worker invitation", "Choose the invitation issued locally on "+host+".")
	if !ok {
		return
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	var invite httpapi.Invitation
	if err != nil || len(data) > 4096 || json.Unmarshal(data, &invite) != nil || !strings.EqualFold(invite.Host, host) {
		nativeQuestion(dialog, "notice", "Claim worker", "The invitation does not identify the selected worker.")
		return
	}
	choice, ok := nativeQuestion(dialog, "choose", "Claim "+host, "Use your personal identity to become administrator of "+host+" with fingerprint "+invite.Fingerprint+"?", "Claim host")
	if !ok || choice != "Claim host" {
		return
	}
	err = node.claim(ctx, invite)
	if err != nil {
		nativeQuestion(dialog, "notice", "Claim worker", err.Error())
		return
	}
	nativeQuestion(dialog, "notice", "Worker claimed", host+" has approved your personal identity as its first administrator.")
}

func (node *Runtime) claim(ctx context.Context, invite httpapi.Invitation) error {
	profile, _, err := node.Personal.Credentials(node.Keys.Public())
	if err != nil {
		return errors.New("unlock your personal and device identities first")
	}
	cluster := identity.ClaimDigest("apparat/personal-cluster/v1/" + identity.Fingerprint(profile.Profile.Public))
	c, err := node.Personal.SignClaim(node.Keys, cluster, invite.Fingerprint, invite.Code)
	if err != nil {
		return err
	}
	err = httpapi.PostClaim(ctx, invite, c)
	if err != nil {
		return err
	}
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.database == nil {
		return errors.New("worker claimed; reopen and retry the invitation to record it locally")
	}
	return node.database.RememberClaim(ctx, c, invite.Host, invite.Address)
}

func nativeQuestion(dialog identity.Dialog, op, title, message string, options ...string) (string, bool) {
	body, _ := json.Marshal(map[string]any{"Op": op, "Title": title, "Message": message, "Options": options})
	var answer *string
	err := json.Unmarshal([]byte(dialog(string(body))), &answer)
	if err != nil || answer == nil {
		return "", false
	}
	return *answer, true
}

func (node *Runtime) EditScreenName(ctx context.Context, dialog identity.Dialog) {
	if node.Personal == nil || !node.claimMu.TryLock() {
		return
	}
	defer node.claimMu.Unlock()
	profile, _, err := node.Personal.Credentials(node.Keys.Public())
	if err != nil {
		nativeQuestion(dialog, "notice", "Personal identity", "Unlock your personal and device identities first.")
		return
	}
	name, ok := nativeQuestion(dialog, "text", "Edit screen name", "Current screen name: "+profile.Profile.Name)
	if !ok {
		return
	}
	err = node.Personal.Rename(name)
	if err == nil {
		var binding identity.SignedBinding
		profile, binding, err = node.Personal.Credentials(node.Keys.Public())
		node.mu.Lock()
		if err == nil && node.database != nil {
			err = node.database.RememberDevice(ctx, profile, binding)
		}
		node.mu.Unlock()
	}
	if err != nil {
		nativeQuestion(dialog, "notice", "Personal identity", err.Error())
	}
}

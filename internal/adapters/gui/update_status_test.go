//go:build gui

package gui

import (
	"strings"
	"testing"
)

func TestUpdateStatusCodesHaveSafeActionableText(t *testing.T) {
	for _, code := range []UpdateStatusCode{
		UpdateStatusChecking,
		UpdateStatusUpToDate,
		UpdateStatusAvailable,
		UpdateStatusUnableNoChecker,
		UpdateStatusUnableBridge,
		UpdateStatusUnableNetwork,
		UpdateStatusUnableHTTP,
		UpdateStatusUnableDownload,
		UpdateStatusUnableVerification,
		UpdateStatusPermissionNeeded,
		UpdateStatusInstallerOpened,
		UpdateStatusUnableInstall,
	} {
		status := updateStatusFor(code)
		if status.Label == "" || status.Detail == "" || status.Reason != code {
			t.Fatalf("status %q = %#v, want label, detail, and matching reason", code, status)
		}
	}
}

func TestUnknownUpdateStatusCodeIsRedacted(t *testing.T) {
	status := updateStatusFor("server error: token=secret path=C:\\private")
	if status.Reason != UpdateStatusUnableNoChecker {
		t.Fatalf("unknown status reason = %q, want %q", status.Reason, UpdateStatusUnableNoChecker)
	}
	for _, forbidden := range []string{"token", "secret", "private", "server error"} {
		if strings.Contains(strings.ToLower(status.Label+status.Detail), forbidden) {
			t.Fatalf("unknown status exposed %q in %#v", forbidden, status)
		}
	}
}

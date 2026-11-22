// Prints a short-lived worker invitation to a local operator for private transfer to a personal app.
// The operator selects a numeric address of an already-advertised, explicitly granted listener.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func claimInvitation(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: apparatd claim-invite IP:port")
		return 2
	}
	addr, err := netip.ParseAddrPort(args[0])
	if err != nil || addr.Port() == 0 || addr.Addr().IsMulticast() || addr.Addr().IsUnspecified() {
		return 2
	}
	ctx := context.Background()
	cfg, err := config.Load(config.Options{DefaultMode: config.ModeHeadless, BinaryName: "apparatd"})
	if err != nil {
		return 1
	}
	session := identity.NewSession(cfg.IdentityDir)
	defer session.Lock()
	if session.State() != "plain" || session.Unlock("") != nil {
		fmt.Fprintln(os.Stderr, "Unattended worker identity required.")
		return 1
	}
	db, err := persistence.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return 1
	}
	defer db.Close()
	code, err := db.IssueWorkerInvitation(ctx, session.Public())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot issue invitation. Ensure the current worker has started and is unclaimed.")
		return 1
	}
	host, err := os.Hostname()
	if err != nil {
		return 1
	}
	invite := httpapi.Invitation{Host: strings.Split(host, ".")[0], Address: addr.String(), Fingerprint: identity.Fingerprint(session.Public()), Code: code}
	err = json.NewEncoder(os.Stdout).Encode(invite)
	if err != nil {
		return 1
	}
	return 0
}

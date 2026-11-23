// Offers offline owner setup for unattended workers, using the same node lock and key workflow.
// Python's getpass reads passwords directly from the controlling terminal.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"github.com/cjtrowbridge/apparat/internal/platform"
)

const terminalDialog = `
import base64,getpass,json,os,sys
from pathlib import Path
r=json.load(sys.stdin)
def ask():
    op=r['Op']
    if op=='available':return None
    with open('/dev/tty','w') as tty, open('/dev/tty','r') as source:
        def prompt(text):
            tty.write(text+' ')
            tty.flush()
            return source.readline().rstrip('\n')
        tty.write(r['Title']+'\n'+r['Message']+'\n')
        if op=='text':return prompt('Hostname:') or None
        if op=='password':return getpass.getpass('Password: ',stream=tty) or None
        if op=='choose':
            for i,choice in enumerate(r['Options'],1):tty.write(str(i)+'. '+choice+'\n')
            selection=prompt('Choose a number, or press Enter to cancel:')
            if selection.isdigit() and 0<int(selection)<=len(r['Options']):return r['Options'][int(selection)-1]
        if op=='file':
            with open(Path(prompt('Private key file:')).expanduser(),'rb') as f:data=f.read(16385)
            if len(data)<=16384:return base64.b64encode(data).decode()
        if op=='save':
            with open(Path(prompt('New backup path:')).expanduser(),'xb',opener=lambda p,f:os.open(p,f,0o600)) as out:out.write(base64.b64decode(r['Value']))
            return ''
try:print(json.dumps(ask()))
except (EOFError,OSError,KeyboardInterrupt):print('null')
`

func setupIdentity(args []string) int {
	generate := len(args) > 0 && args[0] == "--generate-if-missing"
	if generate {
		args = args[1:]
	}
	cfg, err := config.Load(config.Options{Args: args, DefaultMode: config.ModeHeadless, BinaryName: "apparatd"})
	if err != nil {
		return 2
	}
	lock, err := platform.Acquire(cfg.RootDir, "apparatd", "identity setup")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Stop the active node before offline identity setup.")
		return 1
	}
	defer func() {
		_ = lock.Close()
	}()
	session := identity.NewSession(cfg.IdentityDir)
	defer session.Lock()
	dialog := func(packet string) string {
		cmd := exec.CommandContext(context.Background(), "/usr/bin/python3", "-I", "-c", terminalDialog)
		cmd.Stdin = strings.NewReader(packet)
		out, err := cmd.Output()
		if err != nil {
			return "null"
		}
		return string(out)
	}
	if !session.SetupHostname(dialog) {
		fmt.Fprintln(os.Stderr, "Set a recognizable device hostname with apparatd identity-setup from a terminal.")
		return 1
	}
	if generate {
		if session.State() != "missing" {
			fmt.Fprintln(os.Stderr, "A device identity already exists or needs recovery.")
			return 1
		}
		err = session.Install(nil, "", "plain", "", "", true)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not create device identity.")
			return 1
		}
		fmt.Println("Created worker identity", identity.Fingerprint(session.Public()))
		return 0
	}
	session.Setup(dialog)
	fmt.Println("Node identity:", session.State())
	return 0
}

func trustHost(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "Usage: apparatd trust-host [OWNER_FINGERPRINT]")
		return 2
	}
	cfg, err := config.Load(config.Options{DefaultMode: config.ModeHeadless})
	if err != nil {
		return 1
	}
	pin, err := identity.DeviceFingerprint(cfg.IdentityDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("Worker fingerprint:", pin)
	if len(args) == 0 {
		return 0
	}
	ctx := context.Background()
	db, err := persistence.Open(ctx, cfg.DatabasePath)
	if err == nil {
		defer db.Close()
		err = db.ConfirmOwner(ctx, pin, args[0])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("Confirmed owner fingerprint:", args[0])
	return 0
}

func workspaceAccess(args []string) int {
	status := len(args) == 3 && args[0] == "status"
	if !status && (len(args) != 4 || (args[0] != "grant" && args[0] != "revoke")) {
		fmt.Fprintln(os.Stderr, "Usage: apparatd workspace-access status PROJECT_ID USER_FINGERPRINT")
		fmt.Fprintln(os.Stderr, "Usage: apparatd workspace-access grant|revoke PROJECT_ID USER_FINGERPRINT read|write|tasks|terminal-view|terminal-control")
		return 2
	}
	cfg, err := config.Load(config.Options{DefaultMode: config.ModeHeadless})
	if err != nil {
		return 1
	}
	ctx := context.Background()
	db, err := persistence.Open(ctx, cfg.DatabasePath)
	if err == nil {
		defer db.Close()
		if status {
			var scopes []string
			scopes, err = db.WorkspaceScopes(ctx, args[1], args[2])
			if err == nil {
				if len(scopes) == 0 {
					fmt.Println("Registered checkout. No permissions for this user.")
				} else {
					fmt.Println("Checkout permissions:", strings.Join(scopes, ", "))
				}
				return 0
			}
		} else {
			err = db.WorkspaceGrant(ctx, args[1], args[2], args[3], args[0] == "grant")
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("Workspace access updated")
	return 0
}

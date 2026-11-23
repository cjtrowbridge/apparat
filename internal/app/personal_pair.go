// Runs explicit two-device pairing from personal app screens using local native dialogs.
// The Linux sender is reachable solely through its existing network-specific API grant.
package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/discovery"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func (n *Runtime) PairDialog(ctx context.Context, network, host string, dialog identity.Dialog) {
	if n.Personal == nil || !n.claimMu.TryLock() {
		return
	}
	defer n.claimMu.Unlock()
	err := n.personalPair(ctx, network, host, dialog)
	if err != nil {
		nativeQuestion(dialog, "notice", "Pair personal devices", err.Error())
	}
}

func (n *Runtime) personalPair(ctx context.Context, network, host string, dialog identity.Dialog) error {
	if len(n.Keys.Public()) == 0 {
		n.Keys.Setup(dialog)
		if len(n.Keys.Public()) == 0 {
			return nil
		}
	}
	var listener *net.UnixListener
	if host == "" {
		if n.Personal.Fingerprint() == "" {
			return errors.New("Unlock your personal identity before pairing another device.")
		}
		out, err := discovery.Command(ctx, "list", n.config.DatabasePath).Output()
		var rules struct{ Enabled []string }
		if err != nil || json.Unmarshal(out, &rules) != nil {
			return errors.New("Could not read enabled network grants.")
		}
		var rule string
		for _, item := range rules.Enabled {
			if strings.HasPrefix(item, network+"/") {
				if rule != "" {
					return errors.New("Select one API network grant for this interface.")
				}
				rule = item
			}
		}
		if rule == "" {
			return errors.New("Enable API hosting for this network in Settings first.")
		}
		hash := sha256.Sum256([]byte(rule))
		name := "pair-" + hex.EncodeToString(hash[:8]) + ".sock"
		addr := &net.UnixAddr{Name: filepath.Join(filepath.Dir(n.config.DatabasePath), name), Net: "unix"}
		listener, err = net.ListenUnix("unix", addr)
		if err != nil {
			return errors.New("Could not open a private pairing session.")
		}
		defer listener.Close()
		err = os.Chmod(addr.Name, 0600)
		if err != nil {
			return err
		}
	}
	localCtx := ctx
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	code, err := httpapi.PairCode()
	if err != nil {
		return err
	}
	other, ok := nativeQuestion(dialog, "text", "Pair personal devices", "Your code is "+code+". Enter the four-digit code shown on the other device to authorize pairing.")
	if !ok {
		return nil
	}
	if !httpapi.ValidPairCode(other) {
		return errors.New("Enter all four digits from the other device.")
	}
	if listener != nil {
		deadline, _ := ctx.Deadline()
		listener.SetDeadline(deadline)
		stop := context.AfterFunc(ctx, func() {
			listener.Close()
		})
		defer stop()
		conn, err := listener.Accept()
		if err != nil {
			return errors.New("Pairing expired. Start again on both devices.")
		}
		defer conn.Close()
		wire, err := httpapi.PairTLS(conn, n.Keys, true)
		if err != nil {
			return err
		}
		err = httpapi.SendPersonalPair(ctx, wire, code, other, func() ([]byte, error) {
			devices, err := n.KnownWorkers(ctx)
			if err != nil {
				return nil, err
			}
			return n.Personal.PairOffer(httpapi.PairDeviceKey(wire), devices)
		})
		if err == nil {
			err = n.rememberGUIPeer(localCtx, httpapi.PairedPeer(wire))
		}
		if err == nil {
			nativeQuestion(dialog, "notice", "Personal pairing", "Encrypted identity delivered. Finish unlocking it on the other device.")
		}
		return err
	}
	address, err := n.resolvePeer(ctx, network, host)
	if err != nil {
		return err
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()
	wire, err := httpapi.PairTLS(conn, n.Keys, false)
	if err != nil {
		return err
	}
	raw, err := httpapi.ReceivePersonalPair(ctx, wire, other, code)
	if err != nil {
		return err
	}
	defer clear(raw)
	password, ok := nativeQuestion(dialog, "password", "Unlock paired identity", "Enter your personal identity recovery password on this device.")
	if !ok {
		return nil
	}
	devices, err := n.Personal.AcceptPair(raw, []byte(password), n.Keys.Public())
	if err != nil {
		return err
	}
	profile, binding, err := n.Personal.Credentials(n.Keys.Public())
	if err != nil {
		return err
	}
	n.mu.Lock()
	if n.database == nil {
		n.mu.Unlock()
		return errors.New("Local database unavailable.")
	}
	defer n.RequestSync()
	err = n.database.RememberPairedDevice(localCtx, profile, binding, devices)
	if err == nil && n.serviceManager != nil {
		n.serviceManager.notify()
	}
	n.mu.Unlock()
	if err == nil {
		err = n.rememberGUIPeer(localCtx, httpapi.PairedPeer(wire))
	}
	if err == nil {
		n.Personal.OfferPairedUnlock(dialog, password, n.personalUnlockDevice())
	}
	return err
}

func (n *Runtime) personalUnlockDevice() *identity.Session {
	if runtime.GOOS == "android" {
		return n.Keys
	}
	return nil
}

func (n *Runtime) rememberGUIPeer(ctx context.Context, d domain.PoolDevice) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.database == nil {
		return domain.ErrUnavailable
	}
	return n.database.RememberGUIPeer(ctx, n.Personal.Fingerprint(), d)
}
func (n *Runtime) GUIPeers(ctx context.Context) ([]domain.PoolDevice, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.database == nil {
		return nil, domain.ErrUnavailable
	}
	if n.Personal == nil {
		return nil, nil
	}
	return n.database.GUIPeers(ctx, n.Personal.Fingerprint())
}

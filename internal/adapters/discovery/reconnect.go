// Resolves a saved hostname only on the single local interface of its previous subnet.
// Ambiguous, moved or unavailable networks require selecting the peer in discovery again.
package discovery

import (
	"context"
	"errors"
	"net"
	"net/netip"
)

func ResolveSaved(ctx context.Context, host, previous string) (string, error) {
	addr, err := netip.ParseAddrPort(previous)
	if err != nil {
		return "", err
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var selected string
	for _, i := range interfaces {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagMulticast == 0 {
			continue
		}
		prefixes, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, p := range prefixes {
			prefix, err := netip.ParsePrefix(p.String())
			if err == nil && prefix.Contains(addr.Addr()) {
				if selected != "" && selected != i.Name {
					return "", errors.New("ambiguous saved network")
				}
				selected = i.Name
			}
		}
	}
	if selected == "" {
		return "", errors.New("previous network is unavailable")
	}
	return ResolveWorker(ctx, selected, host)
}

// Browses link-local DNS-SD from a selected IPv4 interface using unicast replies.
// Expires bounded candidate labels and closes the socket when the view exits.
package discovery

import (
	"context"
	"errors"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/net/ipv4"
	"net"
	"slices"
	"sort"
	"strings"
	"time"
)

func Browse(ctx context.Context, device string, update func([]string)) error {
	iface, err := net.InterfaceByName(device)
	if err != nil {
		return err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}
	var local net.IP
	for _, a := range addrs {
		ip, _, _ := net.ParseCIDR(a.String())
		if ip.To4() != nil {
			local = ip.To4()
			break
		}
	}
	if local == nil {
		return errors.New("interface has no IPv4 address")
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: local})
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()
	packet := ipv4.NewPacketConn(conn)
	err = packet.SetMulticastInterface(iface)
	if err != nil {
		return err
	}
	err = packet.SetMulticastTTL(255)
	if err != nil {
		return err
	}
	name := dnsmessage.MustNewName("_apparat._tcp.local.")
	query := dnsmessage.Message{Header: dnsmessage.Header{ID: 1}, Questions: []dnsmessage.Question{{Name: name, Type: dnsmessage.TypePTR, Class: dnsmessage.ClassINET}}}
	data, err := query.Pack()
	if err != nil {
		return err
	}
	endpoint := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}
	expires := map[string]time.Time{}
	var previous []string
	update(nil)
	var next time.Time
	buf := make([]byte, 9000)
	for ctx.Err() == nil {
		now := time.Now()
		if now.After(next) {
			_, err = conn.WriteToUDP(data, endpoint)
			if err != nil {
				return err
			}
			next = now.Add(5 * time.Second)
		}
		err = conn.SetReadDeadline(now.Add(time.Second))
		if err != nil {
			return err
		}
		n, from, err := conn.ReadFromUDP(buf)
		var timeout net.Error
		if err != nil && (!errors.As(err, &timeout) || !timeout.Timeout()) {
			return err
		}
		var response dnsmessage.Message
		if err == nil && from.Port == 5353 && response.Unpack(buf[:n]) == nil && response.Response {
			for _, record := range response.Answers {
				ptr, ok := record.Body.(*dnsmessage.PTRResource)
				if !ok || !strings.EqualFold(record.Header.Name.String(), name.String()) {
					continue
				}
				full := ptr.PTR.String()
				suffix := "." + name.String()
				if len(full) <= len(suffix) || !strings.EqualFold(full[len(full)-len(suffix):], suffix) {
					continue
				}
				label := full[:len(full)-len(suffix)]
				if record.Header.TTL == 0 {
					delete(expires, label)
				} else if !expires[label].IsZero() || len(expires) < 128 {
					expires[label] = now.Add(time.Duration(min(record.Header.TTL, 30)) * time.Second)
				}
			}
		}
		var names []string
		for label, expiry := range expires {
			if now.After(expiry) {
				delete(expires, label)
			} else {
				names = append(names, label)
			}
		}
		sort.Strings(names)
		if !slices.Equal(previous, names) {
			update(names)
			previous = names
		}
	}
	return ctx.Err()
}

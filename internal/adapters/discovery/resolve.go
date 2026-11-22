// Resolves a selected DNS-SD instance on its browse interface for one user action.
// A worker's SRV target is its advertising host. Its TLS identity is verified separately.
package discovery

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/net/ipv4"
)

func ResolveWorker(ctx context.Context, device, host string) (string, error) {
	iface, err := net.InterfaceByName(device)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, a := range addrs {
		ip, _, err := net.ParseCIDR(a.String())
		if err == nil && ip.To4() != nil {
			return resolveWorker(ctx, iface, ip, host)
		}
	}
	return "", errors.New("selected network has no IPv4 address")
}

func resolveWorker(ctx context.Context, iface *net.Interface, local net.IP, host string) (string, error) {
	name, err := dnsmessage.NewName(host + "._apparat._tcp.local.")
	if err != nil {
		return "", err
	}
	q := dnsmessage.Message{Header: dnsmessage.Header{ID: 1}, Questions: []dnsmessage.Question{{Name: name, Type: dnsmessage.TypeSRV, Class: dnsmessage.ClassINET}}}
	packet, err := q.Pack()
	if err != nil {
		return "", err
	}
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: local})
	if err != nil {
		return "", err
	}
	defer c.Close()
	err = ipv4.NewPacketConn(c).SetMulticastInterface(iface)
	if err != nil {
		return "", err
	}
	dest := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}
	buf := make([]byte, 9000)
	for attempt := 0; attempt < 3 && ctx.Err() == nil; attempt++ {
		_, err = c.WriteToUDP(packet, dest)
		if err != nil {
			return "", err
		}
		err = c.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			return "", err
		}
		for {
			n, from, err := c.ReadFromUDP(buf)
			if err != nil {
				break
			}
			var msg dnsmessage.Message
			if from.Port != 5353 || msg.Unpack(buf[:n]) != nil || !msg.Response {
				continue
			}
			for _, r := range msg.Answers {
				srv, ok := r.Body.(*dnsmessage.SRVResource)
				if ok && r.Header.TTL > 0 && srv.Port != 0 && strings.EqualFold(r.Header.Name.String(), name.String()) && strings.EqualFold(srv.Target.String(), host+".local.") {
					return net.JoinHostPort(from.IP.String(), strconv.Itoa(int(srv.Port))), nil
				}
			}
		}
	}
	return "", errors.New("the worker is no longer visible on the selected network")
}

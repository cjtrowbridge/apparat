// Defines the portable hostname label used in setup, pairing and cluster policy.
// Generic OS names and loopback addresses cannot identify a device to its peers.
package domain

import (
	"regexp"
	"strings"
)

var hostnameLabel = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

var genericAndroidName = regexp.MustCompile(`(?i)^android-[0-9a-f]{8,}$`)

func ValidDeviceName(name string) bool {
	switch strings.ToLower(name) {
	case "localhost", "localhost4", "localhost6", "ip6-localhost", "ip6-loopback", "loopback", "android", "unknown", "none":
		return false
	}
	return strings.Trim(name, "0123456789") != "" && hostnameLabel.MatchString(name) && !genericAndroidName.MatchString(name)
}

func SystemDeviceName(host string) string {
	host = strings.TrimSuffix(host, ".")
	parts := strings.Split(host, ".")
	for _, part := range parts {
		if !ValidDeviceName(part) {
			return ""
		}
	}
	return parts[0]
}

func (w ClusterDevice) DisplayName() string {
	for _, d := range w.Roster {
		if d.Worker && d.ID == w.Fingerprint {
			return d.DisplayName(w.Host)
		}
	}
	return (PoolDevice{ID: w.Fingerprint, Worker: true}).DisplayName(w.Host)
}

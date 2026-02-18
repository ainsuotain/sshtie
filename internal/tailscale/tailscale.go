// Package tailscale provides helpers for detecting local Tailscale state
// and checking whether a remote host is reachable via the Tailscale network.
package tailscale

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type statusJSON struct {
	BackendState string          `json:"BackendState"`
	Peer         map[string]peer `json:"Peer"`
}

type peer struct {
	HostName     string   `json:"HostName"`
	DNSName      string   `json:"DNSName"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	Online       bool     `json:"Online"`
}

// ClientRunning returns true if Tailscale is installed and actively running
// on the local machine (BackendState == "Running").
func ClientRunning() bool {
	out, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return false
	}
	var s statusJSON
	if err := json.Unmarshal(out, &s); err != nil {
		return false
	}
	return s.BackendState == "Running"
}

// HostInNetwork returns true if host matches any Tailscale peer by:
//   - Tailscale IP (e.g. 100.x.x.x)
//   - HostName (case-insensitive)
//   - DNSName (case-insensitive, trailing dot stripped)
func HostInNetwork(host string) bool {
	out, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return false
	}
	var s statusJSON
	if err := json.Unmarshal(out, &s); err != nil {
		return false
	}
	hostLower := strings.ToLower(host)
	for _, p := range s.Peer {
		for _, ip := range p.TailscaleIPs {
			if ip == host {
				return true
			}
		}
		if strings.ToLower(p.HostName) == hostLower {
			return true
		}
		// DNSName often has a trailing dot: "host.tail.ts.net."
		dnsName := strings.TrimSuffix(strings.ToLower(p.DNSName), ".")
		if dnsName == hostLower {
			return true
		}
	}
	return false
}

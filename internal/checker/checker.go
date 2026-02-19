// Package checker provides background TCP-reachability tracking for profiles.
package checker

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ainsuotain/sshtie/internal/profile"
)

// Checker polls each server's SSH port and tracks whether it is reachable.
type Checker struct {
	mu       sync.RWMutex
	statuses map[string]bool // profile name → reachable
}

func New() *Checker {
	return &Checker{statuses: make(map[string]bool)}
}

// CheckAll dials every profile concurrently (3 s timeout each) and updates
// the internal status table. onChange is called (once) if any value changed.
func (c *Checker) CheckAll(profiles []profile.Profile, onChange func()) {
	type result struct {
		name      string
		reachable bool
	}

	results := make(chan result, len(profiles))
	for _, p := range profiles {
		go func(p profile.Profile) {
			port := p.Port
			if port == 0 {
				port = 22
			}
			conn, err := net.DialTimeout("tcp",
				net.JoinHostPort(p.Host, strconv.Itoa(port)),
				3*time.Second)
			if err == nil {
				conn.Close()
			}
			results <- result{name: p.Name, reachable: err == nil}
		}(p)
	}

	changed := false
	for range profiles {
		r := <-results
		c.mu.Lock()
		old, exists := c.statuses[r.name]
		if !exists || old != r.reachable {
			c.statuses[r.name] = r.reachable
			changed = true
		}
		c.mu.Unlock()
	}

	if changed && onChange != nil {
		onChange()
	}
}

// Get returns (reachable, known). known is false if the profile has never
// been checked yet (shows as 🟡 "checking" in the menu).
func (c *Checker) Get(name string) (reachable bool, known bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.statuses[name]
	return v, ok
}

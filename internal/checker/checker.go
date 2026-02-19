// Package checker provides background TCP-reachability and active-session
// tracking for profiles.
package checker

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/session"
)

// Checker polls each server's SSH port and tracks whether it is reachable,
// and also maintains a map of currently-active sessions.
type Checker struct {
	mu       sync.RWMutex
	statuses map[string]bool    // profile name → reachable
	sessions map[string]session.Session // profile name → active session
}

func New() *Checker {
	return &Checker{
		statuses: make(map[string]bool),
		sessions: make(map[string]session.Session),
	}
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

// RefreshSessions re-reads the session lock files and updates the internal
// sessions map.  onChange is called if the set of active sessions changed.
func (c *Checker) RefreshSessions(onChange func()) {
	active, err := session.ListActive()
	if err != nil {
		return
	}

	// Build new map.
	newMap := make(map[string]session.Session, len(active))
	for _, s := range active {
		newMap[s.Profile] = s
	}

	c.mu.Lock()
	changed := len(newMap) != len(c.sessions)
	if !changed {
		for k := range newMap {
			if _, ok := c.sessions[k]; !ok {
				changed = true
				break
			}
		}
		if !changed {
			for k := range c.sessions {
				if _, ok := newMap[k]; !ok {
					changed = true
					break
				}
			}
		}
	}
	c.sessions = newMap
	c.mu.Unlock()

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

// ActiveSession returns the Session for the named profile, if connected.
func (c *Checker) ActiveSession(name string) (session.Session, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.sessions[name]
	return s, ok
}

// ActiveSessions returns a snapshot of all currently-active sessions.
func (c *Checker) ActiveSessions() []session.Session {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]session.Session, 0, len(c.sessions))
	for _, s := range c.sessions {
		out = append(out, s)
	}
	return out
}

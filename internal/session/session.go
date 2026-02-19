// Package session tracks active sshtie connections via PID lock files.
//
// Each active connection writes a JSON file to ~/.sshtie/sessions/<name>.json.
// Because file names map 1-to-1 with profile names, multiple servers can be
// connected simultaneously — each has its own independent lock file.
package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/ainsuotain/sshtie/internal/profile"
)

// Session describes one active (or recently active) SSH connection.
type Session struct {
	Profile   string    `json:"profile"`
	PID       int       `json:"pid"`
	Method    string    `json:"method"` // "mosh", "ssh+tmux", "ssh"
	StartedAt time.Time `json:"started_at"`
}

// SessionDir returns the directory that stores session lock files.
// It creates the directory if it does not exist.
func SessionDir() (string, error) {
	dir, err := profile.ConfigDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, "sessions")
	if err := os.MkdirAll(d, 0700); err != nil {
		return "", err
	}
	return d, nil
}

// lockPath returns the path to the lock file for the given profile name.
func lockPath(name string) (string, error) {
	dir, err := SessionDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

// Write creates (or overwrites) the session lock file for this session.
func Write(s Session) error {
	path, err := lockPath(s.Profile)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Read loads the session file for the named profile.
// Returns os.ErrNotExist if the file is not present.
func Read(name string) (Session, error) {
	path, err := lockPath(name)
	if err != nil {
		return Session{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return Session{}, err
	}
	return s, nil
}

// Delete removes the session lock file for the named profile.
// It is a no-op if the file does not exist.
func Delete(name string) error {
	path, err := lockPath(name)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ListActive scans the sessions directory and returns every session whose
// process is still alive.  Stale files (dead PID) are removed automatically.
// Multiple entries can be returned when the user is connected to several
// servers at the same time.
func ListActive() ([]Session, error) {
	dir, err := SessionDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var active []Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s Session
		if err := json.Unmarshal(data, &s); err != nil {
			_ = os.Remove(path) // corrupt file
			continue
		}
		if IsAlive(s.PID) {
			active = append(active, s)
		} else {
			_ = os.Remove(path) // stale — clean up
		}
	}
	return active, nil
}

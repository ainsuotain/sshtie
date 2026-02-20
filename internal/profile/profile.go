package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile represents a single SSH connection profile.
type Profile struct {
	Name        string   `yaml:"name"`
	Host        string   `yaml:"host"`
	User        string   `yaml:"user"`
	Port        int      `yaml:"port"`
	Key         string   `yaml:"key,omitempty"`
	TmuxSession string   `yaml:"tmux_session"`
	MoshServer  string   `yaml:"mosh_server,omitempty"`
	Network     string   `yaml:"network"` // auto | tailscale | direct
	Tags        []string `yaml:"tags,omitempty"`

	// Advanced SSH options (0/false = use built-in default).
	ForwardAgent        bool `yaml:"forward_agent,omitempty"`
	ServerAliveInterval int  `yaml:"server_alive_interval,omitempty"` // default 10 s
	ServerAliveCountMax int  `yaml:"server_alive_count_max,omitempty"` // default 60
	ConnectionAttempts  int  `yaml:"connection_attempts,omitempty"`    // default 3
}

type store struct {
	Profiles []Profile `yaml:"profiles"`
}

// ConfigDir returns ~/.sshtie
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sshtie"), nil
}

func configPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles.yaml"), nil
}

// Load reads all profiles from disk. Returns empty slice if file doesn't exist yet.
func Load() ([]Profile, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Profile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read profiles: %w", err)
	}

	var s store
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse profiles.yaml: %w", err)
	}
	return s.Profiles, nil
}

// Save writes profiles to disk, creating the directory if needed.
func Save(profiles []Profile) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(dir, "profiles.yaml")
	s := store{Profiles: profiles}

	data, err := yaml.Marshal(&s)
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// Get returns the profile with the given name, or an error if not found.
func Get(name string) (Profile, error) {
	profiles, err := Load()
	if err != nil {
		return Profile{}, err
	}
	for _, p := range profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return Profile{}, fmt.Errorf("profile %q not found", name)
}

// Add appends a new profile, returning an error if the name already exists.
func Add(p Profile) error {
	profiles, err := Load()
	if err != nil {
		return err
	}
	for _, existing := range profiles {
		if existing.Name == p.Name {
			return fmt.Errorf("profile %q already exists", p.Name)
		}
	}
	profiles = append(profiles, p)
	return Save(profiles)
}

// Rename changes a profile's name from oldName to newName.
func Rename(oldName, newName string) error {
	profiles, err := Load()
	if err != nil {
		return err
	}
	for _, p := range profiles {
		if p.Name == newName {
			return fmt.Errorf("profile %q already exists", newName)
		}
	}
	found := false
	for i, p := range profiles {
		if p.Name == oldName {
			profiles[i].Name = newName
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("profile %q not found", oldName)
	}
	return Save(profiles)
}

// Remove deletes a profile by name.
func Remove(name string) error {
	profiles, err := Load()
	if err != nil {
		return err
	}
	next := profiles[:0]
	found := false
	for _, p := range profiles {
		if p.Name == name {
			found = true
			continue
		}
		next = append(next, p)
	}
	if !found {
		return fmt.Errorf("profile %q not found", name)
	}
	return Save(next)
}

// DefaultKey returns the expanded path for key, falling back to ~/.ssh/id_ed25519.
func (p Profile) DefaultKey() string {
	if p.Key != "" {
		return expandHome(p.Key)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ssh", "id_ed25519")
}

func expandHome(path string) string {
	if len(path) == 0 {
		return path
	}
	// Accept both ~/ (Unix) and ~\ (Windows) as home prefix.
	if path[0] == '~' && len(path) >= 2 && (path[1] == '/' || path[1] == '\\') {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

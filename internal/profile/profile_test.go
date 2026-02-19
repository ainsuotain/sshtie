package profile

import (
	"os"
	"testing"
)

func TestAddGetRemove(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	p := Profile{
		Name:    "myserver",
		Host:    "192.168.1.10",
		User:    "alice",
		Port:    22,
		Network: "auto",
	}

	if err := Add(p); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Duplicate should fail.
	if err := Add(p); err == nil {
		t.Error("expected error on duplicate Add")
	}

	got, err := Get("myserver")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Host != p.Host || got.User != p.User {
		t.Errorf("Get mismatch: got %+v, want %+v", got, p)
	}

	if err := Remove("myserver"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := Get("myserver"); err == nil {
		t.Error("expected error after Remove")
	}
}

func TestAdvancedFields_savedAndLoaded(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	p := Profile{
		Name:                "adv",
		Host:                "10.0.0.1",
		User:                "bob",
		Network:             "direct",
		ForwardAgent:        true,
		ServerAliveInterval: 30,
		ServerAliveCountMax: 20,
		ConnectionAttempts:  5,
	}
	if err := Add(p); err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := Get("adv")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.ForwardAgent {
		t.Error("ForwardAgent should be true")
	}
	if got.ServerAliveInterval != 30 {
		t.Errorf("ServerAliveInterval: got %d, want 30", got.ServerAliveInterval)
	}
	if got.ServerAliveCountMax != 20 {
		t.Errorf("ServerAliveCountMax: got %d, want 20", got.ServerAliveCountMax)
	}
	if got.ConnectionAttempts != 5 {
		t.Errorf("ConnectionAttempts: got %d, want 5", got.ConnectionAttempts)
	}
}

func TestLoad_emptyFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	profiles, err := Load()
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected empty slice, got %d", len(profiles))
	}
}

func TestDefaultKey_fallback(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	p := Profile{Name: "x", Host: "h", User: "u"}
	key := p.DefaultKey()
	if key == "" {
		t.Error("DefaultKey should not be empty")
	}
}

func TestGet_notFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if _, err := Get("nonexistent"); err == nil {
		t.Error("Get on nonexistent profile should error")
	}
}

func TestMultipleProfiles_order(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	names := []string{"alpha", "beta", "gamma"}
	for _, n := range names {
		_ = Add(Profile{Name: n, Host: "h", User: "u", Network: "auto"})
	}

	profiles, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(profiles))
	}
	for i, n := range names {
		if profiles[i].Name != n {
			t.Errorf("order[%d]: got %s, want %s", i, profiles[i].Name, n)
		}
	}
}

// Ensure HOME is always set (belt-and-suspenders for CI).
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

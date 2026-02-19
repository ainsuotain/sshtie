package session

import (
	"os"
	"testing"
	"time"
)

func TestWriteReadDelete(t *testing.T) {
	// Use a temp dir so tests don't touch ~/.sshtie
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	s := Session{
		Profile:   "testserver",
		PID:       os.Getpid(),
		Method:    "ssh+tmux",
		StartedAt: time.Now(),
	}

	if err := Write(s); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := Read("testserver")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Profile != s.Profile || got.PID != s.PID || got.Method != s.Method {
		t.Errorf("Read mismatch: got %+v, want %+v", got, s)
	}

	if err := Delete("testserver"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := Read("testserver"); !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist after Delete, got %v", err)
	}
}

func TestListActive_multipleProfiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	pid := os.Getpid()
	for _, name := range []string{"serverA", "serverB", "serverC"} {
		if err := Write(Session{Profile: name, PID: pid, Method: "ssh"}); err != nil {
			t.Fatalf("Write %s: %v", name, err)
		}
	}

	active, err := ListActive()
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(active) != 3 {
		t.Errorf("expected 3 active sessions, got %d", len(active))
	}
}

func TestListActive_staleRemoved(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// PID 1 is always alive (init/launchd), PID 999999999 is certainly dead
	_ = Write(Session{Profile: "alive", PID: 1, Method: "ssh"})
	_ = Write(Session{Profile: "dead", PID: 999999999, Method: "ssh"})

	active, err := ListActive()
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	// "dead" should be removed
	for _, s := range active {
		if s.Profile == "dead" {
			t.Error("stale session 'dead' should have been removed")
		}
	}
	// Stale file should be gone from disk
	if _, err := Read("dead"); !os.IsNotExist(err) {
		t.Error("stale file should be deleted after ListActive")
	}
}

func TestIsAlive_currentProcess(t *testing.T) {
	if !IsAlive(os.Getpid()) {
		t.Error("IsAlive(os.Getpid()) should return true")
	}
}

func TestIsAlive_deadPID(t *testing.T) {
	if IsAlive(999999999) {
		t.Error("IsAlive(999999999) should return false")
	}
}

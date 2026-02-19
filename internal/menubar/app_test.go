//go:build darwin || windows

package menubar

import (
	"testing"

	"github.com/ainsuotain/sshtie/internal/profile"
)

func TestNextInterval(t *testing.T) {
	cases := []struct {
		input int
		want  int
	}{
		{0, 10},   // 0 → snap to first preset (10)
		{10, 30},  // at 10 → next is 30
		{15, 30},  // between 10 and 30 → snap to 30
		{30, 60},  // at 30 → next is 60
		{59, 60},  // between 30 and 60 → snap to 60
		{60, 10},  // at max → wrap to 10
		{99, 10},  // beyond max → wrap to 10
	}
	for _, c := range cases {
		got := nextInterval(c.input)
		if got != c.want {
			t.Errorf("nextInterval(%d) = %d, want %d", c.input, got, c.want)
		}
	}
}

func TestIntervalLabel(t *testing.T) {
	p := profile.Profile{ServerAliveInterval: 30}
	if got := intervalLabel(p); got != "Interval: 30s" {
		t.Errorf("intervalLabel: got %q", got)
	}
	// Zero value → default 10
	p2 := profile.Profile{}
	if got := intervalLabel(p2); got != "Interval: 10s" {
		t.Errorf("intervalLabel zero: got %q", got)
	}
}

func TestFaLabel(t *testing.T) {
	p := profile.Profile{ForwardAgent: true}
	if got := faLabel(p); got != "Forward agent: on" {
		t.Errorf("faLabel on: got %q", got)
	}
	p.ForwardAgent = false
	if got := faLabel(p); got != "Forward agent: off" {
		t.Errorf("faLabel off: got %q", got)
	}
}

func TestMenuLabel(t *testing.T) {
	// active + reachable → green + bullet
	got := menuLabel("srv", true, true, true)
	if got != "🟢  ● srv" {
		t.Errorf("menuLabel active reachable: %q", got)
	}
	// inactive + unreachable → red
	got = menuLabel("srv", false, true, false)
	if got != "🔴  srv" {
		t.Errorf("menuLabel inactive unreachable: %q", got)
	}
	// unknown (checking)
	got = menuLabel("srv", false, false, false)
	if got != "🟡  srv" {
		t.Errorf("menuLabel checking: %q", got)
	}
}

func TestPortOf(t *testing.T) {
	if portOf(profile.Profile{Port: 0}) != 22 {
		t.Error("portOf default should be 22")
	}
	if portOf(profile.Profile{Port: 2222}) != 2222 {
		t.Error("portOf custom")
	}
}

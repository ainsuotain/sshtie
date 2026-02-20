package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ainsuotain/sshtie/internal/profile"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	editTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	editSub      = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	editActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	editInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	editBar      = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	editBarEmpty = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	editVal      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	editMeta     = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	editSaved    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("72"))
	editErr      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	editHint     = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	editWarn     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
)

const barWidth = 20

// Row indices: 0-3 = sliders/toggles, 4 = rename, 5 = delete.
const (
	numSliderOpts = 4
	optRename     = 4
	optDelete     = 5
	numOpts       = 6
)

// ── Option definitions ────────────────────────────────────────────────────────

type editOpt struct {
	label  string
	min    int
	max    int
	value  int
	unit   string // displayed after value, e.g. "s"
	isBool bool   // renders as on/off toggle instead of slider
	extra  func(v int) string // optional right-side annotation
}

func defaultOpts(p profile.Profile) []editOpt {
	aliveInterval := p.ServerAliveInterval
	if aliveInterval <= 0 { aliveInterval = 10 }
	aliveCount := p.ServerAliveCountMax
	if aliveCount <= 0 { aliveCount = 60 }
	attempts := p.ConnectionAttempts
	if attempts <= 0 { attempts = 3 }

	fa := 0
	if p.ForwardAgent { fa = 1 }

	return []editOpt{
		{
			label: "Connection attempts",
			min: 1, max: 10, value: attempts,
			extra: func(v int) string {
				if v == 1 { return "give up after 1st failure" }
				return fmt.Sprintf("retry up to %d times", v-1)
			},
		},
		{
			label: "Alive interval",
			min: 10, max: 60, value: aliveInterval, unit: "s",
			extra: func(v int) string {
				return fmt.Sprintf("ping every %ds", v)
			},
		},
		{
			label: "Alive count max",
			min: 6, max: 120, value: aliveCount,
			extra: func(v int) string {
				return fmt.Sprintf("drop after %d missed pings", v)
			},
		},
		{
			label: "Forward agent",
			min: 0, max: 1, value: fa, isBool: true,
			extra: func(v int) string {
				if v == 1 { return "SSH key forwarded to remote" }
				return "local SSH key not forwarded"
			},
		},
	}
}

// ── Model ─────────────────────────────────────────────────────────────────────

// EditResult is returned from RunEdit.
type EditResult struct {
	Saved               bool
	Deleted             bool   // true → caller should call profile.Remove()
	NewName             string // non-empty and differs from old → rename
	ForwardAgent        bool
	ServerAliveInterval int
	ServerAliveCountMax int
	ConnectionAttempts  int
}

type editModel struct {
	profileName   string
	cursor        int
	opts          []editOpt
	newName       string // rename text buffer
	deleteConfirm bool   // first Enter on delete row = show warning
	isDeleted     bool   // second Enter on delete row = confirmed
	done          bool
	aborted       bool
	errMsg        string
}

func newEditModel(p profile.Profile) editModel {
	return editModel{
		profileName: p.Name,
		opts:        defaultOpts(p),
	}
}

func (m editModel) Init() tea.Cmd { return nil }

func (m editModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "ctrl+c", "q":
		m.aborted = true
		return m, tea.Quit

	case "esc":
		m.aborted = true
		return m, tea.Quit

	case "enter":
		switch m.cursor {
		case optDelete:
			if m.deleteConfirm {
				m.isDeleted = true
				m.done = true
				return m, tea.Quit
			}
			m.deleteConfirm = true
		default:
			m.done = true
			return m, tea.Quit
		}

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.deleteConfirm = false
		}

	case "down", "j":
		if m.cursor < numOpts-1 {
			m.cursor++
			m.deleteConfirm = false
		}

	case "left", "h":
		if m.cursor < numSliderOpts {
			opt := &m.opts[m.cursor]
			if opt.value > opt.min {
				opt.value--
			}
		}

	case "right", "l":
		if m.cursor < numSliderOpts {
			opt := &m.opts[m.cursor]
			if opt.value < opt.max {
				opt.value++
			}
		}

	case "shift+left":
		if m.cursor < numSliderOpts {
			opt := &m.opts[m.cursor]
			step := largeStep(opt)
			opt.value -= step
			if opt.value < opt.min { opt.value = opt.min }
		}

	case "shift+right":
		if m.cursor < numSliderOpts {
			opt := &m.opts[m.cursor]
			step := largeStep(opt)
			opt.value += step
			if opt.value > opt.max { opt.value = opt.max }
		}

	case "backspace":
		if m.cursor == optRename && len(m.newName) > 0 {
			runes := []rune(m.newName)
			m.newName = string(runes[:len(runes)-1])
		}

	default:
		if m.cursor == optRename {
			runes := []rune(key.String())
			if len(runes) == 1 && runes[0] >= 32 {
				m.newName += string(runes)
			}
		}
	}

	return m, nil
}

func largeStep(opt *editOpt) int {
	span := opt.max - opt.min
	if span <= 10 { return 1 }
	if span <= 30 { return 5 }
	return 10
}

func (m editModel) View() string {
	var sb strings.Builder

	sb.WriteString(editTitle.Render("sshtie edit") + "  ")
	sb.WriteString(editSub.Render(m.profileName) + "\n\n")

	// Controls hint — changes when on rename row
	if m.cursor == optRename {
		sb.WriteString(editHint.Render("  type new name  ·  backspace  delete char  ·  enter  save  ·  esc  cancel") + "\n\n")
	} else {
		sb.WriteString(editHint.Render("  ↑/↓  select  ·  ←/→  adjust  ·  shift+←/→  jump  ·  enter  save  ·  esc  cancel") + "\n\n")
	}

	// ── Slider / toggle rows (0-3) ──
	for i, opt := range m.opts {
		active := i == m.cursor

		if active {
			sb.WriteString(editActive.Render("  ▶ "))
		} else {
			sb.WriteString(editInactive.Render("    "))
		}

		label := fmt.Sprintf("%-22s", opt.label)
		if active {
			sb.WriteString(editActive.Render(label))
		} else {
			sb.WriteString(editInactive.Render(label))
		}

		if opt.isBool {
			if opt.value == 1 {
				sb.WriteString(editBar.Render("● on "))
				sb.WriteString(editBarEmpty.Render("○ off"))
			} else {
				sb.WriteString(editBarEmpty.Render("○ on "))
				sb.WriteString(editBar.Render("● off"))
			}
		} else {
			filled := int(float64(opt.value-opt.min) / float64(opt.max-opt.min) * float64(barWidth))
			sb.WriteString(editBar.Render(strings.Repeat("━", filled)))
			sb.WriteString(editBarEmpty.Render(strings.Repeat("░", barWidth-filled)))

			valStr := fmt.Sprintf("  %3d%s", opt.value, opt.unit)
			sb.WriteString(editVal.Render(valStr))

			sb.WriteString(editMeta.Render(fmt.Sprintf("  (%d–%d)", opt.min, opt.max)))
		}

		if opt.extra != nil {
			sb.WriteString(editMeta.Render("  · " + opt.extra(opt.value)))
		}

		sb.WriteString("\n")
	}

	// Summary line for SSH timing
	sb.WriteString("\n")
	aliveInterval := m.opts[1].value
	aliveCount    := m.opts[2].value
	silenceSec    := aliveInterval * aliveCount
	sb.WriteString(editMeta.Render(fmt.Sprintf(
		"  Effective max silence: %ds (%dm %02ds)\n",
		silenceSec, silenceSec/60, silenceSec%60,
	)))

	// ── Profile management section ──
	sb.WriteString("\n")
	sb.WriteString(editMeta.Render("  ─── Profile ────────────────────────────────────────────") + "\n")

	// Row 4: Rename
	if m.cursor == optRename {
		sb.WriteString(editActive.Render("  ▶ "))
		sb.WriteString(editActive.Render(fmt.Sprintf("%-22s", "Rename")))
		if m.newName == "" {
			sb.WriteString(editMeta.Render(m.profileName))
			sb.WriteString(editActive.Render("█"))
			sb.WriteString(editMeta.Render("  (type to rename, empty = keep current)"))
		} else {
			sb.WriteString(editVal.Render(m.newName + "█"))
		}
	} else {
		sb.WriteString(editInactive.Render("    "))
		sb.WriteString(editInactive.Render(fmt.Sprintf("%-22s", "Rename")))
		if m.newName != "" {
			sb.WriteString(editVal.Render(m.newName))
			sb.WriteString(editMeta.Render(fmt.Sprintf("  (was: %s)", m.profileName)))
		} else {
			sb.WriteString(editMeta.Render(m.profileName))
		}
	}
	sb.WriteString("\n")

	// Row 5: Delete
	if m.cursor == optDelete {
		sb.WriteString(editActive.Render("  ▶ "))
		sb.WriteString(editActive.Render(fmt.Sprintf("%-22s", "Delete profile")))
		if m.deleteConfirm {
			sb.WriteString(editWarn.Render("⚠ press enter again to confirm"))
		} else {
			sb.WriteString(editMeta.Render("press enter to delete"))
		}
	} else {
		sb.WriteString(editInactive.Render("    "))
		sb.WriteString(editInactive.Render(fmt.Sprintf("%-22s", "Delete profile")))
	}
	sb.WriteString("\n")

	if m.errMsg != "" {
		sb.WriteString("\n")
		sb.WriteString(editErr.Render("  ✗ "+m.errMsg) + "\n")
	}

	return sb.String()
}

// ── Public API ────────────────────────────────────────────────────────────────

// RunEdit launches the advanced-options editor for the given profile.
// Returns EditResult with Saved=false if the user cancels.
func RunEdit(p profile.Profile) (EditResult, error) {
	prog := tea.NewProgram(newEditModel(p), tea.WithAltScreen())
	final, err := prog.Run()
	if err != nil {
		return EditResult{}, err
	}
	m := final.(editModel)
	if m.aborted || !m.done {
		return EditResult{Saved: false}, nil
	}

	newName := strings.TrimSpace(m.newName)
	if newName == m.profileName {
		newName = "" // no actual rename
	}

	return EditResult{
		Saved:               true,
		Deleted:             m.isDeleted,
		NewName:             newName,
		ForwardAgent:        m.opts[3].value == 1,
		ConnectionAttempts:  m.opts[0].value,
		ServerAliveInterval: m.opts[1].value,
		ServerAliveCountMax: m.opts[2].value,
	}, nil
}

// SummaryLine returns a one-line summary of the advanced settings for display.
func SummaryLine(p profile.Profile) string {
	aliveInterval := p.ServerAliveInterval
	if aliveInterval <= 0 { aliveInterval = 10 }
	aliveCount := p.ServerAliveCountMax
	if aliveCount <= 0 { aliveCount = 60 }
	attempts := p.ConnectionAttempts
	if attempts <= 0 { attempts = 3 }

	fa := "off"
	if p.ForwardAgent { fa = "on" }

	silenceSec := aliveInterval * aliveCount
	return fmt.Sprintf(
		"attempts=%d  alive=%ds×%d(%dm)  forward-agent=%s",
		attempts, aliveInterval, aliveCount, silenceSec/60, fa,
	)
}

// suppress unused import warning for editSaved
var _ = editSaved

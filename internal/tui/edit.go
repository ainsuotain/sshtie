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
)

const barWidth = 20

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
				// effective silence = interval × count, but we don't know interval here
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
	Saved           bool
	ForwardAgent    bool
	ServerAliveInterval int
	ServerAliveCountMax int
	ConnectionAttempts  int
}

type editModel struct {
	profileName string
	cursor      int
	opts        []editOpt
	done        bool
	aborted     bool
	errMsg      string
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

	opt := &m.opts[m.cursor]

	switch key.String() {
	case "ctrl+c", "q":
		m.aborted = true
		return m, tea.Quit

	case "esc":
		m.aborted = true
		return m, tea.Quit

	case "enter":
		m.done = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.opts)-1 {
			m.cursor++
		}

	case "left", "h":
		if opt.value > opt.min {
			opt.value--
		}

	case "right", "l":
		if opt.value < opt.max {
			opt.value++
		}

	// Fine-grained: shift+arrow = jump by larger step
	case "shift+left":
		step := largeStep(opt)
		opt.value -= step
		if opt.value < opt.min { opt.value = opt.min }

	case "shift+right":
		step := largeStep(opt)
		opt.value += step
		if opt.value > opt.max { opt.value = opt.max }
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
	sb.WriteString(editHint.Render("  ↑/↓  select  ·  ←/→  adjust  ·  shift+←/→  jump  ·  enter  save  ·  esc  cancel") + "\n\n")

	for i, opt := range m.opts {
		active := i == m.cursor

		// Row prefix
		if active {
			sb.WriteString(editActive.Render("  ▶ "))
		} else {
			sb.WriteString(editInactive.Render("    "))
		}

		// Label (fixed width)
		label := fmt.Sprintf("%-22s", opt.label)
		if active {
			sb.WriteString(editActive.Render(label))
		} else {
			sb.WriteString(editInactive.Render(label))
		}

		if opt.isBool {
			// Toggle display
			if opt.value == 1 {
				sb.WriteString(editBar.Render("● on "))
				sb.WriteString(editBarEmpty.Render("○ off"))
			} else {
				sb.WriteString(editBarEmpty.Render("○ on "))
				sb.WriteString(editBar.Render("● off"))
			}
		} else {
			// Slider bar
			filled := int(float64(opt.value-opt.min) / float64(opt.max-opt.min) * float64(barWidth))
			sb.WriteString(editBar.Render(strings.Repeat("━", filled)))
			sb.WriteString(editBarEmpty.Render(strings.Repeat("░", barWidth-filled)))

			// Value label
			valStr := fmt.Sprintf("  %3d%s", opt.value, opt.unit)
			sb.WriteString(editVal.Render(valStr))

			// Range hint
			sb.WriteString(editMeta.Render(fmt.Sprintf("  (%d–%d)", opt.min, opt.max)))
		}

		// Right-side annotation
		if opt.extra != nil {
			sb.WriteString(editMeta.Render("  · " + opt.extra(opt.value)))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	if m.errMsg != "" {
		sb.WriteString(editErr.Render("  ✗ "+m.errMsg) + "\n")
	}

	// Summary line
	aliveInterval := m.opts[1].value
	aliveCount    := m.opts[2].value
	silenceSec    := aliveInterval * aliveCount
	sb.WriteString(editMeta.Render(fmt.Sprintf(
		"  Effective max silence: %ds (%dm %02ds)\n",
		silenceSec, silenceSec/60, silenceSec%60,
	)))

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

	return EditResult{
		Saved:               true,
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

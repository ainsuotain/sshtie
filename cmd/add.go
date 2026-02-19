package cmd

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	wizTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	wizProgress = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	wizDone     = lipgloss.NewStyle().Foreground(lipgloss.Color("72"))
	wizActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	wizPending  = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	wizInput    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	wizHint     = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	wizErr      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	wizSelOn    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	wizSelOff   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

// ── Step definitions ──────────────────────────────────────────────────────────

type stepDef struct {
	label    string
	defVal   string // shown as hint; used when user presses Enter on empty
	required bool
	hint     string // contextual description shown in gray below the input
}
var steps = []stepDef{
	{
		label:    "Profile name",
		required: true,
		hint:     "A nickname for this connection (e.g., macmini, work-server)",
	},
	{
		label:    "Host",
		required: true,
		hint:     "Server IP or domain (e.g., 192.168.1.10, server.com)",
	},
	{
		label:    "User",
		required: true,
		hint:     "SSH login username (e.g., david, ubuntu, root)",
	},
	{
		label:  "Port",
		defVal: "22",
		hint:   "SSH port, usually 22 (press Enter if unsure)",
	},
	{
		label:  "SSH Key",
		defVal: "~/.ssh/id_ed25519",
		hint:   "Path to private key (press Enter if unsure)",
	},
	{
		label:  "tmux session",
		defVal: "main",
		hint:   "Session name (press Enter if unsure)",
	},
	{
		label:  "Network mode",
		defVal: "auto",
		hint:   "Recommended: auto — auto-detects Tailscale/mosh",
	},
}

const netStep = 6 // index of the network selection step

var networkOptions = []string{"auto", "tailscale", "direct"}

// ── Model ─────────────────────────────────────────────────────────────────────

type addWizard struct {
	step    int
	values  []string // resolved value per step (empty = not yet filled)
	buf     string   // live text buffer for the current text step
	netSel  int      // cursor for network selection
	errMsg  string
	done    bool
	aborted bool
}

func newAddWizard() addWizard {
	return addWizard{values: make([]string, len(steps))}
}

func (m addWizard) Init() tea.Cmd { return nil }

func (m addWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {

	case "ctrl+c":
		m.aborted = true
		return m, tea.Quit

	case "esc":
		if m.step > 0 {
			m.step--
			m.buf = m.values[m.step]
			// Restore network cursor when stepping back into netStep.
			if m.step == netStep {
				for i, o := range networkOptions {
					if o == m.values[netStep] {
						m.netSel = i
						break
					}
				}
			}
			m.errMsg = ""
		}

	case "enter":
		if m.step == netStep {
			m.values[netStep] = networkOptions[m.netSel]
			m.done = true
			return m, tea.Quit
		}
		// Validate text input.
		val := strings.TrimSpace(m.buf)
		s := steps[m.step]
		if val == "" {
			if s.required {
				m.errMsg = fmt.Sprintf("%s is required", s.label)
				return m, nil
			}
			val = s.defVal
		}
		// Port range check.
		if m.step == 3 {
			p, err := strconv.Atoi(val)
			if err != nil || p < 1 || p > 65535 {
				m.errMsg = "Port must be a number between 1 and 65535"
				return m, nil
			}
		}
		m.values[m.step] = val
		m.errMsg = ""
		m.step++
		m.buf = m.values[m.step] // restore previously entered value (if any)

	case "backspace":
		if m.step < netStep && len(m.buf) > 0 {
			runes := []rune(m.buf)
			m.buf = string(runes[:len(runes)-1])
		}

	default:
		if m.step == netStep {
			switch keyMsg.String() {
			case "up", "k":
				if m.netSel > 0 {
					m.netSel--
				}
			case "down", "j":
				if m.netSel < len(networkOptions)-1 {
					m.netSel++
				}
			}
		} else {
			// Append printable single characters (including Unicode).
			runes := []rune(keyMsg.String())
			if len(runes) == 1 && runes[0] >= 32 {
				m.buf += string(runes)
			}
		}
	}

	return m, nil
}

func (m addWizard) View() string {
	var sb strings.Builder
	total := len(steps)

	sb.WriteString(wizTitle.Render("sshtie add") + "  New Profile\n")
	sb.WriteString(wizProgress.Render(fmt.Sprintf("  Step %d / %d\n\n", m.step+1, total)))

	for i, s := range steps {
		switch {
		case i < m.step:
			// ── Completed ────────────────────────────────────────────────
			val := m.values[i]
			sb.WriteString(wizDone.Render(fmt.Sprintf("  ✓ %-16s", s.label)))
			sb.WriteString(wizInput.Render(val) + "\n")

		case i == m.step:
			// ── Active ───────────────────────────────────────────────────
			sb.WriteString(wizActive.Render(fmt.Sprintf("  ▶ %-16s", s.label)))

			if i == netStep {
				sb.WriteString("\n")
				for j, opt := range networkOptions {
					if j == m.netSel {
						sb.WriteString(wizSelOn.Render("    ● "+opt) + "\n")
					} else {
						sb.WriteString(wizSelOff.Render("    ○ "+opt) + "\n")
					}
				}
			} else {
				// Text cursor
				sb.WriteString(wizInput.Render(m.buf + "█"))
				if s.defVal != "" && m.buf == "" {
					sb.WriteString(wizHint.Render(
						fmt.Sprintf("  (Enter to use default: %s)", s.defVal)))
				} else if s.required && m.buf == "" {
					sb.WriteString(wizHint.Render("  (required)"))
				}
				sb.WriteString("\n")
			}

			// Contextual hint shown below the input / selection.
			if s.hint != "" {
				sb.WriteString(wizHint.Render("    "+s.hint) + "\n")
			}

			if m.errMsg != "" {
				sb.WriteString(wizErr.Render("    ✗ "+m.errMsg) + "\n")
			}

		default:
			// ── Pending ──────────────────────────────────────────────────
			sb.WriteString(wizPending.Render(fmt.Sprintf("  · %-16s", s.label)))
			if s.defVal != "" {
				sb.WriteString(wizHint.Render(s.defVal))
			} else if s.hint != "" {
				// Show a short preview of the hint for required fields.
				sb.WriteString(wizHint.Render(s.hint))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	if m.step == netStep {
		sb.WriteString(wizHint.Render("  ↑/↓  k/j  select  •  enter  confirm  •  esc  back"))
	} else {
		sb.WriteString(wizHint.Render("  enter  next  •  esc  back  •  ctrl+c  cancel"))
	}
	sb.WriteString("\n")
	return sb.String()
}

// ── Cobra command ─────────────────────────────────────────────────────────────

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new profile (interactive TUI wizard)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		prog := tea.NewProgram(newAddWizard(), tea.WithAltScreen())
		final, err := prog.Run()
		if err != nil {
			return err
		}

		wiz := final.(addWizard)
		if wiz.aborted || !wiz.done {
			fmt.Println("→ Cancelled.")
			return nil
		}

		// Parse port (already validated in wizard).
		port, _ := strconv.Atoi(wiz.values[3])
		if port <= 0 {
			port = 22
		}

		// Store key as empty string when it matches the default
		// so profile.DefaultKey() can manage it.
		key := wiz.values[4]
		if key == "~/.ssh/id_ed25519" {
			key = ""
		}

		p := profile.Profile{
			Name:        wiz.values[0],
			Host:        wiz.values[1],
			User:        wiz.values[2],
			Port:        port,
			Key:         key,
			TmuxSession: wiz.values[5],
			Network:     wiz.values[6],
		}

		if err := profile.Add(p); err != nil {
			return err
		}

		fmt.Printf("✅ Profile '%s' saved!\n", p.Name)
		fmt.Printf("→ Try: sshtie connect %s\n", p.Name)
		return nil
	},
}

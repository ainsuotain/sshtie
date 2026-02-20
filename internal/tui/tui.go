package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ainsuotain/sshtie/internal/profile"
)

// Action indicates what the user selected.
type Action int

const (
	ActionNone    Action = iota
	ActionConnect        // Enter
	ActionDoctor         // d
	ActionEdit           // e
	ActionQuit           // q / ESC / ctrl+c
)

// Result is returned by Run after the TUI exits.
type Result struct {
	Action  Action
	Profile profile.Profile
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Background(lipgloss.Color("237"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type model struct {
	profiles []profile.Profile
	cursor   int
	action   Action
	chosen   profile.Profile
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.action = ActionQuit
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.profiles) > 0 {
				m.action = ActionConnect
				m.chosen = m.profiles[m.cursor]
				return m, tea.Quit
			}
		case "d":
			if len(m.profiles) > 0 {
				m.action = ActionDoctor
				m.chosen = m.profiles[m.cursor]
				return m, tea.Quit
			}
		case "e":
			if len(m.profiles) > 0 {
				m.action = ActionEdit
				m.chosen = m.profiles[m.cursor]
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("sshtie") + "  SSH + mosh + tmux, unified\n\n")

	if len(m.profiles) == 0 {
		sb.WriteString(errorStyle.Render("  No profiles yet.") + "\n")
		sb.WriteString(dimStyle.Render("  Run: sshtie add") + "\n")
	} else {
		for i, p := range m.profiles {
			port := p.Port
			if port == 0 {
				port = 22
			}
			net := p.Network
			if net == "" {
				net = "auto"
			}
			addr := fmt.Sprintf("%s@%s:%d", p.User, p.Host, port)
			row := fmt.Sprintf("%-18s  %-30s  [%s]", p.Name, addr, net)
			if i == m.cursor {
				sb.WriteString(selectedStyle.Render("▶ "+row) + "\n")
			} else {
				sb.WriteString(normalStyle.Render("  "+row) + "\n")
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  ↑/↓  k/j  navigate  •  enter  connect  •  d  doctor  •  e  edit  •  q  quit"))
	sb.WriteString("\n")

	return sb.String()
}

// Run launches the interactive TUI and returns the user's choice.
// The caller must act on Result AFTER this function returns so that the
// terminal is fully restored before ssh/mosh takes over.
func Run(profiles []profile.Profile) (Result, error) {
	m := model{profiles: profiles}
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return Result{}, err
	}
	fm := final.(model)
	return Result{Action: fm.action, Profile: fm.chosen}, nil
}

package tui

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/tailscale"
)

// ConnectAction is what the user chose after seeing the check results.
type ConnectAction int

const (
	ConnectNone    ConnectAction = iota
	ConnectProceed               // enter — go ahead and connect
	ConnectInstall               // i     — install missing tools first
	ConnectQuit                  // q/esc — abort
)

// ConnectResult is returned by RunConnect after the TUI exits.
type ConnectResult struct {
	Action ConnectAction
}

// ── check indices ─────────────────────────────────────────────────────────────

const (
	idxSSH       = 0
	idxTmux      = 1
	idxMosh      = 2
	idxTailscale = 3
	numChecks    = 4
)

// ── check state ───────────────────────────────────────────────────────────────

type cState int

const (
	cChecking cState = iota
	cOK
	cFail
	cSkip
)

type checkItem struct {
	label  string
	state  cState
	detail string
	hint   string // beginner-friendly explanation shown in guide panel
}

// ── messages ──────────────────────────────────────────────────────────────────

type checkDoneMsg struct {
	idx    int
	state  cState
	detail string
	hint   string
}

// remoteDepsMsg carries mosh + tmux results from a single SSH call.
type remoteDepsMsg struct {
	moshState  cState
	moshDetail string
	moshHint   string
	tmuxState  cState
	tmuxDetail string
	tmuxHint   string
}

type cSpinTickMsg struct{}

// ── styles ────────────────────────────────────────────────────────────────────

var (
	cOKStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	cWarnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	cSkipStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	cSpinStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	cKeyStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	cSubStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	cHintBox   = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Padding(0, 2).
			Foreground(lipgloss.Color("252"))
	cStrategyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cNewHostStyle  = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 2).
			Foreground(lipgloss.Color("252"))
)

var cSpinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ── model ─────────────────────────────────────────────────────────────────────

type connectModel struct {
	prof         profile.Profile
	checks       [numChecks]checkItem
	frame        int
	allDone      bool
	action       ConnectAction
	strategy     string
	needsInstall bool
	isNewHost    bool
}

func newConnectModel(p profile.Profile, isNewHost bool) connectModel {
	port := p.Port
	if port == 0 {
		port = 22
	}
	m := connectModel{prof: p, isNewHost: isNewHost}
	m.checks[idxSSH] = checkItem{label: fmt.Sprintf("SSH  (port %d)", port)}
	m.checks[idxTmux] = checkItem{label: "tmux         (server)"}
	m.checks[idxMosh] = checkItem{label: "mosh-server  (server)"}
	m.checks[idxTailscale] = checkItem{label: "Tailscale    (local)"}
	for i := range m.checks {
		m.checks[i].state = cChecking
	}
	return m
}

func (m connectModel) Init() tea.Cmd {
	return tea.Batch(
		cSpinTick(),
		cmdCheckSSH(m.prof),
		cmdCheckTailscale(),
		// remote deps fired later, after SSH check passes
	)
}

func (m connectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Discard key presses while checks are still running (except quit).
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.action = ConnectQuit
			return m, tea.Quit
		}
		if !m.allDone {
			return m, nil
		}
		switch msg.String() {
		case "enter":
			if m.checks[idxSSH].state == cOK {
				m.action = ConnectProceed
				return m, tea.Quit
			}
		case "i":
			if m.needsInstall {
				m.action = ConnectInstall
				return m, tea.Quit
			}
		}

	case checkDoneMsg:
		m.checks[msg.idx].state = msg.state
		m.checks[msg.idx].detail = msg.detail
		m.checks[msg.idx].hint = msg.hint

		var next tea.Cmd
		if msg.idx == idxSSH {
			if msg.state == cOK {
				// SSH reachable — now check remote tools.
				next = cmdCheckRemoteDeps(m.prof)
			} else {
				// SSH failed — skip remote checks immediately.
				m.checks[idxMosh] = checkItem{
					label:  m.checks[idxMosh].label,
					state:  cSkip,
					detail: cSkipStyle.Render("skipped"),
				}
				m.checks[idxTmux] = checkItem{
					label:  m.checks[idxTmux].label,
					state:  cSkip,
					detail: cSkipStyle.Render("skipped"),
				}
			}
		}

		m.allDone = m.isDone()
		if m.allDone {
			m.strategy = m.calcStrategy()
			m.needsInstall = m.checks[idxMosh].state == cFail ||
				m.checks[idxTmux].state == cFail
		}
		return m, next

	case remoteDepsMsg:
		m.checks[idxMosh].state = msg.moshState
		m.checks[idxMosh].detail = msg.moshDetail
		m.checks[idxMosh].hint = msg.moshHint
		m.checks[idxTmux].state = msg.tmuxState
		m.checks[idxTmux].detail = msg.tmuxDetail
		m.checks[idxTmux].hint = msg.tmuxHint

		m.allDone = m.isDone()
		if m.allDone {
			m.strategy = m.calcStrategy()
			m.needsInstall = m.checks[idxMosh].state == cFail ||
				m.checks[idxTmux].state == cFail
		}

	case cSpinTickMsg:
		m.frame = (m.frame + 1) % len(cSpinFrames)
		if !m.allDone {
			return m, cSpinTick()
		}
	}
	return m, nil
}

func (m connectModel) isDone() bool {
	for _, c := range m.checks {
		if c.state == cChecking {
			return false
		}
	}
	return true
}

func (m connectModel) calcStrategy() string {
	moshOK := m.checks[idxMosh].state == cOK
	tmuxOK := m.checks[idxTmux].state == cOK
	switch {
	case moshOK && tmuxOK:
		return cOKStyle.Render("mosh + tmux") + cSubStyle.Render("  ✨ reconnects automatically if network drops")
	case tmuxOK:
		return cStrategyStyle.Render("ssh + tmux") + cSubStyle.Render("  (session stays alive, no auto-reconnect)")
	default:
		return cWarnStyle.Render("ssh only") + cSubStyle.Render("  (basic, no persistent session)")
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m connectModel) View() string {
	var b strings.Builder

	port := m.prof.Port
	if port == 0 {
		port = 22
	}

	// ── header ──
	b.WriteString("\n")
	b.WriteString("  " + titleStyle.Render("sshtie") + "  →  " + titleStyle.Render(m.prof.Name) + "\n")
	b.WriteString(cSubStyle.Render(fmt.Sprintf("  %s@%s · port %d", m.prof.User, m.prof.Host, port)) + "\n\n")
	b.WriteString(cSubStyle.Render("  ───────────────────────────────────────────") + "\n")

	// ── checks ──
	for _, c := range m.checks {
		icon := ""
		detail := c.detail
		switch c.state {
		case cChecking:
			icon = cSpinStyle.Render(cSpinFrames[m.frame])
			detail = cSubStyle.Render("checking…")
		case cOK:
			icon = cOKStyle.Render("✓")
		case cFail:
			icon = cWarnStyle.Render("⚠")
		case cSkip:
			icon = cSkipStyle.Render("─")
		}
		b.WriteString(fmt.Sprintf("  %-28s %s  %s\n",
			cSubStyle.Render(c.label), icon, detail))
	}

	b.WriteString(cSubStyle.Render("  ───────────────────────────────────────────") + "\n\n")

	// ── while still checking ──
	if !m.allDone {
		b.WriteString(cSubStyle.Render("  running checks…   ") +
			cKeyStyle.Render("[q]") + cSubStyle.Render(" cancel") + "\n\n")
		return b.String()
	}

	// ── strategy ──
	b.WriteString("  " + cSubStyle.Render("Strategy: ") + m.strategy + "\n\n")

	// ── new host notice ──
	if m.isNewHost {
		notice := "First time connecting to this server.\n" +
			"SSH will verify and save the server's fingerprint —\n" +
			"this keeps you safe from impersonation on future connections."
		b.WriteString("  " + cNewHostStyle.Render("🔐 Security notice\n\n"+notice) + "\n\n")
	}

	// ── guide: hints for warnings ──
	hints := m.collectHints()
	if len(hints) > 0 {
		hintText := "💡 Guide\n\n" + strings.Join(hints, "\n\n")
		b.WriteString("  " + cHintBox.Render(hintText) + "\n\n")
	}

	// ── action bar ──
	sshOK := m.checks[idxSSH].state == cOK
	if sshOK {
		b.WriteString("  " + cKeyStyle.Render("[enter]") + cSubStyle.Render(" connect"))
		if m.needsInstall {
			b.WriteString("   " + cKeyStyle.Render("[i]") + cSubStyle.Render(" install missing tools"))
		}
	} else {
		b.WriteString("  " + cWarnStyle.Render("Cannot connect — fix the issue above first."))
	}
	b.WriteString("   " + cKeyStyle.Render("[q]") + cSubStyle.Render(" quit") + "\n\n")

	return b.String()
}

func (m connectModel) collectHints() []string {
	var out []string
	for _, c := range m.checks {
		if c.state == cFail && c.hint != "" {
			out = append(out, c.hint)
		}
	}
	return out
}

// ── commands (async checks) ───────────────────────────────────────────────────

func cSpinTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return cSpinTickMsg{}
	})
}

func cmdCheckSSH(p profile.Profile) tea.Cmd {
	return func() tea.Msg {
		port := p.Port
		if port == 0 {
			port = 22
		}
		addr := net.JoinHostPort(p.Host, strconv.Itoa(port))
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			return checkDoneMsg{
				idx:    idxSSH,
				state:  cFail,
				detail: cWarnStyle.Render("unreachable"),
				hint: fmt.Sprintf(
					"SSH couldn't reach %s on port %d.\n"+
						"  Things to check:\n"+
						"  • Is the server online?\n"+
						"  • Is port %d open in the firewall?\n"+
						"  • Is your SSH key authorised?  ssh-copy-id %s@%s",
					p.Host, port, port, p.User, p.Host),
			}
		}
		conn.Close()
		return checkDoneMsg{
			idx:    idxSSH,
			state:  cOK,
			detail: cOKStyle.Render("reachable"),
		}
	}
}

func cmdCheckTailscale() tea.Cmd {
	return func() tea.Msg {
		if tailscale.ClientRunning() {
			return checkDoneMsg{
				idx:    idxTailscale,
				state:  cOK,
				detail: cOKStyle.Render("running"),
			}
		}
		return checkDoneMsg{
			idx:    idxTailscale,
			state:  cSkip,
			detail: cSkipStyle.Render("not running  (optional)"),
			// No hint — Tailscale is entirely optional.
		}
	}
}

func cmdCheckRemoteDeps(p profile.Profile) tea.Cmd {
	return func() tea.Msg {
		port := p.Port
		if port == 0 {
			port = 22
		}
		args := []string{
			"-p", strconv.Itoa(port),
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "ConnectTimeout=5",
			"-o", "BatchMode=yes",
		}
		key := p.DefaultKey()
		if _, err := os.Stat(key); err == nil {
			args = append(args, "-i", key)
		}
		args = append(args, fmt.Sprintf("%s@%s", p.User, p.Host))
		args = append(args, "which mosh-server >/dev/null 2>&1 && echo M; which tmux >/dev/null 2>&1 && echo T")

		out, err := exec.Command("ssh", args...).Output()
		if err != nil {
			// SSH needs a password, or another transient error — skip gracefully.
			return remoteDepsMsg{
				moshState:  cSkip,
				moshDetail: cSkipStyle.Render("couldn't verify  (password auth?)"),
				tmuxState:  cSkip,
				tmuxDetail: cSkipStyle.Render("couldn't verify  (password auth?)"),
			}
		}

		s := string(out)
		hasMosh := strings.Contains(s, "M")
		hasTmux := strings.Contains(s, "T")

		msg := remoteDepsMsg{}

		if hasMosh {
			msg.moshState = cOK
			msg.moshDetail = cOKStyle.Render("installed")
		} else {
			msg.moshState = cFail
			msg.moshDetail = cWarnStyle.Render("not installed")
			msg.moshHint = fmt.Sprintf(
				"mosh-server is not on the server.\n"+
					"  mosh keeps your session alive even if your WiFi drops\n"+
					"  or you close your laptop — you pick up right where you left off.\n"+
					"  Install it with:  sshtie install %s", p.Name)
		}

		if hasTmux {
			msg.tmuxState = cOK
			msg.tmuxDetail = cOKStyle.Render("installed")
		} else {
			msg.tmuxState = cFail
			msg.tmuxDetail = cWarnStyle.Render("not installed")
			msg.tmuxHint = fmt.Sprintf(
				"tmux is not on the server.\n"+
					"  tmux keeps your work running on the server even after you disconnect —\n"+
					"  log back in any time and resume exactly where you stopped.\n"+
					"  Install it with:  sshtie install %s", p.Name)
		}

		return msg
	}
}

// ── public entry point ────────────────────────────────────────────────────────

// RunConnect launches the connection-progress TUI and returns what the user chose.
// The caller must act on the result AFTER this returns so the terminal is restored.
func RunConnect(p profile.Profile, isNewHost bool) (ConnectResult, error) {
	m := newConnectModel(p, isNewHost)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	final, err := prog.Run()
	if err != nil {
		return ConnectResult{}, err
	}
	fm := final.(connectModel)
	return ConnectResult{Action: fm.action}, nil
}

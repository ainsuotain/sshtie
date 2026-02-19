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

	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/tailscale"
)

// DoctorAction is what the user chose after the diagnostic screen.
type DoctorAction int

const (
	DoctorNone    DoctorAction = iota
	DoctorConnect              // c / enter
	DoctorInstall              // i
	DoctorQuit                 // q / esc
)

// DoctorResult is returned by RunDoctor after the TUI exits.
type DoctorResult struct {
	Action DoctorAction
}

// ── check indices ─────────────────────────────────────────────────────────────

const (
	dSSH      = 0
	dTmux     = 1
	dMosh     = 2
	dUDP      = 3
	dTSClient = 4
	dTSServer = 5
	dTotal    = 6
)

// ── messages ──────────────────────────────────────────────────────────────────

type dSingleMsg struct {
	idx    int
	state  cState
	detail string
	hint   string
}

type dRemoteMsg struct {
	tmuxState  cState
	tmuxDetail string
	tmuxHint   string
	moshState  cState
	moshDetail string
	moshHint   string
}

type dTSServerMsg struct {
	state  cState
	detail string
}

type dTickMsg struct{}

// ── model ─────────────────────────────────────────────────────────────────────

type doctorModel struct {
	prof         profile.Profile
	checks       [dTotal]checkItem
	frame        int
	allDone      bool
	action       DoctorAction
	strategy     string
	needsInstall bool
}

func newDoctorModel(p profile.Profile) doctorModel {
	port := p.Port
	if port == 0 {
		port = 22
	}
	m := doctorModel{prof: p}
	m.checks[dSSH] = checkItem{label: fmt.Sprintf("SSH          (port %d)", port)}
	m.checks[dTmux] = checkItem{label: "tmux         (server)"}
	m.checks[dMosh] = checkItem{label: "mosh-server  (server)"}
	m.checks[dUDP] = checkItem{label: "UDP 60001    (firewall)"}
	m.checks[dTSClient] = checkItem{label: "Tailscale    (local)"}
	m.checks[dTSServer] = checkItem{label: "Tailscale    (server)"}
	for i := range m.checks {
		m.checks[i].state = cChecking
	}
	return m
}

func (m doctorModel) Init() tea.Cmd {
	return tea.Batch(
		dTick(),
		cmdDoctorSSH(m.prof),
		cmdDoctorUDP(m.prof.Host),
		cmdDoctorTSClient(),
		// remote deps: fired after SSH passes
		// TS server:   fired after TS client passes
	)
}

func (m doctorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.action = DoctorQuit
			return m, tea.Quit
		}
		if !m.allDone {
			return m, nil
		}
		switch msg.String() {
		case "enter", "c":
			if m.checks[dSSH].state == cOK {
				m.action = DoctorConnect
				return m, tea.Quit
			}
		case "i":
			if m.needsInstall {
				m.action = DoctorInstall
				return m, tea.Quit
			}
		}

	case dSingleMsg:
		m.checks[msg.idx].state = msg.state
		m.checks[msg.idx].detail = msg.detail
		m.checks[msg.idx].hint = msg.hint

		var next tea.Cmd
		switch msg.idx {
		case dSSH:
			if msg.state == cOK {
				next = cmdDoctorRemote(m.prof)
			} else {
				m.checks[dTmux] = checkItem{label: m.checks[dTmux].label, state: cSkip, detail: cSkipStyle.Render("skipped")}
				m.checks[dMosh] = checkItem{label: m.checks[dMosh].label, state: cSkip, detail: cSkipStyle.Render("skipped")}
			}
		case dTSClient:
			if msg.state == cOK {
				next = cmdDoctorTSServer(m.prof.Host)
			} else {
				m.checks[dTSServer] = checkItem{label: m.checks[dTSServer].label, state: cSkip, detail: cSkipStyle.Render("skipped")}
			}
		}

		m.allDone = m.dIsDone()
		if m.allDone {
			m.strategy = m.dStrategy()
			m.needsInstall = m.checks[dMosh].state == cFail || m.checks[dTmux].state == cFail
		}
		return m, next

	case dRemoteMsg:
		m.checks[dTmux].state = msg.tmuxState
		m.checks[dTmux].detail = msg.tmuxDetail
		m.checks[dTmux].hint = msg.tmuxHint
		m.checks[dMosh].state = msg.moshState
		m.checks[dMosh].detail = msg.moshDetail
		m.checks[dMosh].hint = msg.moshHint

		m.allDone = m.dIsDone()
		if m.allDone {
			m.strategy = m.dStrategy()
			m.needsInstall = m.checks[dMosh].state == cFail || m.checks[dTmux].state == cFail
		}

	case dTSServerMsg:
		m.checks[dTSServer].state = msg.state
		m.checks[dTSServer].detail = msg.detail

		m.allDone = m.dIsDone()
		if m.allDone {
			m.strategy = m.dStrategy()
			m.needsInstall = m.checks[dMosh].state == cFail || m.checks[dTmux].state == cFail
		}

	case dTickMsg:
		m.frame = (m.frame + 1) % len(cSpinFrames)
		if !m.allDone {
			return m, dTick()
		}
	}
	return m, nil
}

func (m doctorModel) dIsDone() bool {
	for _, c := range m.checks {
		if c.state == cChecking {
			return false
		}
	}
	return true
}

func (m doctorModel) dStrategy() string {
	moshOK := m.checks[dMosh].state == cOK
	tmuxOK := m.checks[dTmux].state == cOK
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

func (m doctorModel) View() string {
	var b strings.Builder

	port := m.prof.Port
	if port == 0 {
		port = 22
	}

	// header
	b.WriteString("\n")
	b.WriteString("  " + titleStyle.Render("sshtie doctor") + "  →  " + titleStyle.Render(m.prof.Name) + "\n")
	b.WriteString(cSubStyle.Render(fmt.Sprintf("  %s@%s · port %d", m.prof.User, m.prof.Host, port)) + "\n\n")
	b.WriteString(cSubStyle.Render("  ───────────────────────────────────────────") + "\n")

	// checks
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
		b.WriteString(fmt.Sprintf("  %-30s %s  %s\n",
			cSubStyle.Render(c.label), icon, detail))
	}

	b.WriteString(cSubStyle.Render("  ───────────────────────────────────────────") + "\n\n")

	// while still checking
	if !m.allDone {
		b.WriteString(cSubStyle.Render("  running diagnostics…   ") +
			cKeyStyle.Render("[q]") + cSubStyle.Render(" cancel") + "\n\n")
		return b.String()
	}

	// strategy
	b.WriteString("  " + cSubStyle.Render("Strategy: ") + m.strategy + "\n\n")

	// guide hints
	var hints []string
	for _, c := range m.checks {
		if c.state == cFail && c.hint != "" {
			hints = append(hints, c.hint)
		}
	}
	if len(hints) > 0 {
		b.WriteString("  " + cHintBox.Render("💡 Guide\n\n"+strings.Join(hints, "\n\n")) + "\n\n")
	}

	// action bar
	sshOK := m.checks[dSSH].state == cOK
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

// ── commands (async checks) ───────────────────────────────────────────────────

func dTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return dTickMsg{}
	})
}

func cmdDoctorSSH(p profile.Profile) tea.Cmd {
	return func() tea.Msg {
		port := p.Port
		if port == 0 {
			port = 22
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(p.Host, strconv.Itoa(port)), 5*time.Second)
		if err != nil {
			return dSingleMsg{
				idx:    dSSH,
				state:  cFail,
				detail: cWarnStyle.Render("unreachable"),
				hint: fmt.Sprintf(
					"SSH can't reach %s on port %d.\n"+
						"  • Is the server online?\n"+
						"  • Is port %d open in the firewall?\n"+
						"  • Correct host/port?  sshtie edit %s",
					p.Host, port, port, p.Name),
			}
		}
		conn.Close()
		return dSingleMsg{idx: dSSH, state: cOK, detail: cOKStyle.Render("reachable")}
	}
}

func cmdDoctorUDP(host string) tea.Cmd {
	return func() tea.Msg {
		addr := net.JoinHostPort(host, "60001")
		conn, err := net.DialTimeout("udp", addr, 3*time.Second)
		if err != nil {
			return dSingleMsg{
				idx:    dUDP,
				state:  cFail,
				detail: cWarnStyle.Render("blocked"),
				hint:   "UDP port 60001 is blocked — mosh needs ports 60000–61000/UDP.\n  On the server run:\n    sudo ufw allow 60000:61000/udp",
			}
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		conn.Write([]byte{})
		buf := make([]byte, 1)
		_, readErr := conn.Read(buf)
		if readErr != nil && strings.Contains(readErr.Error(), "connection refused") {
			return dSingleMsg{
				idx:    dUDP,
				state:  cFail,
				detail: cWarnStyle.Render("blocked"),
				hint:   "UDP port 60001 is blocked — mosh needs ports 60000–61000/UDP.\n  On the server run:\n    sudo ufw allow 60000:61000/udp",
			}
		}
		return dSingleMsg{idx: dUDP, state: cOK, detail: cOKStyle.Render("open")}
	}
}

func cmdDoctorTSClient() tea.Cmd {
	return func() tea.Msg {
		if tailscale.ClientRunning() {
			return dSingleMsg{idx: dTSClient, state: cOK, detail: cOKStyle.Render("running")}
		}
		if _, err := exec.LookPath("tailscale"); err != nil {
			return dSingleMsg{
				idx:    dTSClient,
				state:  cSkip,
				detail: cSkipStyle.Render("not installed  (optional)"),
			}
		}
		return dSingleMsg{
			idx:    dTSClient,
			state:  cFail,
			detail: cWarnStyle.Render("installed but not active"),
			hint:   "Tailscale is installed but not running.\n  Start it with:\n    tailscale up",
		}
	}
}

func cmdDoctorTSServer(host string) tea.Cmd {
	return func() tea.Msg {
		if tailscale.HostInNetwork(host) {
			return dTSServerMsg{state: cOK, detail: cOKStyle.Render("visible in Tailscale network")}
		}
		return dTSServerMsg{
			state:  cFail,
			detail: cWarnStyle.Render("not visible"),
		}
	}
}

func cmdDoctorRemote(p profile.Profile) tea.Cmd {
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
		// Single round-trip: get tmux version + check mosh-server
		args = append(args, "tmux -V 2>/dev/null; which mosh-server >/dev/null 2>&1 && echo MOSH_OK")

		out, err := exec.Command("ssh", args...).Output()
		if err != nil {
			return dRemoteMsg{
				tmuxState:  cSkip,
				tmuxDetail: cSkipStyle.Render("couldn't verify"),
				moshState:  cSkip,
				moshDetail: cSkipStyle.Render("couldn't verify"),
			}
		}

		s := string(out)
		msg := dRemoteMsg{}

		// tmux: output line looks like "tmux 3.3a"
		if strings.Contains(s, "tmux ") {
			for _, line := range strings.Split(s, "\n") {
				if strings.HasPrefix(line, "tmux ") {
					msg.tmuxState = cOK
					msg.tmuxDetail = cOKStyle.Render(strings.TrimSpace(line))
					break
				}
			}
		} else {
			msg.tmuxState = cFail
			msg.tmuxDetail = cWarnStyle.Render("not installed")
			msg.tmuxHint = fmt.Sprintf(
				"tmux is not on the server.\n"+
					"  tmux keeps your work running even after you disconnect.\n"+
					"  Install it with:  sshtie install %s", p.Name)
		}

		// mosh-server
		if strings.Contains(s, "MOSH_OK") {
			msg.moshState = cOK
			msg.moshDetail = cOKStyle.Render("installed")
		} else {
			msg.moshState = cFail
			msg.moshDetail = cWarnStyle.Render("not installed")
			msg.moshHint = fmt.Sprintf(
				"mosh-server is not on the server.\n"+
					"  mosh keeps your session alive through network drops.\n"+
					"  Install it with:  sshtie install %s", p.Name)
		}

		return msg
	}
}

// ── public entry point ────────────────────────────────────────────────────────

// RunDoctor launches the diagnostic TUI and returns what the user chose.
func RunDoctor(p profile.Profile) (DoctorResult, error) {
	m := newDoctorModel(p)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	final, err := prog.Run()
	if err != nil {
		return DoctorResult{}, err
	}
	fm := final.(doctorModel)
	return DoctorResult{Action: fm.action}, nil
}

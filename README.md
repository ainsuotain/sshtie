# sshtie — Persistent SSH sessions with mosh + tmux, zero config

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](#)
[![Release](https://img.shields.io/github/v/release/ainsuotain/sshtie)](https://github.com/ainsuotain/sshtie/releases)

> **Connect once. Stay connected. Auto-fallback, auto-tmux, Tailscale-aware.**

한국어 문서: [README_KO.md](README_KO.md)

---

## What is sshtie?

`sshtie` manages your SSH/mosh/tmux server profiles and **automatically picks the best connection strategy** based on your network. It also ships a native **macOS menu-bar / Windows system-tray app** that shows live server status and lets you connect with one click.

| Problem | sshtie's Solution |
|---|---|
| SSH keeps dropping | mosh auto-connects first |
| mosh UDP is blocked | Automatic SSH fallback + firewall hint |
| tmux attach every time | Auto attach/create on connect |
| Different settings per server | Unified YAML profiles |
| No mosh/tmux on new server | Auto-detects and offers `sshtie install` |
| On Tailscale network | Auto-detected and routed |
| First-time SSH connection | Fingerprint warning before connecting |
| Want to monitor servers at a glance | Menu-bar app with 🟢/🔴 live status |
| Need to tune keepalive or agent forwarding | Per-profile SSH options with slider UI |

---

## Platform Compatibility

| Client | Server | ssh | mosh | tmux |
|--------|--------|:---:|:----:|:----:|
| Mac | Mac | ✅ | ✅ | ✅ |
| Mac | Linux | ✅ | ✅ | ✅ |
| Linux | Mac | ✅ | ✅ | ✅ |
| Linux | Linux | ✅ | ✅ | ✅ |
| Windows (native) | Mac/Linux | ✅ | ❌ | ✅ |
| Windows + WSL | Mac/Linux | ✅ | ✅ | ✅ |
| Any | Windows server | ✅ | ❌ | ❌ |

> **Windows + WSL:** Install sshtie inside WSL (`linux-amd64` binary). The tray app auto-detects WSL and opens a WSL terminal — so mosh works too.

---

## Getting Started

### 3 Steps to Connect

```
  ┌─────────────────┐      ┌──────────────────────┐      ┌─────────────────────┐
  │   Step 1        │      │   Step 2 (optional)  │      │   Step 3            │
  │                 │      │                      │      │                     │
  │  sshtie add     │ ───▶ │  sshtie install      │ ───▶ │  sshtie connect     │
  │                 │      │      <name>          │      │      <name>         │
  │  Register your  │      │  Auto-install mosh   │      │  Connects via best  │
  │  server profile │      │  + tmux on server    │      │  strategy available │
  └─────────────────┘      └──────────────────────┘      └─────────────────────┘
```

---

### Step 1 — Register your server

`sshtie add` opens an interactive TUI wizard. Press Enter to advance, ESC to go back.

```
$ sshtie add

  sshtie add  New Profile        Step 1 / 7

  ▶ Profile name    homeserver█
    A nickname for this connection (e.g., macmini, work-server)
  · Host            (required)
  · User            (required)
  · Port            22
  · SSH Key         ~/.ssh/id_ed25519
  · tmux session    main
  · Network mode    auto

  enter  next  •  esc  back  •  ctrl+c  cancel
```

Only **name**, **host**, and **user** are required. All others have sensible defaults.

**Advanced SSH options** can be set at creation time with flags:

```bash
sshtie add --forward-agent            # enable SSH agent forwarding
sshtie add --attempts=5               # retry up to 5 times
sshtie add --alive-interval=30        # keepalive every 30s
sshtie add --alive-count=40           # drop after 40 missed pings (20 min)
```

---

### Step 2 — Install dependencies *(optional)*

```bash
sshtie install homeserver             # install mosh + tmux on server
sshtie install homeserver --tailscale # also install Tailscale
```

Supports: `apt` · `dnf` · `yum` · `brew` · `pacman`. Works with password auth too.

> **Tip:** `sshtie connect` auto-detects missing tools and offers to install them.

---

### Step 3 — Connect

```bash
sshtie connect homeserver
sshtie homeserver          # shorthand
```

Auto-selects the best strategy:

```
  sshtie connect homeserver
          │
          ▼
  ┌───────────────────┐
  │  mosh available?  │
  │  UDP 60001 open?  │
  └────────┬──────────┘
         Yes│                    No│
            ▼                      ▼
     mosh + tmux              ssh + tmux
     attach/create            attach/create
            │                      │
        fail│                  fail│
            ▼                      ▼
     ssh + tmux               ssh only
```

---

## Commands

| Command | Description |
|---|---|
| `sshtie` | Interactive TUI profile picker |
| `sshtie add [flags]` | Add a new profile (TUI wizard) |
| `sshtie connect <name>` | Connect to a profile |
| `sshtie <name>` | Shorthand for connect |
| `sshtie edit <name>` | Edit advanced SSH options (slider UI) |
| `sshtie list` | List all profiles |
| `sshtie doctor <name>` | Diagnose connection (6 checks) |
| `sshtie install <name>` | Install mosh + tmux on remote server |
| `sshtie remove <name>` | Remove a profile |

---

## sshtie edit — Slider UI

Adjust per-profile SSH options with an interactive slider TUI:

```
$ sshtie edit homeserver

  sshtie edit  homeserver

  ↑/↓ select  ·  ←/→ adjust  ·  shift+←/→ jump  ·  enter save  ·  esc cancel

  ▶ Connection attempts   [━━━━░░░░░░░░░░░░░░░░]    3  (1–10)
    Alive interval        [━━━━━━━━░░░░░░░░░░░░]   10s (10–60s)
    Alive count max       [━━━━━━━━━━━━━━░░░░░░]   60  (6–120)
    Forward agent         ○ on  ● off

  Effective max silence: 600s (10m 00s)
```

---

## macOS Menu-bar / Windows System Tray

A lightweight status app lives in your menu bar (macOS) or system tray (Windows).

**Per-server sub-menu:**

```
🟢● homeserver
    Connect
    ──────────
    Interval: 10s       ← click cycles 10s → 30s → 60s (saved instantly)
    Forward agent: off  ← click toggles on/off (saved instantly)
    Edit SSH Options…   ← opens terminal with slider TUI
    ──────────
    Disconnect          ← shown only when connected
```

**Status indicators:**
- 🟢 — reachable
- 🔴 — unreachable
- 🟡 — checking
- ● — currently connected (active session tracked by PID)

**Features:**
- Auto-refreshes TCP status every 60s, session status every 5s
- **Open at Login** toggle (macOS: LaunchAgent, Windows: Registry)
- Disconnect kills the connection process and cleans up the session file

**Windows tray — WSL detection:**
When "Connect" is clicked, the tray checks for WSL + sshtie-in-WSL automatically:
1. WSL available + sshtie in WSL → opens WSL terminal (mosh supported ✅)
2. Otherwise → opens native Windows terminal (SSH only)

### Building the tray app

```bash
# macOS .app bundle
make menubar

# Run immediately
make menubar-run

# Windows tray (cross-compiled from Mac)
make tray-windows   # → dist/sshtie-tray-windows-amd64.zip
```

---

## sshtie doctor

```
$ sshtie doctor homeserver

  SSH connection       ✅ OK
  mosh-server          ✅ Found (/opt/homebrew/bin/mosh-server)
  UDP port 60001       ✅ Open
  tmux                 ✅ tmux 3.3a installed
  Tailscale (client)   ✅ Running
  Tailscale (server)   ✅ Found in Tailscale network

→ Recommended strategy: mosh + tmux
→ Ready to connect!
```

---

## Install

### Pre-built binaries *(recommended)*

**Linux / WSL**
```bash
cd ~
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-linux-amd64.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

**macOS — Apple Silicon (M1/M2/M3/M4)**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-darwin-arm64.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

**macOS — Intel**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-darwin-amd64.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

**Windows**
Download `sshtie-windows-amd64.zip` from [Releases](https://github.com/ainsuotain/sshtie/releases) and add to PATH.
For mosh support, also install the `linux-amd64` binary inside WSL.

### macOS *(Homebrew)*
```bash
brew tap ainsuotain/sshtie
brew install sshtie
```

### Build from source
```bash
git clone https://github.com/ainsuotain/sshtie
cd sshtie
go build -o sshtie .
```

Requires Go 1.22+. Single static binary, no runtime dependencies.

---

## Profile Schema

`~/.sshtie/profiles.yaml`

```yaml
profiles:
  - name: homeserver
    host: 192.168.1.100
    user: alice
    port: 22
    key: ~/.ssh/id_ed25519
    tmux_session: main
    mosh_server: /opt/homebrew/bin/mosh-server  # optional, auto-detected
    network: auto                               # auto | tailscale | direct

    # Advanced SSH options (omit to use defaults)
    forward_agent: true          # SSH agent forwarding (default: false)
    server_alive_interval: 10    # keepalive interval in seconds (default: 10)
    server_alive_count_max: 60   # missed pings before disconnect (default: 60)
    connection_attempts: 3       # retry attempts (default: 3)
```

---

## Project Structure

```
sshtie/
├── main.go
├── menubar/main.go           # tray app entry point (darwin/windows)
├── cmd/
│   ├── add.go                # TUI wizard + optional SSH flags
│   ├── connect.go            # connection entry point
│   ├── edit.go               # slider TUI for SSH advanced options
│   ├── doctor.go             # diagnostics
│   ├── install.go            # remote dependency installer
│   ├── list.go
│   └── remove.go
└── internal/
    ├── profile/              # YAML profiles (~/.sshtie/profiles.yaml)
    ├── connector/            # mosh/ssh/tmux strategy + session write
    ├── session/              # PID lock files (~/.sshtie/sessions/*.json)
    ├── checker/              # background TCP + session polling
    ├── menubar/              # systray app (darwin/windows)
    ├── tui/                  # Bubble Tea: connect, doctor, edit UIs
    ├── doctor/               # diagnostics logic
    └── tailscale/            # Tailscale detection
```

---

## Roadmap

### v0.1 — MVP ✅
- [x] Profile YAML, connect, list, doctor, remove, install

### v0.2 — TUI ✅
- [x] Bubble Tea TUI wizard for `sshtie add`
- [x] Real-time pre-connect status checks

### v0.3 — Polish ✅
- [x] Tailscale auto-detection, Homebrew tap, pre-built binaries

### v0.4 — Menu-bar App ✅
- [x] macOS menu-bar + Windows system-tray (fyne.io/systray)
- [x] Live 🟢/🔴 server status (TCP polling)
- [x] Click to connect via terminal
- [x] Open at Login (LaunchAgent / Registry)

### v0.5 — Connection Management + SSH Options ✅
- [x] Active session tracking (PID lock files per profile)
- [x] ● active indicator + Disconnect in tray
- [x] `sshtie edit` — per-profile slider UI for SSH advanced options
- [x] Quick Interval / ForwardAgent toggle directly in tray
- [x] WSL detection — tray auto-opens WSL terminal for mosh support
- [x] Unit tests (session, profile, menubar)

### v0.6 — Next
- [ ] `sshtie jump` — SSH jump host / bastion support
- [ ] Auto-reconnect for dropped sessions

---

*Made with ❤️ by [Donghwan Kim (David Kim)](https://github.com/ainsuotain)*
License: [MIT](LICENSE)

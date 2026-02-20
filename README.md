# sshtie — Persistent SSH sessions with mosh + tmux, zero config

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](#)
[![Release](https://img.shields.io/github/v/release/ainsuotain/sshtie)](https://github.com/ainsuotain/sshtie/releases)

> **Connect once. Stay connected. Auto-reconnect, auto-tmux, Tailscale-aware.**

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
| Want to monitor servers at a glance | Menu-bar app with 🟢/🔴 live status ( ● on WIN) |
| Need to tune keepalive or agent forwarding | Per-profile SSH options with slider UI |
| Want Cursor / VS Code to see my servers | Auto-syncs to `~/.ssh/config` on add/remove |
| Close laptop lid → session drops | SSH+tmux sessions **auto-reconnect** when network returns |

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

> **SSH port:** Default is 22. If you don't know the port, just leave it blank — sshtie uses 22 automatically.

> **Windows + WSL:** Install sshtie inside WSL (`linux-amd64` binary). The tray app auto-detects WSL and opens a WSL terminal for mosh support.

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

```bash
sshtie add
```

Interactive TUI wizard — only **name**, **host**, and **user** are required.
Port defaults to 22. Press Enter to skip optional fields.

```
  ▶ Profile name    homeserver█
  · Host            (required)
  · User            (required)
  · Port            22            ← press Enter if unsure
  · SSH Key         ~/.ssh/id_ed25519
  · tmux session    main
  · Network mode    auto
```

After saving, `~/.ssh/config` is **automatically updated** — Cursor and VS Code will see the new server immediately.

**Advanced SSH options** via flags:

```bash
sshtie add --forward-agent            # SSH agent forwarding (for bastion hosts)
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

> `sshtie connect` auto-detects missing tools and offers to install them.

---

### Step 3 — Connect

```bash
sshtie connect homeserver
sshtie homeserver          # shorthand
```

Auto-selects the best strategy:

```
  mosh + tmux  →  ssh + tmux  →  ssh only
```

---

## Auto-Reconnect

When using `ssh + tmux`, sshtie **automatically reconnects** if your connection drops (laptop lid, WiFi switch, VPN change):

```
→ Connecting to homeserver (alice@192.168.1.100)…
[… working …]
[network drops]

⚠  Connection to 'homeserver' dropped.
   Waiting for network to come back (Ctrl+C to cancel)......... ✓
→ Reconnecting... (attempt 1/10)
[tmux session resumes right where you left off]
```

- Polls every 3 s until the server is reachable again
- Up to **10 reconnect attempts** — Ctrl+C to cancel any time
- If using **mosh**, reconnect is handled by mosh itself (even more resilient)

---

## Commands

| Command | Description |
|---|---|
| `sshtie` | Interactive TUI profile picker |
| `sshtie add [flags]` | Add a new profile (TUI wizard) |
| `sshtie connect <name>` | Connect to a profile |
| `sshtie <name>` | Shorthand for connect |
| `sshtie edit <name>` | Edit advanced SSH options (slider UI) |
| `sshtie copy <src> <dst>` | Duplicate a profile with a new name |
| `sshtie list` | List all profiles |
| `sshtie doctor <name>` | Diagnose connection (6 checks) |
| `sshtie install <name>` | Install mosh + tmux on remote server |
| `sshtie rename <name>` | Rename a profile |
| `sshtie remove <name>` | Remove a profile |
| `sshtie ssh-config` | Manually sync all profiles to `~/.ssh/config` |

---

## Interactive TUI

Run `sshtie` with no arguments to open the profile picker:

```
  sshtie  SSH + mosh + tmux, unified

▶ homeserver          alice@192.168.1.100:22   [auto]
  workserver          david@work.example.com:2222  [tailscale]

  ↑/↓  k/j  navigate  •  enter  connect  •  d  doctor  •  e  edit  •  q  quit
```

| Key | Action |
|---|---|
| `enter` | Connect to selected profile |
| `e` | Open edit UI for selected profile |
| `d` | Run doctor on selected profile |
| `q` / `Esc` | Quit |

---

## Cursor / VS Code Integration

sshtie automatically keeps `~/.ssh/config` in sync whenever you add or remove a profile.

```bash
sshtie add
# ✅ Profile 'homeserver' saved!
# ✅ ~/.ssh/config updated (2 profiles)   ← automatic
```

After that, Cursor and VS Code Remote-SSH show the server in their picker without any extra steps.

If you have existing profiles and want to sync manually once:

```bash
sshtie ssh-config
```

The managed entries are wrapped in a clearly marked block — your own SSH config entries are never touched:

```
# BEGIN sshtie managed — do not edit this block manually

Host homeserver
  HostName 192.168.1.100
  User alice
  ServerAliveInterval 10
  ...

# END sshtie managed
```

---

## sshtie edit — Slider UI

Adjust per-profile SSH options interactively:

```
$ sshtie edit homeserver

  ▶ Connection attempts   [━━━━░░░░░░░░░░░░░░░░]    3  (1–10)
    Alive interval        [━━━━━━━━░░░░░░░░░░░░]   10s (10–60s)
    Alive count max       [━━━━━━━━━━━━━━░░░░░░]   60  (6–120)
    Forward agent         ○ on  ● off

  Effective max silence: 600s (10m 00s)

  ─── Profile ──────────────────────────────────────────────
    Rename                homeserver
    Delete profile
```

Controls: `↑/↓` select · `←/→` adjust · `shift+←/→` jump · `enter` save · `esc` cancel

---

## sshtie copy

Duplicate an existing profile with a new name:

```bash
sshtie copy homeserver homeserver-backup
sshtie cp   workserver workserver-dev
```

All settings are copied (host, user, port, SSH options). Edit the new profile with `sshtie edit <name>`.

---

## macOS Menu-bar / Windows System Tray

A lightweight status app in your menu bar (macOS) or system tray (Windows).

**Per-server sub-menu:**

```
🟢  homeserver [connected]
    Connect
    ──────────
    Interval: 10s       ← click cycles 10s → 30s → 60s (saved instantly)
    Forward agent: off  ← click toggles on/off (saved instantly)
    Edit SSH Options…   ← opens terminal with slider TUI
    ──────────
    Rename…
    Remove Profile
    ──────────
    Disconnect          ← shown only when connected
```

**Status:**
- 🟢 reachable · 🔴 unreachable · 🟡 checking
- `[connected]` — active session tracked by PID

**Features:**
- TCP status refresh every 60s, session status every 5s
- **Open at Login** toggle (macOS: LaunchAgent / Windows: Registry)
- **Dark Mode aware** — icon automatically uses the correct color for light/dark mode
- **Windows:** auto-detects WSL — opens WSL terminal for mosh support

### Build

```bash
make menubar          # macOS .app bundle → dist/sshtie-menubar.app
make menubar-run      # build + open immediately
make tray-windows     # Windows tray → dist/sshtie-tray-windows-amd64.zip
```

---

## sshtie doctor

```
$ sshtie doctor homeserver

  SSH connection       ✅ OK
  mosh-server          ✅ Found
  UDP port 60001       ✅ Open
  tmux                 ✅ tmux 3.3a installed
  Tailscale (client)   ✅ Running
  Tailscale (server)   ✅ Found in Tailscale network

→ Recommended strategy: mosh + tmux
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

> **WSL tip:** Run `cd ~` first to move to your Linux home (`/home/<you>`) before running curl.

**macOS — Apple Silicon (M1/M2/M3/M4)**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-mac-apple-silicon.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

**macOS — Intel**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-mac-intel.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

**Windows**
Download `sshtie-windows-amd64.zip` from [Releases](https://github.com/ainsuotain/sshtie/releases) and add to PATH.
For mosh support, also install the `linux-amd64` binary inside WSL.

**Windows Tray App**
Download `sshtie-tray-windows-amd64.zip`, extract both files to the same folder, run `sshtie-tray.exe`.

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

---

## Profile Schema

`~/.sshtie/profiles.yaml`

```yaml
profiles:
  - name: homeserver
    host: 192.168.1.100
    user: alice
    port: 22                    # default: 22 — can be omitted
    key: ~/.ssh/id_ed25519      # omit to use default key
    tmux_session: main
    mosh_server: /opt/homebrew/bin/mosh-server  # optional, auto-detected
    network: auto               # auto | tailscale | direct

    # Advanced SSH options (omit to use defaults)
    forward_agent: true         # SSH agent forwarding (default: false)
    server_alive_interval: 10   # keepalive interval in seconds (default: 10)
    server_alive_count_max: 60  # missed pings before disconnect (default: 60)
    connection_attempts: 3      # retry attempts (default: 3)
```

---

## Server Prerequisites

**macOS server**
- System Settings → General → Sharing → **Remote Login: ON**
- `brew install mosh tmux` *(or use `sshtie install`)*

**Linux server**
- `sshd` must be running
- `sudo apt install mosh tmux` *(or use `sshtie install`)*

**Windows server**
- Settings → Apps → Optional Features → **OpenSSH Server**
- ⚠ mosh and tmux are not supported on Windows servers

---

## Project Structure

```
sshtie/
├── main.go
├── menubar/main.go           # tray app entry point (darwin/windows)
├── cmd/
│   ├── add.go                # TUI wizard + optional SSH flags
│   ├── connect.go
│   ├── copy.go               # duplicate a profile
│   ├── edit.go               # slider TUI for SSH options
│   ├── rename.go
│   ├── ssh_config.go         # ~/.ssh/config sync
│   ├── doctor.go
│   ├── install.go
│   ├── list.go
│   └── remove.go
└── internal/
    ├── profile/              # YAML profiles (~/.sshtie/profiles.yaml)
    ├── connector/            # mosh/ssh/tmux strategy + auto-reconnect
    ├── session/              # PID lock files (~/.sshtie/sessions/*.json)
    ├── checker/              # background TCP + session polling
    ├── menubar/              # systray app (darwin/windows) + dark mode icon
    ├── tui/                  # Bubble Tea UIs (connect, doctor, edit, list)
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
- [x] macOS menu-bar + Windows system-tray
- [x] Live 🟢/🔴 server status, click to connect
- [x] Open at Login

### v0.5 — Connection Management + SSH Options ✅
- [x] Active session tracking (PID lock files)
- [x] `[connected]` indicator + Disconnect in tray
- [x] `sshtie edit` slider UI for SSH options
- [x] Quick Interval / ForwardAgent toggle in tray
- [x] WSL auto-detection for mosh on Windows
- [x] Auto-sync `~/.ssh/config` on add/remove (Cursor/VS Code integration)
- [x] Unit tests

### v0.6 — Resilience + Workflow ✅
- [x] **Auto-reconnect** — ssh+tmux sessions reconnect when network returns
- [x] `sshtie copy` — duplicate a profile with a new name
- [x] Main TUI `e` key — edit selected profile directly
- [x] Dark mode icon — correct color on macOS and Windows
- [x] TUI profile list: shows `user@host:port [network]` (no duplicate host)

### v0.7 — Next
- [ ] `sshtie jump` — SSH jump host / bastion support
- [ ] Main TUI `a` key — open add wizard directly from profile list

---

*Made with ❤️ by [Donghwan Kim (David Kim)](https://github.com/ainsuotain)*
License: [MIT](LICENSE)

# sshtie

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](#)

> **SSH + mosh + tmux, unified.**
> One command to connect. Zero config headaches.

한국어 문서: [README_KO.md](README_KO.md)

---

## Demo

![sshtie demo](demo.gif)

---

## What is sshtie?

`sshtie` manages your SSH/mosh/tmux server profiles and **automatically picks the best connection strategy** based on your network.

| Problem | sshtie's Solution |
|---|---|
| SSH keeps dropping | mosh auto-connects first |
| mosh UDP is blocked | Automatic SSH fallback |
| tmux attach every time | Auto attach/create on connect |
| Different settings per server | Unified YAML profiles |
| No mosh/tmux on new server | `sshtie install` sets it up |

---

## Platform Compatibility

The key factor is **what OS the server runs**, not the client.

| Client | Server | mosh | tmux |
|--------|--------|:----:|:----:|
| Mac | Mac | ✅ | ✅ |
| Mac | Linux | ✅ | ✅ |
| Mac | Windows | ❌ | ❌ |
| Linux | Mac | ✅ | ✅ |
| Linux | Linux | ✅ | ✅ |
| Linux | Windows | ❌ | ❌ |
| Windows | Mac | ✅ | ✅ |
| Windows | Linux | ✅ | ✅ |
| Windows | Windows | ❌ | ❌ |

> **Why?** `mosh-server` and `tmux` run on the **server** side — Windows servers don't support them.
> Windows clients can use mosh if WSL has mosh installed; SSH always works on any combination.

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

`sshtie add` opens an interactive TUI wizard. Navigate with Enter / ESC.

```
$ sshtie add

  sshtie add  New Profile        Step 1 / 7

  ▶ Profile name    homeserver█
  · Host            (required)
  · User            (required)
  · Port            22
  · SSH Key         ~/.ssh/id_ed25519
  · tmux session    main
  · Network mode    auto

  enter  next  •  esc  back  •  ctrl+c  cancel
```

```
✅ Profile 'homeserver' saved!
→ Try: sshtie connect homeserver
```

---

### Step 2 — Install dependencies on the server *(optional)*

If mosh or tmux is not installed on your remote server:

```
$ sshtie install homeserver

🔧 Installing dependencies on homeserver (192.168.1.100)...

  Detecting OS...           ✅ Ubuntu 22.04 LTS (apt)
  tmux...                   ✅ Already installed
  mosh-server...            Installing...
  mosh-server...            ✅ Installed

→ Server is ready!
→ Running doctor check...
```

Supported package managers: `apt` · `dnf` · `yum` · `brew` · `pacman`

---

### Step 3 — Connect

```bash
sshtie connect homeserver

# shorthand (same thing)
sshtie homeserver
```

sshtie automatically tries the best strategy:

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
                          (bare connection)
```

On failure, you always see *why*:
```
⚠  mosh failed: UDP port 60001 appears blocked
→  Falling back to SSH + tmux
```

---

## Commands

| Command | Description |
|---|---|
| `sshtie` | Launch interactive TUI profile picker |
| `sshtie add` | Add a new profile (TUI wizard) |
| `sshtie list` | List all profiles |
| `sshtie connect <name>` | Connect to a profile |
| `sshtie <name>` | Shorthand for connect |
| `sshtie edit <name>` | Edit a profile in `$EDITOR` |
| `sshtie install <name>` | Install mosh + tmux on remote server |
| `sshtie doctor <name>` | Diagnose connection |
| `sshtie remove <name>` | Remove a profile |

---

## sshtie doctor

Runs a full connectivity check before you connect:

```
$ sshtie doctor homeserver

🔍 Diagnosing: homeserver (192.168.1.100)

  SSH connection       ✅ OK
  mosh-server          ✅ Found (/opt/homebrew/bin/mosh-server)
  UDP port 60001       ✅ Open (or filtered — mosh will confirm)
  tmux                 ✅ tmux 3.3a installed
  Tailscale (client)   ✅ Running
  Tailscale (server)   ✅ Found in Tailscale network

→ Recommended strategy: mosh + tmux
→ Ready to connect!
```

---

## Install

### Client (sshtie tool)

**macOS**
```bash
# Build from source (Homebrew tap coming soon)
git clone https://github.com/ainsuotain/sshtie
cd sshtie
go build -o sshtie .
sudo mv sshtie /usr/local/bin/
```

**Linux**
```bash
git clone https://github.com/ainsuotain/sshtie
cd sshtie
go build -o sshtie .
sudo mv sshtie /usr/local/bin/
```

**Windows**
```powershell
git clone https://github.com/ainsuotain/sshtie
cd sshtie
go build -o sshtie.exe .
# Move sshtie.exe to a directory in %PATH%
```
> WSL is recommended for mosh support on Windows clients.

Requires Go 1.22+. No external runtime dependencies — single static binary.

---

### Server Prerequisites

**macOS server**
- System Settings → General → Sharing → **Remote Login: ON**
- Install mosh + tmux: `brew install mosh tmux` *(or use `sshtie install`)*

**Linux server**
- `sshd` must be running
- Install mosh + tmux: `sudo apt install mosh tmux` *(or use `sshtie install`)*

**Windows server**
- Settings → Apps → Optional Features → **OpenSSH Server**
- ⚠ mosh and tmux are **not supported** on Windows servers — SSH only

---

## Profile Schema

`~/.sshtie/profiles.yaml`

```yaml
profiles:
  - name: homeserver
    host: 192.168.1.100
    user: alice
    port: 22                                    # default: 22
    key: ~/.ssh/id_ed25519                      # uses default key if omitted
    tmux_session: main                          # default: main
    mosh_server: /opt/homebrew/bin/mosh-server  # optional, auto-detected
    network: auto                               # auto | tailscale | direct
    tags: [home, personal]

  - name: work-server
    host: work.example.com
    user: bob
    port: 2222
    tmux_session: work
    network: direct
    tags: [work, production]
```

---

## Project Structure

```
sshtie/
├── main.go
├── go.mod
├── cmd/
│   ├── root.go       # cobra root + sshtie <name> shorthand
│   ├── add.go        # TUI wizard profile creation
│   ├── connect.go    # connection entry point
│   ├── edit.go       # open profile in $EDITOR
│   ├── install.go    # remote mosh + tmux + tailscale installer
│   ├── list.go       # profile listing
│   ├── doctor.go     # connectivity diagnostics
│   └── remove.go     # profile deletion
└── internal/
    ├── profile/      # YAML read/write (~/.sshtie/profiles.yaml)
    ├── connector/    # connection strategy (mosh/ssh/tmux fallback)
    ├── doctor/       # diagnostics logic
    ├── tailscale/    # Tailscale client/peer detection
    └── tui/          # Bubble Tea interactive profile picker
```

---

## Tech Stack

| | |
|---|---|
| Language | Go 1.22 — single binary, cross-platform |
| CLI framework | [Cobra](https://github.com/spf13/cobra) |
| Config format | YAML ([gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3)) |
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss) |

---

## Roadmap

### v0.1 — MVP ✅
- [x] Profile YAML read/write
- [x] `sshtie add` interactive input
- [x] `sshtie connect` — mosh → ssh fallback → tmux attach
- [x] `sshtie list`
- [x] `sshtie doctor` diagnostics
- [x] `sshtie remove`
- [x] `sshtie install` — remote mosh/tmux installer (apt/dnf/yum/brew/pacman)
- [x] Windows / Linux / macOS cross-platform support

### v0.2 — TUI ✅
- [x] Bubble Tea TUI (runs when no args given)
- [x] `sshtie edit <name>` — open profile in $EDITOR
- [x] TUI wizard for `sshtie add`

### v0.3 — Polish ✅
- [x] Tailscale auto-detection (client + server)
- [x] `sshtie install --tailscale`
- [ ] Homebrew tap distribution
- [ ] Live connection status display

---

*Made with ❤️ by [Donghwan Kim (David Kim)](https://github.com/ainsuotain)*
License: [Apache 2.0](LICENSE)

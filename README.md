# sshtie

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](#)
[![Release](https://img.shields.io/github/v/release/ainsuotain/sshtie)](https://github.com/ainsuotain/sshtie/releases)

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
| mosh UDP is blocked | Automatic SSH fallback + firewall hint |
| tmux attach every time | Auto attach/create on connect |
| Different settings per server | Unified YAML profiles |
| No mosh/tmux on new server | Auto-detects and offers `sshtie install` |
| On Tailscale network | Auto-detected and routed |
| First-time SSH connection | Fingerprint warning before connecting |
| Server uses password auth | `sshtie install` works without SSH key |

---

## Platform Compatibility

The key factor is **what OS the server runs**, not the client.

| Client | Server | ssh | mosh | tmux |
|--------|--------|:---:|:----:|:----:|
| Mac | Mac | ✅ | ✅ | ✅ |
| Mac | Linux | ✅ | ✅ | ✅ |
| Mac | Windows | ✅ | ❌ | ❌ |
| Linux | Mac | ✅ | ✅ | ✅ |
| Linux | Linux | ✅ | ✅ | ✅ |
| Linux | Windows | ✅ | ❌ | ❌ |
| Windows | Mac | ✅ | ✅ | ✅ |
| Windows | Linux | ✅ | ✅ | ✅ |
| Windows | Windows | ✅ | ❌ | ❌ |

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

`sshtie add` opens an interactive 7-step TUI wizard. Press Enter to advance, ESC to go back.

```
$ sshtie add

  sshtie add  New Profile        Step 1 / 7

  ▶ Profile name    homeserver█
    A nickname for this connection  (e.g. macmini, work-server, linux01)
  · Host            (required)
  · User            (required)
  · Port            22
  · SSH Key         ~/.ssh/id_ed25519
  · tmux session    main
  · Network mode    auto

  enter  next  •  esc  back  •  ctrl+c  cancel
```

Only three fields are required: **name**, **host**, and **user**.
All others have sensible defaults — just press Enter.

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

Works with **both SSH key and password authentication** — if you haven't set up SSH keys yet, it will prompt for your server password.

> **Tip:** You don't need to run `sshtie install` manually.
> `sshtie connect` automatically detects missing dependencies and asks if you want to install them.

To also install **Tailscale** on the server:

```bash
sshtie install homeserver --tailscale
```

Supported package managers: `apt` · `dnf` · `yum` · `brew` · `pacman`

Error guidance:
- No sudo access → prints manual install commands
- macOS without Homebrew → directs to `https://brew.sh`
- Unknown OS → shows 5 manual install commands

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

On failure, you always see *why* — and what to do:
```
⚠  mosh: UDP 포트가 차단되어 있습니다.
   서버에서 실행하세요: sudo ufw allow 60000:61000/udp
→ SSH로 폴백합니다.
```

**Smart pre-connect checks** (new in v0.2):

*First-time host — fingerprint warning:*
```
⚠  처음 접속하는 서버입니다 (192.168.1.100)
   SSH 키가 자동으로 저장됩니다.
계속할까요? (y/n):
```

*Missing mosh / tmux — auto-install offer:*
```
⚠  서버에 mosh-server, tmux 가 설치되어 있지 않습니다.
지금 설치할까요? (y/n): y

🔧 Installing dependencies on homeserver (192.168.1.100)...
```

---

## Commands

| Command | Description |
|---|---|
| `sshtie` | Launch interactive TUI profile picker |
| `sshtie add` | Add a new profile (7-step TUI wizard) |
| `sshtie list` | List all profiles |
| `sshtie connect <name>` | Connect to a profile |
| `sshtie <name>` | Shorthand for connect |
| `sshtie edit <name>` | Edit a profile in `$EDITOR` |
| `sshtie install <name>` | Install mosh + tmux on remote server |
| `sshtie install <name> --tailscale` | Also install Tailscale on remote server |
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

## Network Modes

Set via the `network` field in a profile (or chosen in `sshtie add` wizard):

| Mode | Behavior |
|---|---|
| `auto` *(default)* | Detects Tailscale automatically; tries mosh first, falls back to SSH |
| `tailscale` | Requires Tailscale — fails fast if unavailable or host not found |
| `direct` | Skips Tailscale and mosh entirely; connects via SSH directly |

---

## Install

### Pre-built binaries *(recommended)*

**Linux**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-linux-amd64.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

**macOS — Apple Silicon (M1/M2/M3)**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-darwin-arm64.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

**macOS — Intel**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-darwin-amd64.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

**Windows (WSL)**
```bash
# Important: run from your Linux home directory, not the Windows path.
cd ~
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-linux-amd64.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

> **Why `cd ~` first?**
> WSL often starts in `/mnt/c/Users/<you>` — a Windows-mounted path with restricted permissions.
> Running `curl` or `sudo mv` from there causes `Permission denied` errors.
> `cd ~` takes you to your real Linux home (`/home/<you>`) where everything works normally.
>
> To make WSL always start in your Linux home, add this to `/etc/wsl.conf`:
> ```ini
> [user]
> default=<your-username>
> ```
> Then restart WSL: `wsl --shutdown` in PowerShell.

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
sudo mv sshtie /usr/local/bin/
```

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
├── Makefile                  # cross-compile + release automation
├── Formula/sshtie.rb         # Homebrew tap formula
├── cmd/
│   ├── root.go               # cobra root + sshtie <name> shorthand
│   ├── add.go                # 7-step TUI wizard profile creation
│   ├── connect.go            # connection entry point
│   ├── edit.go               # open profile in $EDITOR
│   ├── install.go            # remote mosh + tmux + tailscale installer
│   ├── list.go               # profile listing
│   ├── doctor.go             # connectivity diagnostics
│   └── remove.go             # profile deletion
└── internal/
    ├── profile/              # YAML read/write (~/.sshtie/profiles.yaml)
    ├── connector/            # connection strategy (mosh/ssh/tmux fallback)
    ├── doctor/               # diagnostics logic (6 checks)
    ├── tailscale/            # Tailscale client/peer detection
    └── tui/                  # Bubble Tea interactive profile picker
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
- [x] TUI wizard for `sshtie add` (7-step with hints)

### v0.3 — Polish ✅
- [x] Tailscale auto-detection (client + server)
- [x] `sshtie install --tailscale`
- [x] Homebrew tap (`ainsuotain/homebrew-sshtie`)
- [x] Pre-built binaries for all platforms

### v0.2.1 — Smart UX ✅
- [x] Auto-detect missing mosh/tmux on connect → offer to install
- [x] `sshtie install` supports password authentication (no SSH key required)
- [x] UDP blocked: show server-side firewall command (`sudo ufw allow 60000:61000/udp`)
- [x] First-time SSH fingerprint warning before connecting
- [x] WSL `cd ~` guidance in README

### v0.4 — Next
- [ ] Live connection status display
- [ ] `sshtie jump` — SSH jump host / bastion support

---

*Made with ❤️ by [Donghwan Kim (David Kim)](https://github.com/ainsuotain)*
License: [MIT](LICENSE)

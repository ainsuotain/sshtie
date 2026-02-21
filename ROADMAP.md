# sshtie Roadmap

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

### v0.7 — Windows UX + macOS signing ✅
- [x] **Windows CMD window hide** — clicking X keeps SSH alive in background; [connected] persists
- [x] **Blank window flash fixed** — tray now spawns sshtie.exe directly (no CMD wrapper)
- [x] **macOS ad-hoc codesigning** — CLI + menubar app signed at build time; right-click → Open to bypass Gatekeeper
- [x] `sshtie copy` + main TUI `e` key (shipped with v0.6)

### v0.8 — Next
- [ ] `sshtie jump` — SSH jump host / bastion support
- [ ] Main TUI `a` key — open add wizard directly from profile list

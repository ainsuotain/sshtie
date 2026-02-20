# sshtie — mosh + tmux로 끊기지 않는 SSH 세션, 설정 불필요

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](#)
[![Release](https://img.shields.io/github/v/release/ainsuotain/sshtie)](https://github.com/ainsuotain/sshtie/releases)

> **한 번 접속하면 끊기지 않습니다. 자동 재연결, 자동 tmux, Tailscale 지원.**

English docs: [README.md](README.md)

---

## sshtie란?

`sshtie`는 SSH/mosh/tmux 서버 프로파일을 관리하고 **네트워크 환경에 따라 최적 연결 방식을 자동으로 선택**해주는 CLI 툴입니다. macOS 메뉴바 / Windows 시스템 트레이 앱도 포함되어 있어 서버 상태를 한눈에 확인하고 클릭 한 번으로 접속할 수 있습니다.

| 문제 | sshtie의 해결 |
|---|---|
| SSH가 자꾸 끊긴다 | mosh를 우선 시도해 안정적 연결 |
| mosh UDP가 막힌 환경 | SSH로 자동 폴백 + 방화벽 명령어 안내 |
| tmux 매번 수동 attach | 접속 시 자동 attach/create |
| 서버마다 설정이 달라 헷갈림 | YAML 프로파일로 통합 관리 |
| 새 서버에 mosh/tmux가 없다 | `sshtie install`로 자동 설치 |
| 서버 상태를 한눈에 보고 싶다 | 메뉴바 앱으로 🟢/🔴 실시간 확인( ● on WIN) |
| keepalive나 에이전트 포워딩 조정 | 슬라이더 UI로 프로파일별 설정 |
| Cursor / VS Code에서 서버를 보고 싶다 | add/remove 시 `~/.ssh/config` 자동 동기화 |
| 노트북 덮었다 열면 세션이 끊김 | ssh+tmux 세션 **자동 재연결** (네트워크 복구 시) |

---

## 플랫폼 호환성

| 클라이언트 | 서버 | ssh | mosh | tmux |
|--------|--------|:---:|:----:|:----:|
| Mac | Mac | ✅ | ✅ | ✅ |
| Mac | Linux | ✅ | ✅ | ✅ |
| Linux | Mac | ✅ | ✅ | ✅ |
| Linux | Linux | ✅ | ✅ | ✅ |
| Windows (네이티브) | Mac/Linux | ✅ | ❌ | ✅ |
| Windows + WSL | Mac/Linux | ✅ | ✅ | ✅ |
| 모든 클라이언트 | Windows 서버 | ✅ | ❌ | ❌ |

> **SSH 포트:** 기본은 22입니다. 포트를 모르면 그냥 Enter — sshtie가 22번을 자동으로 사용합니다.

> **Windows + WSL:** WSL 안에 sshtie(`linux-amd64` 바이너리)를 설치하면 됩니다. 트레이 앱이 WSL을 자동으로 감지해서 mosh까지 지원하는 WSL 터미널로 열어줍니다.

---

## 시작하기

### 3단계로 접속

```
  ┌──────────────────┐      ┌────────────────────────┐      ┌──────────────────────┐
  │   1단계          │      │   2단계 (선택)         │      │   3단계              │
  │                  │      │                        │      │                      │
  │   sshtie add     │ ───▶ │   sshtie install       │ ───▶ │   sshtie connect     │
  │                  │      │       <name>           │      │       <name>         │
  │  서버 프로파일을 │      │  원격 서버에 mosh +    │      │  가능한 최적 방식으로│
  │  등록합니다      │      │  tmux 자동 설치        │      │  자동 접속           │
  └──────────────────┘      └────────────────────────┘      └──────────────────────┘
```

---

### 1단계 — 서버 등록

```bash
sshtie add
```

TUI 위자드 실행 — 필수 항목은 **이름·호스트·유저** 3가지뿐. Enter로 다음, ESC로 이전.

```
  ▶ Profile name    homeserver█
  · Host            (required)
  · User            (required)
  · Port            22            ← 모르면 Enter
  · SSH Key         ~/.ssh/id_ed25519
  · tmux session    main
  · Network mode    auto
```

저장 후 `~/.ssh/config`가 **자동으로 업데이트**되어 Cursor와 VS Code에서 즉시 서버를 확인할 수 있습니다.

**고급 SSH 옵션** (생성 시 플래그로 설정):

```bash
sshtie add --forward-agent            # SSH 에이전트 포워딩 활성화
sshtie add --attempts=5               # 연결 최대 5회 재시도
sshtie add --alive-interval=30        # 30초마다 keepalive
sshtie add --alive-count=40           # 40회 무응답 시 연결 끊기 (20분)
```

---

### 2단계 — 원격 서버에 의존성 설치 *(선택)*

```bash
sshtie install homeserver             # mosh + tmux 설치
sshtie install homeserver --tailscale # Tailscale도 함께 설치
```

지원 패키지 매니저: `apt` · `dnf` · `yum` · `brew` · `pacman`. 비밀번호 인증도 지원.

> `sshtie connect`가 누락된 도구를 자동으로 감지하고 설치 여부를 물어봅니다.

---

### 3단계 — 접속

```bash
sshtie connect homeserver
sshtie homeserver          # 단축 사용
```

자동으로 최적 전략 선택:

```
  mosh + tmux  →  ssh + tmux  →  ssh only
```

---

## 자동 재연결

`ssh + tmux` 사용 시, 네트워크가 끊겼다가 복구되면 **자동으로 재접속**합니다 (노트북 덮기, WiFi 전환, VPN 변경 등):

```
→ Connecting to homeserver (alice@192.168.1.100)…
[… 작업 중 …]
[네트워크 끊김]

⚠  Connection to 'homeserver' dropped.
   Waiting for network to come back (Ctrl+C to cancel)......... ✓
→ Reconnecting... (attempt 1/10)
[tmux 세션이 끊기기 전 상태 그대로 재개]
```

- 3초마다 서버 TCP 연결 가능 여부 확인
- 최대 **10회** 재시도 — 언제든 Ctrl+C로 취소 가능
- **mosh** 사용 시: mosh 자체가 재연결 처리 (더 안정적)

---

## 커맨드 목록

| 커맨드 | 설명 |
|---|---|
| `sshtie` | TUI 프로파일 선택기 실행 |
| `sshtie add [flags]` | 프로파일 추가 (TUI 위자드) |
| `sshtie connect <name>` | 접속 |
| `sshtie <name>` | connect 단축키 |
| `sshtie edit <name>` | 고급 SSH 옵션 슬라이더 UI |
| `sshtie copy <src> <dst>` | 프로파일 복제 (새 이름으로) |
| `sshtie list` | 프로파일 목록 |
| `sshtie doctor <name>` | 연결 진단 (6가지 체크) |
| `sshtie install <name>` | 원격 서버에 mosh + tmux 자동 설치 |
| `sshtie rename <name>` | 프로파일 이름 변경 |
| `sshtie remove <name>` | 프로파일 삭제 |
| `sshtie ssh-config` | 전체 프로파일을 `~/.ssh/config`에 수동 동기화 |

---

## 인터랙티브 TUI

인수 없이 `sshtie`를 실행하면 프로파일 선택 화면이 열립니다:

```
  sshtie  SSH + mosh + tmux, unified

▶ homeserver          alice@192.168.1.100:22   [auto]
  workserver          david@work.example.com:2222  [tailscale]

  ↑/↓  k/j  navigate  •  enter  connect  •  d  doctor  •  e  edit  •  q  quit
```

| 키 | 동작 |
|---|---|
| `enter` | 선택한 프로파일에 접속 |
| `e` | 선택한 프로파일 편집 화면 열기 |
| `d` | 선택한 프로파일 doctor 실행 |
| `q` / `Esc` | 종료 |

---

## Cursor / VS Code 통합

sshtie는 프로파일을 추가하거나 삭제할 때마다 `~/.ssh/config`를 자동으로 동기화합니다.

```bash
sshtie add
# ✅ Profile 'homeserver' saved!
# ✅ ~/.ssh/config updated (2 profiles)   ← 자동
```

이후 Cursor와 VS Code Remote-SSH의 서버 목록에 서버가 자동으로 표시됩니다.

기존 프로파일을 한 번에 동기화하려면:

```bash
sshtie ssh-config
```

관리되는 항목은 명확한 블록으로 구분되며, 기존 SSH config 항목은 절대 건드리지 않습니다:

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

## sshtie edit — 슬라이더 UI

프로파일별 SSH 옵션을 슬라이더로 직관적으로 조정:

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

조작: `↑/↓` 선택 · `←/→` 조정 · `shift+←/→` 점프 · `enter` 저장 · `esc` 취소

---

## sshtie copy

프로파일을 새 이름으로 복제:

```bash
sshtie copy homeserver homeserver-backup
sshtie cp   workserver workserver-dev
```

모든 설정(host, user, port, SSH 옵션 등)이 그대로 복제됩니다. 이후 `sshtie edit <name>`으로 수정하면 됩니다.

---

## macOS 메뉴바 / Windows 시스템 트레이

메뉴바(macOS) 또는 시스템 트레이(Windows)에 상주하는 가벼운 상태 앱.

**서버별 서브메뉴:**

```
🟢  homeserver [connected]
    Connect
    ──────────
    Interval: 10s       ← 클릭 시 10s → 30s → 60s 순환 (즉시 저장)
    Forward agent: off  ← 클릭 시 on/off 토글 (즉시 저장)
    Edit SSH Options…   ← 터미널 열고 슬라이더 TUI 실행
    ──────────
    Rename…
    Remove Profile
    ──────────
    Disconnect          ← 연결 중일 때만 표시
```

**상태 표시:**
- 🟢 — 접속 가능 · 🔴 — 접속 불가 · 🟡 — 확인 중
- `[connected]` — PID 기반 활성 세션 추적

**주요 기능:**
- TCP 상태 60초마다 자동 갱신, 세션 상태 5초마다 갱신
- **Open at Login** 토글 (macOS: LaunchAgent, Windows: 레지스트리)
- **다크모드 대응** — macOS/Windows 모두 다크모드에서 아이보리 색 아이콘 자동 적용
- **Windows WSL 자동 감지** — mosh 지원 WSL 터미널로 자동 연결

### 트레이 앱 빌드

```bash
make menubar          # macOS .app 번들 → dist/sshtie-menubar.app
make menubar-run      # 빌드 후 바로 실행
make tray-windows     # Windows 트레이 → dist/sshtie-tray-windows-amd64.zip
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

## 설치

### 사전 빌드 바이너리 *(권장)*

**Linux / WSL**
```bash
cd ~
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-linux-amd64.tar.gz | tar -xz
sudo mv sshtie /usr/local/bin/
```

> **WSL 팁:** `cd ~`로 Linux 홈(`/home/<이름>`)으로 이동 후 실행하세요.

**macOS — Apple Silicon (M1/M2/M3/M4)**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-mac-apple-silicon.tar.gz | tar -xz
xattr -c sshtie          # macOS 격리 플래그 해제
sudo mv sshtie /usr/local/bin/
```

**macOS — Intel**
```bash
curl -L https://github.com/ainsuotain/sshtie/releases/latest/download/sshtie-mac-intel.tar.gz | tar -xz
xattr -c sshtie          # macOS 격리 플래그 해제
sudo mv sshtie /usr/local/bin/
```

> **macOS 보안 경고가 뜨는 경우:** *"개발자를 확인할 수 없어 열 수 없습니다"* 라는 메시지가 뜨면, 같은 폴더에서 `xattr -c sshtie`를 실행한 뒤 이동하세요. Homebrew로 설치하면 이 과정이 자동으로 처리됩니다.

**Windows**
[Releases](https://github.com/ainsuotain/sshtie/releases)에서 `sshtie-windows-amd64.zip` 다운로드 후 PATH에 추가.
mosh 지원을 원하면 WSL 안에 `linux-amd64` 바이너리도 설치하세요.

**Windows 트레이 앱**
`sshtie-tray-windows-amd64.zip`을 다운로드하고, 두 파일을 같은 폴더에 압축 해제한 뒤 `sshtie-tray.exe`를 실행하세요.

> **프로파일 데이터는 `~/.sshtie/profiles.yaml`에 저장됩니다.** 바이너리를 업데이트하거나 재설치해도 기존 서버 정보는 그대로 유지됩니다.

### macOS *(Homebrew)*
```bash
brew tap ainsuotain/sshtie
brew install sshtie
```

### 소스에서 빌드
```bash
git clone https://github.com/ainsuotain/sshtie
cd sshtie
go build -o sshtie .
```

Go 1.22 이상 필요. 단일 바이너리, 외부 런타임 의존성 없음.

---

## 프로파일 설정

`~/.sshtie/profiles.yaml`

```yaml
profiles:
  - name: homeserver
    host: 192.168.1.100
    user: alice
    port: 22
    key: ~/.ssh/id_ed25519
    tmux_session: main
    mosh_server: /opt/homebrew/bin/mosh-server  # 생략 시 자동 감지
    network: auto                               # auto | tailscale | direct

    # 고급 SSH 옵션 (생략 시 기본값 사용)
    forward_agent: true          # SSH 에이전트 포워딩 (기본: false)
    server_alive_interval: 10    # keepalive 간격 초 (기본: 10)
    server_alive_count_max: 60   # 무응답 허용 횟수 (기본: 60)
    connection_attempts: 3       # 연결 재시도 횟수 (기본: 3)
```

---

## 서버 사전 조건

**macOS 서버**
- 시스템 설정 → 일반 → 공유 → **원격 로그인: ON**
- `brew install mosh tmux` *(또는 `sshtie install` 사용)*

**Linux 서버**
- `sshd` 실행 중이어야 함
- `sudo apt install mosh tmux` *(또는 `sshtie install` 사용)*

**Windows 서버**
- 설정 → 앱 → 선택적 기능 → **OpenSSH 서버**
- ⚠ Windows 서버에서는 mosh, tmux 미지원

---

## 프로젝트 구조

```
sshtie/
├── main.go
├── menubar/main.go           # 트레이 앱 진입점 (darwin/windows)
├── cmd/
│   ├── add.go                # TUI 위자드 + SSH 옵션 플래그
│   ├── connect.go
│   ├── copy.go               # 프로파일 복제
│   ├── edit.go               # 슬라이더 TUI
│   ├── rename.go
│   ├── ssh_config.go         # ~/.ssh/config 동기화
│   ├── doctor.go
│   ├── install.go
│   ├── list.go
│   └── remove.go
└── internal/
    ├── profile/              # YAML 프로파일 (~/.sshtie/profiles.yaml)
    ├── connector/            # mosh/ssh/tmux 전략 + 자동 재연결
    ├── session/              # PID 락 파일 (~/.sshtie/sessions/*.json)
    ├── checker/              # 백그라운드 TCP + 세션 폴링
    ├── menubar/              # systray 앱 + 다크모드 아이콘
    ├── tui/                  # Bubble Tea UI (connect, doctor, edit, list)
    ├── doctor/               # 진단 로직
    └── tailscale/            # Tailscale 감지
```

---

## 로드맵

### v0.1 — MVP ✅
- [x] 프로파일 YAML, connect, list, doctor, remove, install

### v0.2 — TUI ✅
- [x] Bubble Tea TUI 위자드 (`sshtie add`)
- [x] 실시간 사전 연결 상태 체크

### v0.3 — 완성도 ✅
- [x] Tailscale 자동 감지, Homebrew tap, 사전 빌드 바이너리

### v0.4 — 메뉴바 앱 ✅
- [x] macOS 메뉴바 + Windows 시스템 트레이
- [x] 실시간 🟢/🔴 서버 상태
- [x] Open at Login (LaunchAgent / 레지스트리)

### v0.5 — 연결 관리 + SSH 옵션 ✅
- [x] 활성 세션 추적 (PID 락 파일)
- [x] `[connected]` 표시 + 트레이에서 Disconnect
- [x] `sshtie edit` 슬라이더 UI
- [x] 트레이에서 Interval / ForwardAgent 빠른 조정
- [x] WSL 자동 감지 (트레이 → WSL 터미널 → mosh 지원)
- [x] `~/.ssh/config` 자동 동기화 (Cursor/VS Code 통합)
- [x] 단위 테스트

### v0.6 — 안정성 + 워크플로우 ✅
- [x] **자동 재연결** — ssh+tmux 세션이 네트워크 복구 시 자동 재접속
- [x] `sshtie copy` — 프로파일 복제
- [x] 메인 TUI `e` 키 — 선택한 프로파일 바로 편집
- [x] 다크모드 아이콘 — macOS/Windows 다크모드에서 아이보리 색
- [x] TUI 목록: `user@host:port [network]` 형식 (host 중복 제거)

### v0.7 — 다음
- [ ] `sshtie jump` — SSH 점프 호스트 / 배스쳔 지원
- [ ] 메인 TUI `a` 키 — 목록에서 바로 프로파일 추가

---

*Made with ❤️ by [Donghwan Kim (David Kim)](https://github.com/ainsuotain)*
License: [MIT](LICENSE)

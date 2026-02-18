# sshtie

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](#)

> **SSH + mosh + tmux를 하나로.**
> 명령어 하나로 접속. 설정 고민 없이.

English docs: [README.md](README.md)

---

## 데모

![sshtie demo](demo.gif)

---

## sshtie란?

`sshtie`는 SSH/mosh/tmux 서버 프로파일을 관리하고,
**네트워크 환경에 따라 최적 연결 방식을 자동으로 선택**해주는 CLI 툴입니다.

| 문제 | sshtie의 해결 |
|---|---|
| SSH가 자꾸 끊긴다 | mosh를 우선 시도해 안정적 연결 |
| mosh UDP가 막힌 환경 | SSH로 자동 폴백 |
| tmux 매번 수동 attach | 접속 시 자동 attach/create |
| 서버마다 설정이 달라 헷갈림 | YAML 프로파일로 통합 관리 |
| 새 서버에 mosh/tmux가 없다 | `sshtie install`로 자동 설치 |

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

```
$ sshtie add

Profile name           : homeserver
Host                   : 192.168.1.100
User                   : alice
Port                   [22]:
SSH Key                [~/.ssh/id_ed25519]:
tmux session           [main]:
Network mode           [auto]:

✅ Profile 'homeserver' saved!
→ Try: sshtie connect homeserver
```

필수 항목은 이름·호스트·유저 3가지만.
나머지는 엔터로 기본값 사용 가능합니다.

---

### 2단계 — 원격 서버에 의존성 설치 *(선택)*

원격 서버에 mosh 또는 tmux가 없을 때:

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

지원 패키지 매니저: `apt` · `dnf` · `yum` · `brew` · `pacman`

에러 시 안내:
- sudo 권한 없음 → 수동 설치 명령어 출력
- macOS + brew 없음 → `https://brew.sh` 안내
- OS 감지 실패 → 5종 수동 명령어 안내

---

### 3단계 — 접속

```bash
sshtie connect homeserver

# 단축 사용 (connect 생략 가능)
sshtie homeserver
```

sshtie가 자동으로 최적 전략을 선택합니다:

```
  sshtie connect homeserver
          │
          ▼
  ┌──────────────────────┐
  │  mosh 설치 확인      │
  │  UDP 60001 열려 있나?│
  └──────────┬───────────┘
           Yes│                     No│
              ▼                       ▼
       mosh + tmux              ssh + tmux
       attach/create            attach/create
              │                       │
          실패│                   실패│
              ▼                       ▼
       ssh + tmux               ssh only
                             (bare connection)
```

실패 시 이유 출력:
```
⚠  mosh failed: UDP port 60001 appears blocked
→  Falling back to SSH + tmux
```

---

## 커맨드 목록

| 커맨드 | 설명 |
|---|---|
| `sshtie add` | 프로파일 추가 (대화형) |
| `sshtie list` | 프로파일 목록 |
| `sshtie connect <name>` | 접속 |
| `sshtie <name>` | connect 단축키 |
| `sshtie install <name>` | 원격 서버에 mosh + tmux 자동 설치 |
| `sshtie doctor <name>` | 연결 진단 |
| `sshtie remove <name>` | 프로파일 삭제 |

---

## sshtie doctor

접속 전 연결 상태를 미리 점검합니다:

```
$ sshtie doctor homeserver

🔍 Diagnosing: homeserver (192.168.1.100)

  SSH connection       ✅ OK
  mosh-server          ✅ Found (/opt/homebrew/bin/mosh-server)
  UDP port 60001       ✅ Open (or filtered — mosh will confirm)
  tmux                 ✅ tmux 3.3a installed
  Tailscale            ⚠  Not detected (optional)

→ Recommended strategy: mosh + tmux
→ Ready to connect!
```

---

## 설치

```bash
git clone https://github.com/ainsuotain/sshtie
cd sshtie
go build -o sshtie .

# PATH에 추가
sudo mv sshtie /usr/local/bin/
```

Go 1.22 이상 필요. 외부 런타임 의존성 없음 — 단일 바이너리.

---

## 프로파일 설정

`~/.sshtie/profiles.yaml`

```yaml
profiles:
  - name: homeserver
    host: 192.168.1.100
    user: alice
    port: 22                                    # 기본값: 22
    key: ~/.ssh/id_ed25519                      # 생략 시 기본 키 사용
    tmux_session: main                          # 기본값: main
    mosh_server: /opt/homebrew/bin/mosh-server  # 생략 시 자동 감지
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

## 프로젝트 구조

```
sshtie/
├── main.go
├── go.mod
├── cmd/
│   ├── root.go       # cobra root + sshtie <name> 단축키
│   ├── add.go        # 대화형 프로파일 추가
│   ├── connect.go    # 연결 진입점
│   ├── install.go    # 원격 mosh + tmux 설치
│   ├── list.go       # 프로파일 목록
│   ├── doctor.go     # 진단
│   └── remove.go     # 삭제
└── internal/
    ├── profile/      # YAML 읽기/쓰기 (~/.sshtie/profiles.yaml)
    ├── connector/    # 연결 전략 (mosh/ssh/tmux 폴백)
    ├── doctor/       # 진단 로직
    └── tui/          # Bubble Tea 대화형 프로파일 선택기
```

---

## 기술 스택

| | |
|---|---|
| 언어 | Go 1.22 — 단일 바이너리, 크로스플랫폼 |
| CLI 프레임워크 | [Cobra](https://github.com/spf13/cobra) |
| 설정 형식 | YAML ([gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3)) |
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |

---

## 로드맵

### v0.1 — MVP ✅
- [x] 프로파일 YAML 읽기/쓰기
- [x] `sshtie add` 대화형 입력
- [x] `sshtie connect` — mosh → ssh fallback → tmux attach
- [x] `sshtie list`
- [x] `sshtie doctor` 기본 진단
- [x] `sshtie remove`
- [x] `sshtie install` — 원격 mosh/tmux 자동 설치 (apt/dnf/yum/brew/pacman)
- [x] Windows / Linux / macOS 크로스플랫폼 지원

### v0.2 — TUI ✅
- [x] Bubble Tea 기반 TUI (인자 없이 실행 시)
- [ ] `sshtie edit <name>`
- [ ] 실시간 연결 상태 표시

### v0.3 — 완성도
- [ ] Tailscale 자동 감지
- [ ] Homebrew tap 배포

---

*Made with ❤️ by [Donghwan Kim (David Kim)](https://github.com/ainsuotain)*
License: [Apache 2.0](LICENSE)

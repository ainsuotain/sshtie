#!/usr/bin/env bash
# Auto-typed demo script — runs inside the asciinema recording
# Called by demo_record.sh

SSHTIE="${1:-./sshtie}"

# Helper: type text with a natural delay then press enter
run() {
  sleep 0.8
  echo "$ $*"
  sleep 0.5
}

clear

# ── 1. help ──────────────────────────────────────────────────────
run "sshtie --help"
$SSHTIE --help
sleep 2

# ── 2. add ───────────────────────────────────────────────────────
echo ""
run "sshtie add"
printf "homeserver\n192.168.1.100\nalice\n\n\n\n\n" | $SSHTIE add
sleep 2

# ── 3. list ──────────────────────────────────────────────────────
echo ""
run "sshtie list"
$SSHTIE list
sleep 2

# ── 4. doctor (offline — shows diagnostics flow) ─────────────────
echo ""
run "sshtie doctor homeserver"
$SSHTIE doctor homeserver 2>&1 || true
sleep 2

# ── 5. remove ────────────────────────────────────────────────────
echo ""
run "sshtie remove homeserver"
$SSHTIE remove homeserver
sleep 2

echo ""
echo "→ sshtie: SSH + mosh + tmux, unified."
sleep 3

#!/usr/bin/env bash
# sshtie demo recording script
# Usage: bash demo_record.sh
# Output: demo.cast → demo.gif

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSHTIE="$SCRIPT_DIR/sshtie"
CAST="$SCRIPT_DIR/demo.cast"
GIF="$SCRIPT_DIR/demo.gif"

# Build binary first
echo "→ Building sshtie..."
go build -o "$SSHTIE" "$SCRIPT_DIR"

echo "→ Starting recording... (follow the prompts)"
echo "   Press Ctrl+D or type 'exit' when done."
echo ""

asciinema rec "$CAST" \
  --title "sshtie demo" \
  --cols 90 \
  --rows 24 \
  --command "bash $SCRIPT_DIR/demo_script.sh $SSHTIE"

echo ""
echo "→ Converting to GIF..."
agg \
  --font-size 14 \
  --theme dracula \
  --speed 1.5 \
  "$CAST" "$GIF"

echo ""
echo "✅ Done! → $GIF"
echo "→ git add demo.gif && git commit -m 'docs: add demo gif' && git push"

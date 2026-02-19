#!/usr/bin/env bash
# build-app.sh — assemble sshtie-menubar.app for macOS
# Usage: bash build/darwin/build-app.sh [VERSION]
set -euo pipefail

VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
DIST="dist"
APP="$DIST/sshtie-menubar.app"
MACOS="$APP/Contents/MacOS"
RES="$APP/Contents/Resources"

echo "Building sshtie-menubar $VERSION …"
echo ""

# ── 1. Build menubar binary (universal) ──────────────────────────────────────
mkdir -p "$DIST"

printf "  %-24s" "darwin/arm64"
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
    go build -ldflags "-s -w -X main.version=$VERSION" \
    -o "$DIST/sshtie-menubar-arm64" ./menubar/
echo "✅"

printf "  %-24s" "darwin/amd64"
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
    go build -ldflags "-s -w -X main.version=$VERSION" \
    -o "$DIST/sshtie-menubar-amd64" ./menubar/
echo "✅"

printf "  %-24s" "lipo (universal)"
lipo -create \
    -output "$DIST/sshtie-menubar" \
    "$DIST/sshtie-menubar-arm64" \
    "$DIST/sshtie-menubar-amd64"
rm "$DIST/sshtie-menubar-arm64" "$DIST/sshtie-menubar-amd64"
echo "✅"

# ── 2. Build CLI binary (placed beside menubar in bundle) ────────────────────
printf "  %-24s" "sshtie CLI (arm64)"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
    go build -ldflags "-s -w -X main.version=$VERSION" \
    -o "$DIST/sshtie-cli-arm64" .
echo "✅"

printf "  %-24s" "sshtie CLI (amd64)"
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
    go build -ldflags "-s -w -X main.version=$VERSION" \
    -o "$DIST/sshtie-cli-amd64" .
echo "✅"

printf "  %-24s" "lipo CLI (universal)"
lipo -create \
    -output "$DIST/sshtie-cli" \
    "$DIST/sshtie-cli-arm64" \
    "$DIST/sshtie-cli-amd64"
rm "$DIST/sshtie-cli-arm64" "$DIST/sshtie-cli-amd64"
echo "✅"

# ── 3. Assemble .app bundle ───────────────────────────────────────────────────
echo ""
printf "  %-24s" "assembling .app"
rm -rf "$APP"
mkdir -p "$MACOS" "$RES"

cp "$DIST/sshtie-menubar" "$MACOS/sshtie-menubar"
cp "$DIST/sshtie-cli"     "$MACOS/sshtie"          # CLI next to menubar binary
rm "$DIST/sshtie-menubar" "$DIST/sshtie-cli"

sed "s/VERSION_PLACEHOLDER/$VERSION/g" build/darwin/Info.plist \
    > "$APP/Contents/Info.plist"
echo "✅"

# ── 4. Ad-hoc code sign (removes Gatekeeper "damaged app" warning) ───────────
printf "  %-24s" "codesign (ad-hoc)"
codesign --deep --force --sign - "$APP" 2>/dev/null && echo "✅" || echo "⚠  skipped"

# ── 5. Zip for distribution ───────────────────────────────────────────────────
printf "  %-24s" "zipping"
cd "$DIST"
zip -qr "sshtie-menubar-darwin-universal.zip" "sshtie-menubar.app"
cd - > /dev/null
echo "✅"

echo ""
echo "✅  Done!"
echo "   App : $APP"
echo "   Zip : $DIST/sshtie-menubar-darwin-universal.zip"
echo ""
echo "To run:"
echo "   open $APP"
echo ""
echo "First launch: if macOS blocks it, run:"
echo "   xattr -d com.apple.quarantine $APP"

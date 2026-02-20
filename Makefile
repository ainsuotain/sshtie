# ── sshtie release build ──────────────────────────────────────────────────────
#
#   make build                      cross-compile all 4 platforms → dist/
#   make release VERSION=v0.1.0     gh release create with all artifacts
#   make update-formula VERSION=v0.1.0  update sha256 in Formula/sshtie.rb
#   make clean                      remove dist/
#
# Prerequisites: go, zip, gh (GitHub CLI)

BINARY  := sshtie
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)
DIST    := dist

.PHONY: all build menubar menubar-run release update-formula clean

all: build

# ── Cross-compile ─────────────────────────────────────────────────────────────

build:
	@mkdir -p $(DIST)
	@echo "Building $(BINARY) $(VERSION) for all platforms..."
	@echo ""

	@printf "  %-18s" "linux/amd64"
	@GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" \
	    -o $(DIST)/$(BINARY) .
	@tar -czf $(DIST)/$(BINARY)-linux-amd64.tar.gz \
	    -C $(DIST) $(BINARY)
	@rm $(DIST)/$(BINARY)
	@echo "✅  $(DIST)/$(BINARY)-linux-amd64.tar.gz"

	@printf "  %-18s" "darwin/amd64 (Intel)"
	@GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" \
	    -o $(DIST)/$(BINARY) .
	@tar -czf $(DIST)/$(BINARY)-mac-intel.tar.gz \
	    -C $(DIST) $(BINARY)
	@rm $(DIST)/$(BINARY)
	@echo "✅  $(DIST)/$(BINARY)-mac-intel.tar.gz"

	@printf "  %-18s" "darwin/arm64 (Apple Silicon)"
	@GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" \
	    -o $(DIST)/$(BINARY) .
	@tar -czf $(DIST)/$(BINARY)-mac-apple-silicon.tar.gz \
	    -C $(DIST) $(BINARY)
	@rm $(DIST)/$(BINARY)
	@echo "✅  $(DIST)/$(BINARY)-mac-apple-silicon.tar.gz"

	@printf "  %-18s" "windows/amd64"
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" \
	    -o $(DIST)/$(BINARY).exe .
	@zip -j $(DIST)/$(BINARY)-windows-amd64.zip \
	    $(DIST)/$(BINARY).exe > /dev/null
	@rm $(DIST)/$(BINARY).exe
	@echo "✅  $(DIST)/$(BINARY)-windows-amd64.zip"

	@echo ""
	@echo "Generating SHA256SUMS..."
	@cd $(DIST) && \
	    (sha256sum *.tar.gz *.zip 2>/dev/null \
	     || shasum -a 256 *.tar.gz *.zip) > SHA256SUMS
	@echo ""
	@echo "✅  Done!  $(DIST)/"
	@echo ""
	@cat $(DIST)/SHA256SUMS

# ── macOS Menu-bar App (darwin only, requires CGO) ───────────────────────────

menubar:
	@bash build/darwin/build-app.sh $(VERSION)

menubar-run: menubar
	@open dist/sshtie-menubar.app

tray-windows:
	@echo "Building sshtie-tray.exe (windows/amd64)..."
	@mkdir -p $(DIST)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
	    go build -ldflags "$(LDFLAGS) -H windowsgui" \
	    -o $(DIST)/sshtie-tray.exe ./menubar/
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
	    go build -ldflags "$(LDFLAGS)" \
	    -o $(DIST)/sshtie.exe .
	@zip -j $(DIST)/sshtie-tray-windows-amd64.zip \
	    $(DIST)/sshtie-tray.exe $(DIST)/sshtie.exe > /dev/null
	@rm $(DIST)/sshtie-tray.exe $(DIST)/sshtie.exe
	@echo "✅  $(DIST)/sshtie-tray-windows-amd64.zip"

# ── GitHub Release ────────────────────────────────────────────────────────────
# Usage: make release VERSION=v0.1.0
#        (runs build first if dist/ is empty)

release: build
	@test -n "$(VERSION)" || \
	    (echo "Error: specify a version: make release VERSION=v0.1.0"; exit 1)
	@echo ""
	@echo "Creating GitHub release $(VERSION)..."
	gh release create $(VERSION) \
	    --title "sshtie $(VERSION)" \
	    --generate-notes \
	    $(DIST)/$(BINARY)-linux-amd64.tar.gz \
	    $(DIST)/$(BINARY)-mac-apple-silicon.tar.gz \
	    $(DIST)/$(BINARY)-mac-intel.tar.gz \
	    $(DIST)/$(BINARY)-windows-amd64.zip \
	    $(DIST)/$(BINARY)-menubar-darwin-universal.zip \
	    $(DIST)/$(BINARY)-tray-windows-amd64.zip \
	    $(DIST)/SHA256SUMS
	@echo ""
	@echo "✅  Release $(VERSION) published!"
	@echo "→   https://github.com/ainsuotain/sshtie/releases/tag/$(VERSION)"
	@echo "→   Next: make update-formula VERSION=$(VERSION)"

# ── Update Homebrew formula ───────────────────────────────────────────────────
# Usage: make update-formula VERSION=v0.1.0
#        Fetches the release tarball SHA256 and patches Formula/sshtie.rb.
#        Then copy Formula/sshtie.rb → homebrew-sshtie repo and push.

update-formula:
	@test -n "$(VERSION)" || \
	    (echo "Error: make update-formula VERSION=v0.1.0"; exit 1)
	@echo "Fetching SHA256 for $(VERSION) tarball..."
	@SHA256=$$(curl -sL \
	    "https://github.com/ainsuotain/sshtie/archive/refs/tags/$(VERSION).tar.gz" \
	    | shasum -a 256 | cut -d' ' -f1) && \
	test -n "$$SHA256" || \
	    (echo "Failed — has $(VERSION) been released yet?"; exit 1) && \
	perl -i -pe "s|sha256 \".*\"|sha256 \"$$SHA256\"|" Formula/sshtie.rb && \
	perl -i -pe 's|/tags/v[\d.]+\.tar\.gz|/tags/$(VERSION).tar.gz|' \
	    Formula/sshtie.rb && \
	echo "" && \
	echo "✅  Formula/sshtie.rb updated" && \
	echo "   version: $(VERSION)" && \
	echo "   sha256 : $$SHA256" && \
	echo "" && \
	echo "→  Copy Formula/sshtie.rb to ainsuotain/homebrew-sshtie and push."

# ── Clean ─────────────────────────────────────────────────────────────────────

clean:
	@rm -rf $(DIST)
	@echo "✅  Cleaned dist/"

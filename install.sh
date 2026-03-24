#!/usr/bin/env bash

# BLUEPRINT SDLC Installer
# Works via curl pipe (remote) or from cloned repo (local).
# Repo: github.com/skaisser/blueprint

set -euo pipefail

BLUEPRINT_VERSION="1.0.0"
GITHUB_REPO="skaisser/blueprint"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
GITHUB_RAW="https://raw.githubusercontent.com/${GITHUB_REPO}/main"
TS="$(date +%Y%m%d-%H%M%S)"

# ============================================================================
# UNINSTALL
# ============================================================================

if [[ "${1:-}" == "--uninstall" ]]; then
    echo ""
    echo "  Uninstalling BLUEPRINT SDLC..."
    echo ""

    # Remove blueprint bin, templates, statusline
    if [ -d "$HOME/.blueprint" ]; then
        rm -rf "$HOME/.blueprint"
        echo "  Removed ~/.blueprint/"
    fi

    # Remove blueprint skills from ~/.claude/skills/
    BLUEPRINT_SKILLS=(
        address-pr backlog batch-flow bp-branch bp-commit bp-context
        bp-push bp-ship bp-status bp-tdd-review bp-test complete
        finish flow flow-auto flow-auto-wt hotfix plan plan-approved
        plan-check plan-review pr quick resume review skill-creator start
    )
    for skill in "${BLUEPRINT_SKILLS[@]}"; do
        if [ -d "$HOME/.claude/skills/$skill" ]; then
            rm -rf "$HOME/.claude/skills/$skill"
        fi
    done
    echo "  Removed Blueprint skills from ~/.claude/skills/"

    # Remove audit hook from settings.json
    if [ -f "$HOME/.claude/settings.json" ]; then
        if grep -q 'blueprint audit' "$HOME/.claude/settings.json" 2>/dev/null; then
            cp "$HOME/.claude/settings.json" "$HOME/.claude/settings.json.bak.${TS}"
            sed -i.tmp '/blueprint audit/d' "$HOME/.claude/settings.json"
            rm -f "$HOME/.claude/settings.json.tmp"
            echo "  Removed audit hook from ~/.claude/settings.json (backup: settings.json.bak.${TS})"
        fi
        # Remove statusLine referencing blueprint
        if grep -q 'blueprint/statusline' "$HOME/.claude/settings.json" 2>/dev/null; then
            sed -i.tmp '/statusLine/,/}/d' "$HOME/.claude/settings.json"
            rm -f "$HOME/.claude/settings.json.tmp"
            echo "  Removed statusLine from ~/.claude/settings.json"
        fi
    fi

    # Remove PATH entry from shell RC
    for rc in "$HOME/.zshrc" "$HOME/.bashrc"; do
        if [ -f "$rc" ] && grep -q '.blueprint/bin' "$rc" 2>/dev/null; then
            cp "$rc" "${rc}.bak.${TS}"
            grep -v '.blueprint/bin' "$rc" > "${rc}.tmp" && mv "${rc}.tmp" "$rc"
            echo "  Removed PATH entry from $(basename "$rc") (backup: $(basename "$rc").bak.${TS})"
        fi
    done

    echo ""
    echo "  BLUEPRINT has been uninstalled."
    echo "  Your Claude Code settings and MCP servers were preserved (except Blueprint entries)."
    echo ""
    exit 0
fi

# ============================================================================
# PHASE 1: CORE FRAMEWORK
# ============================================================================

# Detect install mode: local (from cloned repo) vs remote (curl pipe)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd || pwd)"
if [ -d "$SCRIPT_DIR/skills" ] && [ -d "$SCRIPT_DIR/cli" ]; then
    INSTALL_MODE="local"
else
    INSTALL_MODE="remote"
fi

# Platform detection
detect_platform() {
    local os arch
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$os" in
        darwin) OS="darwin" ;;
        linux)  OS="linux" ;;
        *)
            echo "  Unsupported OS: $os"
            exit 1
            ;;
    esac

    case "$arch" in
        x86_64)  ARCH="amd64" ;;
        amd64)   ARCH="amd64" ;;
        arm64)   ARCH="arm64" ;;
        aarch64) ARCH="arm64" ;;
        *)
            echo "  Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

detect_platform

# Shell RC detection
detect_shell_rc() {
    if [ -f "$HOME/.zshrc" ]; then
        SHELL_RC="$HOME/.zshrc"
    elif [ -f "$HOME/.bashrc" ]; then
        SHELL_RC="$HOME/.bashrc"
    else
        SHELL_RC=""
    fi
}

detect_shell_rc

# gum bootstrap
HAS_GUM=false
bootstrap_gum() {
    if command -v gum &>/dev/null; then
        HAS_GUM=true
        return
    fi

    # Try to install gum
    if [ "$OS" = "darwin" ]; then
        if command -v brew &>/dev/null; then
            echo "  Installing gum via Homebrew..."
            brew install gum 2>/dev/null && HAS_GUM=true || true
        fi
    else
        # Linux: download gum binary
        local gum_version="0.14.5"
        local gum_url="https://github.com/charmbracelet/gum/releases/download/v${gum_version}/gum_${gum_version}_Linux_x86_64.tar.gz"
        if [ "$ARCH" = "arm64" ]; then
            gum_url="https://github.com/charmbracelet/gum/releases/download/v${gum_version}/gum_${gum_version}_Linux_aarch64.tar.gz"
        fi
        local tmp_dir
        tmp_dir="$(mktemp -d)"
        if curl -fsSL "$gum_url" -o "$tmp_dir/gum.tar.gz" 2>/dev/null; then
            tar -xzf "$tmp_dir/gum.tar.gz" -C "$tmp_dir" 2>/dev/null
            if [ -f "$tmp_dir/gum" ]; then
                mkdir -p "$HOME/.local/bin"
                mv "$tmp_dir/gum" "$HOME/.local/bin/gum"
                chmod +x "$HOME/.local/bin/gum"
                export PATH="$HOME/.local/bin:$PATH"
                HAS_GUM=true
            fi
        fi
        rm -rf "$tmp_dir"
    fi

    if [ "$HAS_GUM" = false ]; then
        return
    fi
}

bootstrap_gum

# ── ANSI 256-color gradient printer ──────────────────────────────────────────
_gradient_print() {
    local text="$1"
    local theme="$2"

    local colors
    IFS=' ' read -ra colors <<< "$theme"
    local nc="${#colors[@]}"

    local i=0
    while IFS= read -r line; do
        local color="${colors[$((i % nc))]}"
        printf "\033[38;5;%sm%s\033[0m\n" "$color" "$line"
        (( i++ ))
    done <<< "$text"
}

# ── Block-art logo ──────────────────────────────────────────────────────────
_logo_lines() {
    printf '%s\n' \
        "" \
        "  ██████╗ ██╗     ██╗   ██╗███████╗██████╗ ██████╗ ██╗███╗   ██╗████████╗" \
        "  ██╔══██╗██║     ██║   ██║██╔════╝██╔══██╗██╔══██╗██║████╗  ██║╚══██╔══╝" \
        "  ██████╔╝██║     ██║   ██║█████╗  ██████╔╝██████╔╝██║██╔██╗ ██║   ██║   " \
        "  ██╔══██╗██║     ██║   ██║██╔══╝  ██╔═══╝ ██╔══██╗██║██║╚██╗██║   ██║   " \
        "  ██████╔╝███████╗╚██████╔╝███████╗██║     ██║  ██║██║██║ ╚████║   ██║   " \
        "  ╚═════╝ ╚══════╝ ╚═════╝ ╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝   ╚═╝   " \
        ""
}

# ── Styled helpers (gum or plain) ───────────────────────────────────────────
ok()   { if [ "$HAS_GUM" = true ]; then gum style --foreground="#50fa7b" "  ✓  $*"; else echo "  ✓  $*"; fi; }
info() { if [ "$HAS_GUM" = true ]; then gum style --foreground="#8be9fd" "  ℹ  $*"; else echo "  ℹ  $*"; fi; }
warn() { if [ "$HAS_GUM" = true ]; then gum style --foreground="#ffb86c" "  ⚠  $*"; else echo "  ⚠  $*"; fi; }
err()  { if [ "$HAS_GUM" = true ]; then gum style --foreground="#ff5555" "  ✗  $*"; else echo "  ✗  $*"; fi; }
step() {
    echo ""
    if [ "$HAS_GUM" = true ]; then
        gum style --foreground="#bd93f9" --bold "━━  $*"
    else
        echo "━━  $*"
    fi
}

print_step() { step "$2"; }
print_success() { ok "$@"; }
print_warn() { warn "$@"; }
print_error() { err "$@"; }

spin_exec() {
    local title="$1"
    shift
    if [ "$HAS_GUM" = true ] && [ -t 1 ]; then
        gum spin --spinner dot --title "$title" -- "$@"
    else
        echo "  $title"
        "$@"
    fi
}

# ── Banner ───────────────────────────────────────────────────────────────────
show_banner() {
    local mode_label="local (cloned repo)"
    [ "$INSTALL_MODE" = "remote" ] && mode_label="remote (curl pipe)"

    clear 2>/dev/null || true
    echo ""

    if [ "$HAS_GUM" = true ]; then
        # Gradient themes (8 colors per theme — one per logo line incl. blanks)
        local themes=(
            "39  39  33  27  21  57  93  129"   # Blueprint blue → purple
            "51  51  45  39  33  27  21  21"    # Ice: cyan → blue
            "201 201 165 129 93  57  21  21"    # Cyberpunk: magenta → blue
            "99  99  141 183 219 225 231 231"   # Lavender rise
            "87  87  81  75  69  63  57  51"    # Deep ocean fade
        )

        local chosen="${themes[$((RANDOM % ${#themes[@]}))]}"
        _gradient_print "$(_logo_lines)" "$chosen"

        echo ""
        gum style \
            --border rounded \
            --border-foreground="#bd93f9" \
            --padding "1 4" \
            --margin "0 2" \
            "$(gum style --foreground='#ff79c6' --bold 'BLUEPRINT SDLC') $(gum style --foreground='#6272a4' "v${BLUEPRINT_VERSION}")" \
            "" \
            "$(gum style --foreground='#8be9fd' 'A complete software development lifecycle for Claude Code.')" \
            "$(gum style --foreground='#8be9fd' 'Up to 200x faster. Fully automatic. Stack-agnostic.')" \
            "" \
            "$(gum style --foreground='#6272a4' "${OS}/${ARCH}  ·  ${mode_label}  ·  gum: enabled")"

        echo ""
        printf "\033[38;5;141m  27 skills  ·  15 audit rules  ·  Go CLI  ·  zero paid dependencies\033[0m\n"
    else
        echo ""
        echo "  ██████╗ ██╗     ██╗   ██╗███████╗██████╗ ██████╗ ██╗███╗   ██╗████████╗"
        echo "  ██╔══██╗██║     ██║   ██║██╔════╝██╔══██╗██╔══██╗██║████╗  ██║╚══██╔══╝"
        echo "  ██████╔╝██║     ██║   ██║█████╗  ██████╔╝██████╔╝██║██╔██╗ ██║   ██║   "
        echo "  ██╔══██╗██║     ██║   ██║██╔══╝  ██╔═══╝ ██╔══██╗██║██║╚██╗██║   ██║   "
        echo "  ██████╔╝███████╗╚██████╔╝███████╗██║     ██║  ██║██║██║ ╚████║   ██║   "
        echo "  ╚═════╝ ╚══════╝ ╚═════╝ ╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝   ╚═╝   "
        echo ""
        echo "  BLUEPRINT SDLC v${BLUEPRINT_VERSION}"
        echo "  ${OS}/${ARCH} · ${mode_label}"
        echo ""
        echo "  27 skills · 15 audit rules · Go CLI · zero paid dependencies"
    fi

    echo ""
    if [ "$HAS_GUM" = true ]; then
        printf "\033[38;5;83m  Press Enter to begin →\033[0m\n"
        echo ""
        read -r -s -n 1
    fi
}

show_banner

# Create directories (silent — no output between screens)
mkdir -p "$HOME/.blueprint/bin"
mkdir -p "$HOME/.blueprint/templates"
mkdir -p "$HOME/.claude/skills"

# ============================================================================
# PHASE 2: INTERACTIVE COMPONENT MENU
# ============================================================================

# Components: required ones are always true
COMP_BINARY=true        # always
COMP_SKILLS=true        # always
COMP_AUDIT=true         # audit hook (15 rules)
COMP_STATUSLINE=true    # status line script
COMP_PERMISSIONS=true   # git, sequential-thinking permissions
COMP_AGENTTEAMS=true    # experimental agent teams env var
COMP_PLUGINS=true       # ralph-loop, skill-creator, playground
COMP_MCP=true           # Context7 + Sequential Thinking MCP servers
COMP_GITHOOKS=true      # git hook templates
COMP_GHACTIONS=true     # GitHub Action templates

select_components() {
    clear 2>/dev/null || true
    echo ""

    if [ "$HAS_GUM" = true ]; then
        gum style \
            --border rounded \
            --border-foreground="#bd93f9" \
            --padding "1 2" \
            --margin "0 2" \
            "$(gum style --foreground='#bd93f9' --bold '📦  Choose what to install')" \
            "" \
            "$(gum style --foreground='#50fa7b' 'Always installed (core):')" \
            "  Blueprint CLI binary  ·  27 SDLC skills" \
            "" \
            "$(gum style --foreground='#8be9fd' 'Default: Yes for all. Press n to skip any.')"

        echo ""

        # Simple gum confirm per component — reliable, no multi-select bugs
        gum confirm --default=true --prompt.foreground="#50fa7b" "  Audit hook — 15 rules (recommended)?"       || COMP_AUDIT=false
        gum confirm --default=true --prompt.foreground="#50fa7b" "  Status line — context/git/duration bar (recommended)?" || COMP_STATUSLINE=false
        gum confirm --default=true --prompt.foreground="#8be9fd" "  Permissions — git push / sequential-thinking?" || COMP_PERMISSIONS=false
        gum confirm --default=true --prompt.foreground="#8be9fd" "  Agent Teams — experimental env flag?"        || COMP_AGENTTEAMS=false
        gum confirm --default=true --prompt.foreground="#8be9fd" "  Plugins — ralph-loop / skill-creator / playground?" || COMP_PLUGINS=false
        gum confirm --default=true --prompt.foreground="#8be9fd" "  MCP servers — Context7 / Sequential Thinking?" || COMP_MCP=false
        gum confirm --default=true --prompt.foreground="#8be9fd" "  Git hook templates?"                         || COMP_GITHOOKS=false
        gum confirm --default=true --prompt.foreground="#8be9fd" "  GitHub Action templates?"                    || COMP_GHACTIONS=false
    else
        echo ""
        echo "  ━━  Choose what to install"
        echo ""
        echo "  Always installed (core):"
        echo "    [x] Blueprint CLI binary"
        echo "    [x] 27 SDLC skills"
        echo ""
        echo "  Optional (default: yes):"
        echo ""

        read -r -p "    Audit hook — 15 rules (recommended)? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && COMP_AUDIT=false

        read -r -p "    Status line — context/git/duration bar (recommended)? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && COMP_STATUSLINE=false

        read -r -p "    Permissions — git push / sequential-thinking? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && COMP_PERMISSIONS=false

        read -r -p "    Agent Teams — experimental env flag? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && COMP_AGENTTEAMS=false

        read -r -p "    Plugins — ralph-loop / skill-creator / playground? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && COMP_PLUGINS=false

        read -r -p "    MCP servers — Context7 / Sequential Thinking? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && COMP_MCP=false

        read -r -p "    Git hook templates? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && COMP_GITHOOKS=false

        read -r -p "    GitHub Action templates? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && COMP_GHACTIONS=false
    fi
}

# Only show menu if running interactively
if [ -t 0 ]; then
    select_components
fi

# Install summary
show_summary() {
    if [ "$HAS_GUM" = true ]; then
        clear 2>/dev/null || true
        echo ""

        # Build selected/skipped lists
        local selected_lines=()
        local skipped_lines=()

        selected_lines+=("  • Blueprint CLI binary  →  ~/.blueprint/bin/blueprint")
        selected_lines+=("  • 27 SDLC skills  →  ~/.claude/skills/")

        [ "$COMP_AUDIT" = true ]       && selected_lines+=("  • Audit hook (15 rules)  →  settings.json")       || skipped_lines+=("  • Audit hook (15 rules)")
        [ "$COMP_STATUSLINE" = true ]  && selected_lines+=("  • Status line  →  ~/.blueprint/statusline.sh")    || skipped_lines+=("  • Status line")
        [ "$COMP_PERMISSIONS" = true ] && selected_lines+=("  • Permissions  →  settings.json")                  || skipped_lines+=("  • Permissions")
        [ "$COMP_AGENTTEAMS" = true ]  && selected_lines+=("  • Agent Teams  →  settings.json")                  || skipped_lines+=("  • Agent Teams")
        [ "$COMP_PLUGINS" = true ]     && selected_lines+=("  • Plugins  →  settings.json")                      || skipped_lines+=("  • Plugins")
        [ "$COMP_MCP" = true ]         && selected_lines+=("  • MCP servers  →  mcp.json")                       || skipped_lines+=("  • MCP servers")
        [ "$COMP_GITHOOKS" = true ]    && selected_lines+=("  • Git hook templates  →  ~/.blueprint/templates/") || skipped_lines+=("  • Git hook templates")
        [ "$COMP_GHACTIONS" = true ]   && selected_lines+=("  • GitHub Action templates  →  ~/.blueprint/templates/") || skipped_lines+=("  • GitHub Action templates")

        local summary_args=(
            "$(gum style --foreground='#bd93f9' --bold '📦  Installation Summary')"
            ""
            "$(gum style --foreground='#6272a4' "${OS}/${ARCH}  ·  ${INSTALL_MODE} mode")"
            ""
            "$(gum style --foreground='#50fa7b' 'Selected:')"
        )
        for line in "${selected_lines[@]}"; do
            summary_args+=("$line")
        done

        if [ "${#skipped_lines[@]}" -gt 0 ]; then
            summary_args+=("")
            summary_args+=("$(gum style --foreground='#6272a4' 'Skipped:')")
            for line in "${skipped_lines[@]}"; do
                summary_args+=("$(gum style --foreground='#6272a4' "$line")")
            done
        fi

        gum style \
            --border rounded \
            --border-foreground="#bd93f9" \
            --padding "1 2" \
            --margin "0 2" \
            "${summary_args[@]}"
    else
        echo ""
        echo "  ━━  Installation Summary"
        echo ""
        echo "  Platform: ${OS}/${ARCH}  ·  Mode: ${INSTALL_MODE}"
        echo ""
        echo "  Selected:"
        echo "    [x] Blueprint CLI binary       → ~/.blueprint/bin/blueprint"
        echo "    [x] 27 SDLC skills             → ~/.claude/skills/"
        [ "$COMP_AUDIT" = true ]       && echo "    [x] Audit hook (15 rules)"       || echo "    [ ] Audit hook (15 rules)"
        [ "$COMP_STATUSLINE" = true ]  && echo "    [x] Status line"                  || echo "    [ ] Status line"
        [ "$COMP_PERMISSIONS" = true ] && echo "    [x] Permissions"                  || echo "    [ ] Permissions"
        [ "$COMP_AGENTTEAMS" = true ]  && echo "    [x] Agent Teams"                  || echo "    [ ] Agent Teams"
        [ "$COMP_PLUGINS" = true ]     && echo "    [x] Plugins"                      || echo "    [ ] Plugins"
        [ "$COMP_MCP" = true ]         && echo "    [x] MCP servers"                  || echo "    [ ] MCP servers"
        [ "$COMP_GITHOOKS" = true ]    && echo "    [x] Git hook templates"           || echo "    [ ] Git hook templates"
        [ "$COMP_GHACTIONS" = true ]   && echo "    [x] GitHub Action templates"      || echo "    [ ] GitHub Action templates"
        echo ""
    fi
}

show_summary

# Confirm
confirm_install() {
    echo ""
    if [ "$HAS_GUM" = true ]; then
        if ! gum confirm \
            --prompt.foreground="#bd93f9" \
            --selected.background="#bd93f9" \
            --selected.foreground="#000000" \
            "  Proceed with installation?"; then
            echo ""
            warn "Installation cancelled."
            exit 0
        fi
    else
        read -r -p "  Proceed with installation? [Y/n] " ans
        [[ "$ans" =~ ^[Nn] ]] && { echo "  Cancelled."; exit 0; }
    fi
}

if [ -t 0 ]; then
    confirm_install
fi

# Clear for install output
clear 2>/dev/null || true
echo ""
if [ "$HAS_GUM" = true ]; then
    gum style --foreground="#bd93f9" --bold --margin "0 2" "Installing your setup..."
else
    echo "  Installing your setup..."
fi
echo ""

# ============================================================================
# PHASE 3: BINARY DOWNLOAD & INSTALL
# ============================================================================

install_binary() {
    echo ""
    print_step "🔧" "Installing Blueprint CLI binary..."

    local binary_name="blueprint-${OS}-${ARCH}"
    local target="$HOME/.blueprint/bin/blueprint"

    if [ "$INSTALL_MODE" = "local" ]; then
        # Local: copy from repo
        local src="$SCRIPT_DIR/cli/${binary_name}"
        if [ -f "$src" ]; then
            cp -f "$src" "$target"
            chmod +x "$target"
            print_success "Copied binary from repo (${binary_name})"
        else
            print_warn "Binary not found at ${src}"
            print_warn "Available: $(ls "$SCRIPT_DIR"/cli/blueprint-* 2>/dev/null | xargs -I{} basename {} | tr '\n' ' ')"
            print_warn "Build it: cd cli && make build-all"
            return 1
        fi
    else
        # Remote: download from GitHub Releases
        print_step "⬇" "Fetching latest release from GitHub..."

        local release_json
        release_json="$(curl -fsSL "$GITHUB_API" 2>/dev/null)" || {
            print_error "Failed to fetch release info from GitHub"
            return 1
        }

        # Extract tag_name (version)
        local tag
        tag="$(echo "$release_json" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//')"

        # Extract download URL for our platform binary
        local download_url
        download_url="$(echo "$release_json" | grep '"browser_download_url"' | grep "$binary_name" | head -1 | sed 's/.*"browser_download_url": *"//;s/".*//')"

        if [ -z "$download_url" ]; then
            print_error "No binary found for ${binary_name} in release ${tag:-unknown}"
            return 1
        fi

        print_step "⬇" "Downloading ${binary_name} (${tag})..."

        if [ "$HAS_GUM" = true ]; then
            gum spin --spinner dot --title "Downloading ${binary_name}..." -- \
                curl -fSL -o "$target" "$download_url"
        else
            curl -fSL --progress-bar -o "$target" "$download_url"
        fi

        chmod +x "$target"
        print_success "Downloaded ${binary_name} (${tag})"
    fi

    # Verify
    if [ -x "$target" ]; then
        local ver
        ver="$("$target" --version 2>/dev/null || echo "unknown")"
        print_success "Binary installed: ${ver}"
    fi

    # Add to PATH
    if [ -n "$SHELL_RC" ]; then
        if ! grep -q '.blueprint/bin' "$SHELL_RC" 2>/dev/null; then
            echo '' >> "$SHELL_RC"
            echo '# BLUEPRINT SDLC' >> "$SHELL_RC"
            echo 'export PATH="$HOME/.blueprint/bin:$PATH"' >> "$SHELL_RC"
            print_success "Added ~/.blueprint/bin to PATH in $(basename "$SHELL_RC")"
        else
            print_success "PATH already configured in $(basename "$SHELL_RC")"
        fi
    else
        print_warn "No .zshrc or .bashrc found — add manually: export PATH=\"\$HOME/.blueprint/bin:\$PATH\""
    fi

    # Make available in current session
    export PATH="$HOME/.blueprint/bin:$PATH"
}

install_binary || true

# ============================================================================
# PHASE 4: SKILLS & TEMPLATES INSTALLATION
# ============================================================================

install_skills() {
    echo ""
    print_step "📦" "Installing SDLC skills..."

    local src_dir=""
    local tmp_dir=""

    if [ "$INSTALL_MODE" = "local" ]; then
        src_dir="$SCRIPT_DIR/skills"
    else
        # Remote: clone repo to temp dir
        tmp_dir="$(mktemp -d)"
        print_step "⬇" "Cloning skills from GitHub..."
        if [ "$HAS_GUM" = true ]; then
            gum spin --spinner dot --title "Cloning repository..." -- \
                git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$tmp_dir" 2>/dev/null
        else
            git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$tmp_dir" 2>/dev/null
        fi
        src_dir="$tmp_dir/skills"
    fi

    if [ ! -d "$src_dir" ]; then
        print_error "Skills directory not found"
        [ -n "$tmp_dir" ] && rm -rf "$tmp_dir"
        return 1
    fi

    # Count and install skills
    local total=0
    local count=0
    for item in "$src_dir/"*/; do
        [ -d "$item" ] && total=$((total + 1))
    done

    for item in "$src_dir/"*/; do
        if [ -d "$item" ]; then
            count=$((count + 1))
            local name
            name="$(basename "$item")"
            mkdir -p "$HOME/.claude/skills/$name"
            if [ "$HAS_GUM" = true ]; then
                gum spin --spinner dot --title "Installing skill ${count}/${total}: ${name}" -- \
                    cp -rf "$item"* "$HOME/.claude/skills/$name/"
            else
                printf "  Installing skill %d/%d: %s\r" "$count" "$total" "$name"
                cp -rf "$item"* "$HOME/.claude/skills/$name/"
            fi
        fi
    done

    if [ "$HAS_GUM" = false ]; then
        echo ""
    fi

    print_success "Installed ${count} skills to ~/.claude/skills/"

    # Clean up temp dir
    [ -n "$tmp_dir" ] && rm -rf "$tmp_dir"
}

install_templates() {
    echo ""
    print_step "📦" "Installing templates..."

    local src_dir=""
    local tmp_dir=""

    if [ "$INSTALL_MODE" = "local" ]; then
        src_dir="$SCRIPT_DIR/templates"
    else
        # For remote mode, clone to temp
        tmp_dir="$(mktemp -d)"
        git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$tmp_dir" 2>/dev/null || true
        src_dir="$tmp_dir/templates"
    fi

    if [ ! -d "$src_dir" ]; then
        print_warn "Templates directory not found"
        [ -n "$tmp_dir" ] && rm -rf "$tmp_dir"
        return 0
    fi

    # Copy all templates (including dotfiles like .githooks/ and .github/)
    cp -rf "$src_dir/"* "$HOME/.blueprint/templates/" 2>/dev/null || true
    cp -rf "$src_dir/".[!.]* "$HOME/.blueprint/templates/" 2>/dev/null || true

    # Preserve executable permissions on git hooks
    if [ "$COMP_GITHOOKS" = true ]; then
        chmod +x "$HOME/.blueprint/templates/.githooks/commit-msg" 2>/dev/null || true
        chmod +x "$HOME/.blueprint/templates/.githooks/pre-push" 2>/dev/null || true
        print_success "Git hook templates installed"
    fi

    if [ "$COMP_GHACTIONS" = true ]; then
        print_success "GitHub Action templates installed"
    fi

    # Copy CLAUDE.md template
    if [ -f "$src_dir/CLAUDE.md" ]; then
        print_success "CLAUDE.md template installed"
    fi

    # Copy blueprint scaffold
    if [ -d "$src_dir/blueprint" ]; then
        print_success "Blueprint workspace scaffold installed"
    fi

    print_success "Templates installed to ~/.blueprint/templates/"

    [ -n "$tmp_dir" ] && rm -rf "$tmp_dir"
}

install_skills
install_templates

# ============================================================================
# PHASE 5: CLAUDE CODE INTEGRATION
# ============================================================================

# Helper: get a source file path (local or remote temp clone)
# For remote mode, we reuse a single temp clone for config files
REMOTE_CLONE_DIR=""
get_config_source() {
    local filename="$1"
    if [ "$INSTALL_MODE" = "local" ]; then
        echo "$SCRIPT_DIR/config/$filename"
        return
    fi
    # Remote: clone once, reuse
    if [ -z "$REMOTE_CLONE_DIR" ]; then
        REMOTE_CLONE_DIR="$(mktemp -d)"
        git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$REMOTE_CLONE_DIR" 2>/dev/null || true
    fi
    echo "$REMOTE_CLONE_DIR/config/$filename"
}

# --- Status line installation ---
install_statusline() {
    if [ "$COMP_STATUSLINE" = false ]; then
        return 0
    fi

    echo ""
    print_step "📊" "Installing status line..."

    local src
    src="$(get_config_source "statusline.sh")"

    if [ -f "$src" ]; then
        cp -f "$src" "$HOME/.blueprint/statusline.sh"
        chmod +x "$HOME/.blueprint/statusline.sh"
        print_success "Status line installed to ~/.blueprint/statusline.sh"
    else
        # Write a minimal version if source not found
        print_warn "statusline.sh not found in source, skipping"
        COMP_STATUSLINE=false
    fi
}

install_statusline

# --- Settings.json merge (granular, section by section) ---
register_settings() {
    local settings_file="$HOME/.claude/settings.json"
    local needs_settings=false

    # Check if any settings-related component is selected
    if [ "$COMP_AUDIT" = true ] || [ "$COMP_STATUSLINE" = true ] || \
       [ "$COMP_PERMISSIONS" = true ] || [ "$COMP_AGENTTEAMS" = true ] || \
       [ "$COMP_PLUGINS" = true ]; then
        needs_settings=true
    fi

    if [ "$needs_settings" = false ]; then
        return 0
    fi

    echo ""
    print_step "🔗" "Registering Claude Code settings..."

    # Back up existing settings
    if [ -f "$settings_file" ]; then
        cp "$settings_file" "${settings_file}.bak.${TS}"
        print_step "💾" "Backed up settings.json"
    fi

    # If no settings.json exists, build from selected sections
    if [ ! -f "$settings_file" ]; then
        echo '{' > "$settings_file"
        local first_section=true

        # env (Agent Teams)
        if [ "$COMP_AGENTTEAMS" = true ]; then
            [ "$first_section" = false ] && echo ',' >> "$settings_file"
            first_section=false
            cat >> "$settings_file" << 'ENVEOF'
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
ENVEOF
        fi

        # permissions
        if [ "$COMP_PERMISSIONS" = true ]; then
            [ "$first_section" = false ] && echo ',' >> "$settings_file"
            first_section=false
            cat >> "$settings_file" << 'PERMEOF'
  "permissions": {
    "allow": [
      "Bash(git add:*)",
      "Bash(git push:*)",
      "mcp__sequential-thinking__sequentialthinking",
      "/commit"
    ],
    "defaultMode": "default"
  }
PERMEOF
        fi

        # hooks (Audit)
        if [ "$COMP_AUDIT" = true ]; then
            [ "$first_section" = false ] && echo ',' >> "$settings_file"
            first_section=false
            cat >> "$settings_file" << 'HOOKEOF'
  "hooks": {
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "$HOME/.blueprint/bin/blueprint audit"
      }]
    }]
  }
HOOKEOF
        fi

        # statusLine
        if [ "$COMP_STATUSLINE" = true ]; then
            [ "$first_section" = false ] && echo ',' >> "$settings_file"
            first_section=false
            cat >> "$settings_file" << 'SLEOF'
  "statusLine": {
    "type": "command",
    "command": "$HOME/.blueprint/statusline.sh"
  }
SLEOF
        fi

        # enabledPlugins
        if [ "$COMP_PLUGINS" = true ]; then
            [ "$first_section" = false ] && echo ',' >> "$settings_file"
            first_section=false
            cat >> "$settings_file" << 'PLUGEOF'
  "enabledPlugins": {
    "ralph-loop@claude-plugins-official": true,
    "skill-creator@claude-plugins-official": true,
    "playground@claude-plugins-official": true
  }
PLUGEOF
        fi

        echo '}' >> "$settings_file"
        print_success "Created settings.json with selected sections"
        return 0
    fi

    # --- Settings.json exists: merge each section individually ---

    # Audit hook
    if [ "$COMP_AUDIT" = true ]; then
        if grep -q 'blueprint audit' "$settings_file" 2>/dev/null; then
            print_success "Audit hook already registered"
        elif grep -q 'claude-cli audit\|audit\.py' "$settings_file" 2>/dev/null; then
            print_success "Existing audit hook detected — keeping current hook"
        elif grep -q '"PreToolUse"' "$settings_file" 2>/dev/null; then
            # PreToolUse exists — append our matcher block
            sed -i.tmp 's/\("PreToolUse": *\[\)/\1\n      {"matcher":".*","hooks":[{"type":"command","command":"$HOME\/.blueprint\/bin\/blueprint audit"}]},/' "$settings_file"
            rm -f "${settings_file}.tmp"
            print_success "Audit hook appended to existing PreToolUse"
        elif grep -q '"hooks"' "$settings_file" 2>/dev/null; then
            # hooks object exists but no PreToolUse
            sed -i.tmp 's/\("hooks": *{\)/\1\n    "PreToolUse": [{"matcher":".*","hooks":[{"type":"command","command":"$HOME\/.blueprint\/bin\/blueprint audit"}]}],/' "$settings_file"
            rm -f "${settings_file}.tmp"
            print_success "Added PreToolUse with audit hook"
        else
            # No hooks at all — add before closing brace
            local tmp_file
            tmp_file="$(mktemp)"
            sed '$ s/}//' "$settings_file" > "$tmp_file"
            cat >> "$tmp_file" << 'AHEOF'
  ,"hooks": {
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "$HOME/.blueprint/bin/blueprint audit"
      }]
    }]
  }
}
AHEOF
            mv "$tmp_file" "$settings_file"
            print_success "Added hooks section with audit hook"
        fi
    fi

    # Status line
    if [ "$COMP_STATUSLINE" = true ]; then
        if grep -q 'statusLine' "$settings_file" 2>/dev/null; then
            print_success "Status line already registered"
        else
            local tmp_file
            tmp_file="$(mktemp)"
            sed '$ s/}//' "$settings_file" > "$tmp_file"
            cat >> "$tmp_file" << 'STEOF'
  ,"statusLine": {
    "type": "command",
    "command": "$HOME/.blueprint/statusline.sh"
  }
}
STEOF
            mv "$tmp_file" "$settings_file"
            print_success "Status line registered"
        fi
    fi

    # Permissions
    if [ "$COMP_PERMISSIONS" = true ]; then
        if grep -q '"permissions"' "$settings_file" 2>/dev/null; then
            # Check for specific entries and add missing ones
            local added=false
            for perm in 'Bash(git add:*)' 'Bash(git push:*)' 'mcp__sequential-thinking__sequentialthinking' '/commit'; do
                if ! grep -q "$perm" "$settings_file" 2>/dev/null; then
                    added=true
                fi
            done
            if [ "$added" = true ]; then
                print_warn "Permissions section exists — verify Blueprint permissions are present"
            else
                print_success "Permissions already configured"
            fi
        else
            local tmp_file
            tmp_file="$(mktemp)"
            sed '$ s/}//' "$settings_file" > "$tmp_file"
            cat >> "$tmp_file" << 'PMEOF'
  ,"permissions": {
    "allow": [
      "Bash(git add:*)",
      "Bash(git push:*)",
      "mcp__sequential-thinking__sequentialthinking",
      "/commit"
    ],
    "defaultMode": "default"
  }
}
PMEOF
            mv "$tmp_file" "$settings_file"
            print_success "Permissions added"
        fi
    fi

    # Agent Teams env
    if [ "$COMP_AGENTTEAMS" = true ]; then
        if grep -q 'CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS' "$settings_file" 2>/dev/null; then
            print_success "Agent Teams env already set"
        elif grep -q '"env"' "$settings_file" 2>/dev/null; then
            # env exists — add our var
            sed -i.tmp 's/\("env": *{\)/\1\n    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",/' "$settings_file"
            rm -f "${settings_file}.tmp"
            print_success "Agent Teams env added to existing env section"
        else
            local tmp_file
            tmp_file="$(mktemp)"
            sed '$ s/}//' "$settings_file" > "$tmp_file"
            cat >> "$tmp_file" << 'ATEOF'
  ,"env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
ATEOF
            mv "$tmp_file" "$settings_file"
            print_success "Agent Teams env section added"
        fi
    fi

    # Plugins
    if [ "$COMP_PLUGINS" = true ]; then
        if grep -q '"enabledPlugins"' "$settings_file" 2>/dev/null; then
            # Check if our plugins are there
            local missing=false
            for plug in 'ralph-loop' 'skill-creator' 'playground'; do
                if ! grep -q "$plug" "$settings_file" 2>/dev/null; then
                    missing=true
                fi
            done
            if [ "$missing" = true ]; then
                # Add missing plugins to existing section
                for plug_entry in '"ralph-loop@claude-plugins-official": true' '"skill-creator@claude-plugins-official": true' '"playground@claude-plugins-official": true'; do
                    local plug_name="${plug_entry%%@*}"
                    plug_name="${plug_name#\"}"
                    if ! grep -q "$plug_name" "$settings_file" 2>/dev/null; then
                        sed -i.tmp "s/\(\"enabledPlugins\": *{\)/\1\n    ${plug_entry},/" "$settings_file"
                        rm -f "${settings_file}.tmp"
                    fi
                done
                print_success "Missing plugins added to enabledPlugins"
            else
                print_success "Plugins already configured"
            fi
        else
            local tmp_file
            tmp_file="$(mktemp)"
            sed '$ s/}//' "$settings_file" > "$tmp_file"
            cat >> "$tmp_file" << 'PLEOF'
  ,"enabledPlugins": {
    "ralph-loop@claude-plugins-official": true,
    "skill-creator@claude-plugins-official": true,
    "playground@claude-plugins-official": true
  }
}
PLEOF
            mv "$tmp_file" "$settings_file"
            print_success "Plugins section added"
        fi
    fi
}

register_settings

# --- MCP servers ---
register_mcp() {
    if [ "$COMP_MCP" = false ]; then
        return 0
    fi

    echo ""
    print_step "🔗" "Registering MCP servers..."

    local mcp_file="$HOME/.claude/mcp.json"

    # Back up existing mcp.json
    if [ -f "$mcp_file" ]; then
        cp "$mcp_file" "${mcp_file}.bak.${TS}"
        print_step "💾" "Backed up mcp.json"
    fi

    # Try claude CLI first for each server
    local c7_done=false
    local st_done=false

    if command -v claude &>/dev/null; then
        # Context7
        if [ -f "$mcp_file" ] && grep -q '"context7"' "$mcp_file" 2>/dev/null; then
            c7_done=true
            print_success "Context7 MCP already registered"
        elif claude mcp add-json context7 '{"command":"npx","args":["-y","@upstash/context7-mcp@latest"]}' 2>/dev/null; then
            c7_done=true
            print_success "Context7 MCP registered via claude CLI"
        fi

        # Sequential Thinking
        if [ -f "$mcp_file" ] && grep -q '"sequential-thinking"' "$mcp_file" 2>/dev/null; then
            st_done=true
            print_success "Sequential Thinking MCP already registered"
        elif claude mcp add-json sequential-thinking '{"command":"npx","args":["-y","@anthropic-ai/mcp-sequential-thinking"]}' 2>/dev/null; then
            st_done=true
            print_success "Sequential Thinking MCP registered via claude CLI"
        fi
    fi

    # Fallback: pure bash
    if [ "$c7_done" = true ] && [ "$st_done" = true ]; then
        return 0
    fi

    if [ ! -f "$mcp_file" ]; then
        # No mcp.json — use template or write directly
        local config_src
        config_src="$(get_config_source "mcp.json")"
        if [ -f "$config_src" ]; then
            cp "$config_src" "$mcp_file"
        else
            cat > "$mcp_file" << 'MCP_EOF'
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    },
    "sequential-thinking": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-sequential-thinking"]
    }
  }
}
MCP_EOF
        fi
        chmod 600 "$mcp_file" 2>/dev/null || true
        print_success "Created mcp.json with Context7 and Sequential Thinking"
        return 0
    fi

    # mcp.json exists — merge missing servers
    if [ "$c7_done" = false ] && ! grep -q '"context7"' "$mcp_file" 2>/dev/null; then
        local c7_block='"context7": {"command": "npx", "args": ["-y", "@upstash/context7-mcp@latest"]}'
        if grep -q '"mcpServers"' "$mcp_file" 2>/dev/null; then
            sed -i.tmp "s/\"mcpServers\": *{/\"mcpServers\": {\n    ${c7_block},/" "$mcp_file"
            rm -f "${mcp_file}.tmp"
            print_success "Context7 MCP added to mcp.json"
        fi
    fi

    if [ "$st_done" = false ] && ! grep -q '"sequential-thinking"' "$mcp_file" 2>/dev/null; then
        local st_block='"sequential-thinking": {"command": "npx", "args": ["-y", "@anthropic-ai/mcp-sequential-thinking"]}'
        if grep -q '"mcpServers"' "$mcp_file" 2>/dev/null; then
            sed -i.tmp "s/\"mcpServers\": *{/\"mcpServers\": {\n    ${st_block},/" "$mcp_file"
            rm -f "${mcp_file}.tmp"
            print_success "Sequential Thinking MCP added to mcp.json"
        fi
    fi

    chmod 600 "$mcp_file" 2>/dev/null || true
}

register_mcp

# Clean up remote clone temp dir
[ -n "$REMOTE_CLONE_DIR" ] && rm -rf "$REMOTE_CLONE_DIR"

# ============================================================================
# PHASE 6: POST-INSTALL SUMMARY
# ============================================================================

# Count installed items
skill_count="$(find "$HOME/.claude/skills" -maxdepth 2 -name 'SKILL.md' 2>/dev/null | wc -l | tr -d ' ')"
binary_ver="$(command -v blueprint &>/dev/null && blueprint --version 2>/dev/null || echo 'installed (restart shell)')"
has_hooks="$([ -f "$HOME/.blueprint/templates/.githooks/commit-msg" ] && echo '✓' || echo '–')"
has_actions="$([ -d "$HOME/.blueprint/templates/.github/workflows" ] && echo '✓' || echo '–')"
has_audit="$(grep -q 'audit' "$HOME/.claude/settings.json" 2>/dev/null && echo '✓' || echo '–')"
has_statusline="$([ -f "$HOME/.blueprint/statusline.sh" ] && echo '✓' || echo '–')"
has_mcp="$(grep -q 'context7' "$HOME/.claude/mcp.json" 2>/dev/null && echo '✓' || echo '–')"
has_permissions="$(grep -q 'git add' "$HOME/.claude/settings.json" 2>/dev/null && echo '✓' || echo '–')"
has_plugins="$(grep -q 'enabledPlugins' "$HOME/.claude/settings.json" 2>/dev/null && echo '✓' || echo '–')"
has_teams="$(grep -q 'AGENT_TEAMS' "$HOME/.claude/settings.json" 2>/dev/null && echo '✓' || echo '–')"

clear 2>/dev/null || true
echo ""

if [ "$HAS_GUM" = true ]; then
    gum style \
        --border double \
        --border-foreground="#50fa7b" \
        --padding "1 4" \
        --margin "0 2" \
        "$(gum style --foreground='#50fa7b' --bold '  Installation complete!')" \
        "" \
        "$(gum style --foreground='#ffffff' --bold 'Installed:')" \
        "" \
        "$(gum style --foreground='#8be9fd' "  CLI binary:     ${binary_ver}")" \
        "$(gum style --foreground='#8be9fd' "  Skills:         ${skill_count} installed")" \
        "$(gum style --foreground='#8be9fd' "  Audit hook:     ${has_audit}    Status line:  ${has_statusline}")" \
        "$(gum style --foreground='#8be9fd' "  Permissions:    ${has_permissions}    Agent Teams:  ${has_teams}")" \
        "$(gum style --foreground='#8be9fd' "  Plugins:        ${has_plugins}    MCP servers:  ${has_mcp}")" \
        "$(gum style --foreground='#8be9fd' "  Git hooks:      ${has_hooks}    GH Actions:   ${has_actions}")" \
        "" \
        "$(gum style --foreground='#ffffff' --bold 'Next steps:')" \
        "" \
        "$(gum style --foreground='#8be9fd' "  1.  Restart your terminal (or: source ${SHELL_RC:-~/.zshrc})")" \
        "$(gum style --foreground='#8be9fd' '  2.  Verify:  blueprint --version')" \
        "$(gum style --foreground='#8be9fd' '  3.  Go to a project:  cd ~/your-project')" \
        "$(gum style --foreground='#8be9fd' '  4.  Initialize:  /start')" \
        "" \
        "$(gum style --foreground='#bd93f9' "  github.com/${GITHUB_REPO}")"

    echo ""

    if [ -t 0 ]; then
        if gum confirm \
            --prompt.foreground="#ffb86c" \
            --selected.background="#ffb86c" \
            --selected.foreground="#000000" \
            "  ⭐  Star BLUEPRINT on GitHub!"; then
            open "https://github.com/${GITHUB_REPO}" 2>/dev/null || true
        fi
    fi
else
    echo "  ✅  Installation complete!"
    echo ""
    echo "  CLI binary:     ${binary_ver}"
    echo "  Skills:         ${skill_count} installed"
    echo "  Audit hook:     ${has_audit}    Status line:  ${has_statusline}"
    echo "  Permissions:    ${has_permissions}    Agent Teams:  ${has_teams}"
    echo "  Plugins:        ${has_plugins}    MCP servers:  ${has_mcp}"
    echo "  Git hooks:      ${has_hooks}    GH Actions:   ${has_actions}"
    echo ""
    echo "  Next steps:"
    echo "    1. Restart your terminal (or: source ${SHELL_RC:-~/.zshrc})"
    echo "    2. Verify:  blueprint --version"
    echo "    3. Go to a project:  cd ~/your-project"
    echo "    4. Initialize:  /start"
    echo ""
    echo "  github.com/${GITHUB_REPO}"
fi

echo ""
echo "  Happy shipping! 🚀"
echo ""

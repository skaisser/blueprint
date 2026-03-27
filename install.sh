#!/usr/bin/env bash

# BLUEPRINT SDLC Installer
# Works via curl pipe (remote) or from cloned repo (local).
# Repo: github.com/skaisser/blueprint

set -euo pipefail

BLUEPRINT_VERSION="2.0.1"
GITHUB_REPO="skaisser/blueprint"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
TS="$(date +%Y%m%d-%H%M%S)"

# ============================================================================
# UNINSTALL
# ============================================================================

if [[ "${1:-}" == "--uninstall" ]]; then
    echo ""
    echo "  Uninstalling BLUEPRINT SDLC..."
    echo ""

    [ -d "$HOME/.blueprint" ] && rm -rf "$HOME/.blueprint" && echo "  Removed ~/.blueprint/"

    BLUEPRINT_SKILLS=(
        address-pr backlog batch-flow bp-branch bp-commit bp-context
        bp-push bp-ship bp-status bp-tdd-review bp-test complete
        finish flow flow-auto flow-auto-wt hotfix plan plan-approved
        plan-check plan-review pr quick resume review skill-creator start
    )
    for skill in "${BLUEPRINT_SKILLS[@]}"; do
        [ -d "$HOME/.claude/skills/$skill" ] && rm -rf "$HOME/.claude/skills/$skill"
    done
    echo "  Removed Blueprint skills from ~/.claude/skills/"

    if [ -f "$HOME/.claude/settings.json" ]; then
        cp "$HOME/.claude/settings.json" "$HOME/.claude/settings.json.bak.${TS}"
        if grep -q 'blueprint audit' "$HOME/.claude/settings.json" 2>/dev/null; then
            sed -i.tmp '/blueprint audit/d' "$HOME/.claude/settings.json"
            rm -f "$HOME/.claude/settings.json.tmp"
            echo "  Removed audit hook from settings.json"
        fi
        if grep -q 'blueprint/statusline' "$HOME/.claude/settings.json" 2>/dev/null; then
            sed -i.tmp '/statusLine/,/}/d' "$HOME/.claude/settings.json"
            rm -f "$HOME/.claude/settings.json.tmp"
            echo "  Removed statusLine from settings.json"
        fi
    fi

    for rc in "$HOME/.zshrc" "$HOME/.bashrc"; do
        if [ -f "$rc" ] && grep -q '.blueprint/bin' "$rc" 2>/dev/null; then
            cp "$rc" "${rc}.bak.${TS}"
            grep -v '.blueprint/bin' "$rc" > "${rc}.tmp" && mv "${rc}.tmp" "$rc"
            echo "  Removed PATH entry from $(basename "$rc")"
        fi
    done

    echo ""
    echo "  BLUEPRINT has been uninstalled."
    echo ""
    exit 0
fi

# ============================================================================
# PLATFORM & TOOLS
# ============================================================================

# Platform detection
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$OS" in
    darwin) ;; linux) ;; *) echo "  Unsupported OS: $OS"; exit 1 ;;
esac
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    amd64)   ;;
    arm64)   ;;
    aarch64) ARCH="arm64" ;;
    *) echo "  Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Shell RC
SHELL_RC=""
[ -f "$HOME/.zshrc" ] && SHELL_RC="$HOME/.zshrc"
[ -z "$SHELL_RC" ] && [ -f "$HOME/.bashrc" ] && SHELL_RC="$HOME/.bashrc"

# Install mode: local (from cloned repo) vs remote (curl pipe)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd || pwd)"
if [ -d "$SCRIPT_DIR/skills" ] && [ -d "$SCRIPT_DIR/cli" ]; then
    INSTALL_MODE="local"
else
    INSTALL_MODE="remote"
fi

# Detect gum (optional — for banner only)
HAS_GUM=false
command -v gum &>/dev/null && HAS_GUM=true

# ============================================================================
# BANNER
# ============================================================================

_gradient_print() {
    local text="$1" theme="$2"
    local colors
    IFS=' ' read -ra colors <<< "$theme"
    local nc="${#colors[@]}" i=0
    while IFS= read -r line; do
        printf "\033[38;5;%sm%s\033[0m\n" "${colors[$((i % nc))]}" "$line"
        (( i++ ))
    done <<< "$text"
}

show_banner() {
    clear 2>/dev/null || true
    echo ""

    local logo
    logo="$(printf '%s\n' \
        "" \
        "  ██████╗ ██╗     ██╗   ██╗███████╗██████╗ ██████╗ ██╗███╗   ██╗████████╗" \
        "  ██╔══██╗██║     ██║   ██║██╔════╝██╔══██╗██╔══██╗██║████╗  ██║╚══██╔══╝" \
        "  ██████╔╝██║     ██║   ██║█████╗  ██████╔╝██████╔╝██║██╔██╗ ██║   ██║   " \
        "  ██╔══██╗██║     ██║   ██║██╔══╝  ██╔═══╝ ██╔══██╗██║██║╚██╗██║   ██║   " \
        "  ██████╔╝███████╗╚██████╔╝███████╗██║     ██║  ██║██║██║ ╚████║   ██║   " \
        "  ╚═════╝ ╚══════╝ ╚═════╝ ╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝   ╚═╝   " \
        "")"

    if [ "$HAS_GUM" = true ]; then
        local themes=(
            "39  39  33  27  21  57  93  129"
            "51  51  45  39  33  27  21  21"
            "201 201 165 129 93  57  21  21"
            "99  99  141 183 219 225 231 231"
            "87  87  81  75  69  63  57  51"
        )
        _gradient_print "$logo" "${themes[$((RANDOM % ${#themes[@]}))]}"
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
            "$(gum style --foreground='#6272a4' "${OS}/${ARCH}  ·  ${INSTALL_MODE} mode")"
        echo ""
        printf "\033[38;5;141m  27 skills  ·  15 audit rules  ·  Go CLI  ·  zero paid dependencies\033[0m\n"
    else
        echo "$logo"
        echo ""
        echo "  BLUEPRINT SDLC v${BLUEPRINT_VERSION}"
        echo "  ${OS}/${ARCH} · ${INSTALL_MODE} mode"
        echo ""
        echo "  27 skills · 15 audit rules · Go CLI · zero paid dependencies"
    fi
    echo ""
}

show_banner

# ============================================================================
# HELPERS
# ============================================================================

ok()   { echo "  ✅ $*"; }
info() { echo "  ℹ️  $*"; }
warn() { echo "  ⚠️  $*"; }
err()  { echo "  ❌ $*"; }
step() { echo ""; echo "━━  $*"; }

# ============================================================================
# CREATE DIRECTORIES
# ============================================================================

mkdir -p "$HOME/.blueprint/bin"
mkdir -p "$HOME/.blueprint/templates"
mkdir -p "$HOME/.claude/skills"

# ============================================================================
# SOURCE: local repo or remote clone
# ============================================================================

SRC_DIR="$SCRIPT_DIR"
CLEANUP_DIR=""

if [ "$INSTALL_MODE" = "remote" ]; then
    step "Cloning repository..."
    CLEANUP_DIR="$(mktemp -d)"
    git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$CLEANUP_DIR" 2>/dev/null
    SRC_DIR="$CLEANUP_DIR"
    ok "Repository cloned"
fi

# ============================================================================
# 1. BINARY
# ============================================================================

step "Installing Blueprint CLI binary..."

BINARY_NAME="blueprint-${OS}-${ARCH}"

if [ "$INSTALL_MODE" = "local" ] && [ -f "$SRC_DIR/cli/${BINARY_NAME}" ]; then
    cp -f "$SRC_DIR/cli/${BINARY_NAME}" "$HOME/.blueprint/bin/blueprint"
    chmod +x "$HOME/.blueprint/bin/blueprint"
    ok "Copied binary from repo (${BINARY_NAME})"
else
    # Remote: download from GitHub Releases
    RELEASE_JSON="$(curl -fsSL "$GITHUB_API" 2>/dev/null)" || { err "Failed to fetch release info"; }

    if [ -n "${RELEASE_JSON:-}" ]; then
        TAG="$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//')"
        DOWNLOAD_URL="$(echo "$RELEASE_JSON" | grep '"browser_download_url"' | grep "$BINARY_NAME" | head -1 | sed 's/.*"browser_download_url": *"//;s/".*//')"

        if [ -n "${DOWNLOAD_URL:-}" ]; then
            curl -fSL --progress-bar -o "$HOME/.blueprint/bin/blueprint" "$DOWNLOAD_URL"
            chmod +x "$HOME/.blueprint/bin/blueprint"
            ok "Downloaded ${BINARY_NAME} (${TAG})"
        else
            warn "No binary found for ${BINARY_NAME} in release ${TAG:-unknown}"
        fi
    fi
fi

# Verify binary
if [ -x "$HOME/.blueprint/bin/blueprint" ]; then
    VER="$("$HOME/.blueprint/bin/blueprint" --version 2>/dev/null || echo "installed")"
    ok "Binary: ${VER}"
fi

# Add to PATH
if [ -n "$SHELL_RC" ] && ! grep -q '.blueprint/bin' "$SHELL_RC" 2>/dev/null; then
    echo '' >> "$SHELL_RC"
    echo '# BLUEPRINT SDLC' >> "$SHELL_RC"
    echo 'export PATH="$HOME/.blueprint/bin:$PATH"' >> "$SHELL_RC"
    ok "Added ~/.blueprint/bin to PATH in $(basename "$SHELL_RC")"
fi
export PATH="$HOME/.blueprint/bin:$PATH"

# ============================================================================
# 2. SKILLS
# ============================================================================

step "Installing skills..."

if [ -d "$SRC_DIR/skills" ]; then
    COUNT=0
    TOTAL=$(find "$SRC_DIR/skills" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')

    for item in "$SRC_DIR/skills/"*/; do
        [ -d "$item" ] || continue
        NAME="$(basename "$item")"
        COUNT=$((COUNT + 1))
        mkdir -p "$HOME/.claude/skills/$NAME"
        cp -f "$item"* "$HOME/.claude/skills/$NAME/" 2>/dev/null || true
        # Also copy subdirectories (references, etc.)
        for subdir in "$item"*/; do
            [ -d "$subdir" ] || continue
            SUBNAME="$(basename "$subdir")"
            mkdir -p "$HOME/.claude/skills/$NAME/$SUBNAME"
            cp -f "$subdir"* "$HOME/.claude/skills/$NAME/$SUBNAME/" 2>/dev/null || true
        done
        printf "  Installing skill %d/%d: %s\r" "$COUNT" "$TOTAL" "$NAME"
    done
    echo ""
    ok "Installed ${COUNT} skills to ~/.claude/skills/"
else
    err "Skills directory not found at $SRC_DIR/skills"
fi

# ============================================================================
# 3. TEMPLATES
# ============================================================================

step "Installing templates..."

if [ -d "$SRC_DIR/templates" ]; then
    # Copy regular files and directories
    cp -rf "$SRC_DIR/templates/"* "$HOME/.blueprint/templates/" 2>/dev/null || true
    # Copy dotfiles (.githooks, .github)
    cp -rf "$SRC_DIR/templates/".[!.]* "$HOME/.blueprint/templates/" 2>/dev/null || true

    # Preserve executable permissions on git hooks
    chmod +x "$HOME/.blueprint/templates/.githooks/commit-msg" 2>/dev/null || true
    chmod +x "$HOME/.blueprint/templates/.githooks/pre-push" 2>/dev/null || true

    ok "Templates installed (git hooks, GitHub Actions, CLAUDE.md, scaffold)"
else
    warn "Templates directory not found"
fi

# ============================================================================
# 4. STATUS LINE
# ============================================================================

step "Installing status line..."

if [ -f "$SRC_DIR/config/statusline.sh" ]; then
    cp -f "$SRC_DIR/config/statusline.sh" "$HOME/.blueprint/statusline.sh"
    chmod +x "$HOME/.blueprint/statusline.sh"
    ok "Status line installed"
else
    warn "statusline.sh not found"
fi

# ============================================================================
# 5. SETTINGS.JSON (granular merge — never overwrite)
# ============================================================================

step "Configuring Claude Code settings..."

SETTINGS="$HOME/.claude/settings.json"

# Back up existing
if [ -f "$SETTINGS" ]; then
    cp "$SETTINGS" "${SETTINGS}.bak.${TS}"
    info "Backed up settings.json"
fi

# Create if missing
if [ ! -f "$SETTINGS" ]; then
    cat > "$SETTINGS" << 'EOF'
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  },
  "permissions": {
    "allow": [
      "Bash(git add:*)",
      "Bash(git push:*)",
      "mcp__sequential-thinking__sequentialthinking",
      "/commit"
    ],
    "defaultMode": "default"
  },
  "hooks": {
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "$HOME/.blueprint/bin/blueprint audit"
      }]
    }]
  },
  "statusLine": {
    "type": "command",
    "command": "$HOME/.blueprint/statusline.sh"
  },
  "enabledPlugins": {
    "ralph-loop@claude-plugins-official": true,
    "skill-creator@claude-plugins-official": true,
    "playground@claude-plugins-official": true
  }
}
EOF
    ok "Created settings.json"
else
    # Merge individual sections into existing settings

    # Audit hook
    if grep -q 'blueprint audit' "$SETTINGS" 2>/dev/null; then
        ok "Audit hook already registered"
    elif ! grep -q 'audit' "$SETTINGS" 2>/dev/null; then
        TMP="$(mktemp)"
        sed '$ s/}//' "$SETTINGS" > "$TMP"
        [ "$(tail -c 1 "$TMP")" != "," ] && echo ',' >> "$TMP" || true
        cat >> "$TMP" << 'EOF'
  "hooks": {
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "$HOME/.blueprint/bin/blueprint audit"
      }]
    }]
  }
}
EOF
        mv "$TMP" "$SETTINGS"
        ok "Audit hook registered"
    else
        info "Existing audit hook detected — keeping yours"
    fi

    # Status line
    if ! grep -q 'statusLine' "$SETTINGS" 2>/dev/null; then
        TMP="$(mktemp)"
        sed '$ s/}//' "$SETTINGS" > "$TMP"
        cat >> "$TMP" << 'EOF'
  ,"statusLine": {
    "type": "command",
    "command": "$HOME/.blueprint/statusline.sh"
  }
}
EOF
        mv "$TMP" "$SETTINGS"
        ok "Status line registered"
    else
        ok "Status line already registered"
    fi

    # Agent Teams env
    if ! grep -q 'CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS' "$SETTINGS" 2>/dev/null; then
        if grep -q '"env"' "$SETTINGS" 2>/dev/null; then
            sed -i.tmp 's/"env": *{/"env": {\n    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",/' "$SETTINGS"
            rm -f "${SETTINGS}.tmp"
        else
            TMP="$(mktemp)"
            sed '$ s/}//' "$SETTINGS" > "$TMP"
            cat >> "$TMP" << 'EOF'
  ,"env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
EOF
            mv "$TMP" "$SETTINGS"
        fi
        ok "Agent Teams env set"
    else
        ok "Agent Teams already configured"
    fi

    # Permissions
    if ! grep -q '"permissions"' "$SETTINGS" 2>/dev/null; then
        TMP="$(mktemp)"
        sed '$ s/}//' "$SETTINGS" > "$TMP"
        cat >> "$TMP" << 'EOF'
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
EOF
        mv "$TMP" "$SETTINGS"
        ok "Permissions added"
    else
        ok "Permissions already configured"
    fi

    # Plugins
    if ! grep -q '"enabledPlugins"' "$SETTINGS" 2>/dev/null; then
        TMP="$(mktemp)"
        sed '$ s/}//' "$SETTINGS" > "$TMP"
        cat >> "$TMP" << 'EOF'
  ,"enabledPlugins": {
    "ralph-loop@claude-plugins-official": true,
    "skill-creator@claude-plugins-official": true,
    "playground@claude-plugins-official": true
  }
}
EOF
        mv "$TMP" "$SETTINGS"
        ok "Plugins enabled"
    else
        ok "Plugins already configured"
    fi
fi

# ============================================================================
# 6. MCP SERVERS
# ============================================================================

step "Registering MCP servers..."

MCP_FILE="$HOME/.claude/mcp.json"

# Back up existing
if [ -f "$MCP_FILE" ]; then
    cp "$MCP_FILE" "${MCP_FILE}.bak.${TS}"
fi

# Try claude CLI first (cleanest approach)
C7_DONE=false
ST_DONE=false

if command -v claude &>/dev/null; then
    if [ -f "$MCP_FILE" ] && grep -q '"context7"' "$MCP_FILE" 2>/dev/null; then
        C7_DONE=true
        ok "Context7 already registered"
    elif claude mcp add-json context7 '{"command":"npx","args":["-y","@upstash/context7-mcp@latest"]}' 2>/dev/null; then
        C7_DONE=true
        ok "Context7 registered via claude CLI"
    fi

    if [ -f "$MCP_FILE" ] && grep -q '"sequential-thinking"' "$MCP_FILE" 2>/dev/null; then
        ST_DONE=true
        ok "Sequential Thinking already registered"
    elif claude mcp add-json sequential-thinking '{"command":"npx","args":["-y","@anthropic-ai/mcp-sequential-thinking"]}' 2>/dev/null; then
        ST_DONE=true
        ok "Sequential Thinking registered via claude CLI"
    fi
fi

# Fallback: write mcp.json directly
if [ "$C7_DONE" = false ] || [ "$ST_DONE" = false ]; then
    if [ ! -f "$MCP_FILE" ]; then
        cat > "$MCP_FILE" << 'EOF'
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
EOF
        chmod 600 "$MCP_FILE"
        ok "Created mcp.json with Context7 + Sequential Thinking"
    else
        # Add missing servers
        if [ "$C7_DONE" = false ] && ! grep -q '"context7"' "$MCP_FILE" 2>/dev/null; then
            sed -i.tmp 's/"mcpServers": *{/"mcpServers": {\n    "context7": {"command": "npx", "args": ["-y", "@upstash\/context7-mcp@latest"]},/' "$MCP_FILE"
            rm -f "${MCP_FILE}.tmp"
            ok "Context7 added to mcp.json"
        fi
        if [ "$ST_DONE" = false ] && ! grep -q '"sequential-thinking"' "$MCP_FILE" 2>/dev/null; then
            sed -i.tmp 's/"mcpServers": *{/"mcpServers": {\n    "sequential-thinking": {"command": "npx", "args": ["-y", "@anthropic-ai\/mcp-sequential-thinking"]},/' "$MCP_FILE"
            rm -f "${MCP_FILE}.tmp"
            ok "Sequential Thinking added to mcp.json"
        fi
        chmod 600 "$MCP_FILE"
    fi
fi

# ============================================================================
# CLEANUP
# ============================================================================

[ -n "$CLEANUP_DIR" ] && rm -rf "$CLEANUP_DIR"

# ============================================================================
# SUMMARY
# ============================================================================

SKILL_COUNT=$(find "$HOME/.claude/skills" -maxdepth 2 -name 'SKILL.md' 2>/dev/null | wc -l | tr -d ' ')
BINARY_VER="$(command -v blueprint &>/dev/null && blueprint --version 2>/dev/null || echo 'installed (restart shell)')"

echo ""
echo ""

if [ "$HAS_GUM" = true ]; then
    gum style \
        --border double \
        --border-foreground="#50fa7b" \
        --padding "1 4" \
        --margin "0 2" \
        "$(gum style --foreground='#50fa7b' --bold '  Installation complete!')" \
        "" \
        "$(gum style --foreground='#8be9fd' "  CLI:      ${BINARY_VER}")" \
        "$(gum style --foreground='#8be9fd' "  Skills:   ${SKILL_COUNT} installed")" \
        "" \
        "$(gum style --foreground='#ffffff' --bold 'Next steps:')" \
        "" \
        "$(gum style --foreground='#8be9fd' "  1.  Restart your terminal (or: source ${SHELL_RC:-~/.zshrc})")" \
        "$(gum style --foreground='#8be9fd' '  2.  Verify:  blueprint --version')" \
        "$(gum style --foreground='#8be9fd' '  3.  Go to a project:  cd ~/your-project')" \
        "$(gum style --foreground='#8be9fd' '  4.  Initialize:  /start')" \
        "" \
        "$(gum style --foreground='#bd93f9' "  github.com/${GITHUB_REPO}")"
else
    echo "  ✅  Installation complete!"
    echo ""
    echo "  CLI:      ${BINARY_VER}"
    echo "  Skills:   ${SKILL_COUNT} installed"
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

# Star prompt
if [ -t 0 ]; then
    if [ "$HAS_GUM" = true ]; then
        if gum confirm \
            --prompt.foreground="#ffb86c" \
            --selected.background="#ffb86c" \
            --selected.foreground="#000000" \
            "  ⭐  Star BLUEPRINT on GitHub?"; then
            open "https://github.com/${GITHUB_REPO}" 2>/dev/null || \
                xdg-open "https://github.com/${GITHUB_REPO}" 2>/dev/null || true
        fi
    else
        read -r -p "  ⭐  Star BLUEPRINT on GitHub? [Y/n] " ans
        if [[ ! "$ans" =~ ^[Nn] ]]; then
            open "https://github.com/${GITHUB_REPO}" 2>/dev/null || \
                xdg-open "https://github.com/${GITHUB_REPO}" 2>/dev/null || true
        fi
    fi
fi

echo ""
echo "  Happy shipping! 🚀"
echo ""

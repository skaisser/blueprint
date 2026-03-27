# Privacy Policy

**Last updated:** March 27, 2026

## Overview

Blueprint SDLC is an open-source tool that runs entirely on your local machine. We are committed to your privacy and want to be transparent about our practices.

## Data Collection

**Blueprint collects zero user data.** No telemetry, no analytics, no tracking, no personal information.

## Network Requests

Blueprint makes the following optional network requests, all initiated by the user:

| Request | When | What | Data sent |
|---------|------|------|-----------|
| GitHub Releases API | `blueprint update` or auto-update check | Checks for newer CLI version | None (public API, no auth required) |
| GitHub Releases API | First-run binary download via setup script | Downloads the CLI binary | None |
| MCP servers (Context7, Sequential Thinking) | When skills invoke them during a session | Fetches library docs or runs structured reasoning | Query text only (no personal data) |

**No data is sent to Anthropic, the plugin author, or any third party by Blueprint itself.** MCP servers (Context7 by Upstash, Sequential Thinking by Anthropic) are subject to their own privacy policies.

## Local Storage

Blueprint stores the following files locally on your machine:

- `~/.blueprint/bin/blueprint` — CLI binary
- `~/.blueprint/.update-check` — cached version check result (JSON, no personal data)
- `~/.blueprint/logs/` — audit hook logs (session-local, never transmitted)
- `.planning/` — plan files in your project directory (committed to your git repo by you)

## Third-Party Services

Blueprint integrates with these services, each governed by their own privacy policies:

- **GitHub API** — [GitHub Privacy Statement](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)
- **Context7 by Upstash** — [Upstash Privacy Policy](https://upstash.com/trust/privacy.html)
- **Anthropic (Claude Code, Sequential Thinking)** — [Anthropic Privacy Policy](https://www.anthropic.com/privacy)

## Open Source

Blueprint is fully open source under the Apache 2.0 license. You can audit every line of code:

- **Main repo:** https://github.com/skaisser/blueprint
- **Plugin repo:** https://github.com/skaisser/blueprint-plugin

## Contact

If you have questions about this privacy policy:

- **GitHub Issues:** https://github.com/skaisser/blueprint/issues
- **Author:** Shirleyson Kaisser

## Changes

Any changes to this privacy policy will be reflected in this file and committed to the repository with a clear changelog in the git history.

#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# ///
# ── FALLBACK ONLY ────────────────────────────────────────────────────────────
# blueprint audit (Go binary) is 200x faster and is the primary hook.
# This file is used only if ~/.blueprint/bin/blueprint is not installed.
# Run ./install.sh to register the Go binary as the active PreToolUse hook.
# ─────────────────────────────────────────────────────────────────────────────
"""hooks/blueprint.py — Pre-tool-use audit hook (runner-agnostic).

Audits tool calls, catches plan/workflow violations, enforces prerequisite reads,
blocks dangerous commands, and detects task deletion (lying by omission).

This hook is designed to be portable across agent runners.
Different runners emit different payload shapes and tool names, so we:
- accept multiple payload key variants
- normalize tool names
- support both single-file tools and multi-file patch tools

Exit codes (best-effort):
  0 = allow (with optional warning to stderr)
  2 = block (for runners that treat exit code 2 as "deny tool call")
"""

import json
import os
import re
import sys
from datetime import datetime
from pathlib import Path

# ─── Config ──────────────────────────────────────────────────────────────────

LOG_DIR = Path.home() / ".blueprint" / "hooks" / "logs"
LOG_DIR.mkdir(parents=True, exist_ok=True)

# ─── Read staging branch from config ─────────────────────────────────────────

def _read_staging_branch() -> str:
    """Read staging_branch from blueprint/.config.yml, fallback to 'staging'."""
    config_path = Path("blueprint/.config.yml")
    if config_path.exists():
        try:
            for line in config_path.read_text().splitlines():
                if line.strip().startswith("staging_branch:"):
                    val = line.split(":", 1)[1].strip()
                    if val:
                        return val
        except Exception:
            pass
    return "staging"

STAGING_BRANCH = _read_staging_branch()

# ─── Parse stdin payload ─────────────────────────────────────────────────────

try:
    payload = json.load(sys.stdin)
except (json.JSONDecodeError, ValueError):
    sys.exit(0)  # Can't parse → allow


def _first_key(d: dict, keys: tuple[str, ...], default=None):
    for k in keys:
        if k in d and d.get(k) is not None:
            return d.get(k)
    return default


def normalize_tool_name(name: str) -> str:
    name = (name or "").strip()
    if not name:
        return ""
    # runners may provide names like "functions.read" or "Read".
    return name.lower().split(".")[-1]


# payload variants seen across runners
raw_tool_name = _first_key(payload, ("tool_name", "toolName", "tool", "name"), "")
session_id = _first_key(
    payload, ("session_id", "sessionId", "run_id", "runId"), "unknown"
)
tool_input = (
    _first_key(payload, ("tool_input", "toolInput", "input", "arguments"), {}) or {}
)

# some runners nest tool info
if isinstance(raw_tool_name, dict):
    raw_tool_name = _first_key(raw_tool_name, ("name", "tool_name", "toolName"), "")

tool_name = normalize_tool_name(str(raw_tool_name))

timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
session_short = session_id[:8]

# ─── Session state dir ───────────────────────────────────────────────────────

SESSION_DIR = Path(f"/tmp/agent-audit-{session_id}")
SESSION_DIR.mkdir(parents=True, exist_ok=True)

LOG_FILE = LOG_DIR / f"{datetime.now().strftime('%Y-%m-%d')}.log"

# ─── Helpers ─────────────────────────────────────────────────────────────────


def log(msg: str):
    with open(LOG_FILE, "a") as f:
        f.write(f"[{timestamp}] [SESSION:{session_short}] [{tool_name}] {msg}\n")


def warn(msg: str):
    log(f"⚠️  WARNING: {msg}")
    print(f"⚠️  HOOK WARNING: {msg}", file=sys.stderr)


def block(msg: str):
    log(f"🚫 BLOCKED: {msg}")
    print(f"\n🚫 HOOK BLOCKED: {msg}\n", file=sys.stderr)
    sys.exit(2)


# ─── Audit log: record every tool call ───────────────────────────────────────

summary_keys = [
    "description",
    "command",
    "cmd",
    "file_path",
    "filePath",
    "path",
    "name",
    "subagent_type",
    "model",
    "team_name",
]
parts = []
for k in summary_keys:
    if k in tool_input:
        v = str(tool_input[k])[:80]
        parts.append(f"{k}={v}")
input_summary = " | ".join(parts[:3]) if parts else "no-summary"
log(f"CALL | {input_summary}")

# ─── Extract common fields ───────────────────────────────────────────────────

file_path = (
    tool_input.get("file_path")
    or tool_input.get("filePath")
    or tool_input.get("path")
    or ""
)
command = tool_input.get("command") or tool_input.get("cmd") or ""


def patch_paths(patch_text: str) -> list[str]:
    """Extract file paths from apply_patch-style patch text."""
    if not patch_text:
        return []
    paths: list[str] = []
    for line in patch_text.splitlines():
        m = re.match(r"^\*\*\* (Add File|Update File|Delete File): (.+)$", line.strip())
        if m:
            paths.append(m.group(2).strip())
    return paths


patch_text = str(tool_input.get("patchText") or tool_input.get("patch_text") or "")
patched_files = patch_paths(patch_text) if patch_text else []

# ─── ENFORCEMENT 1: Skill read before frontend edits ─────────────────────────

SKILL_MAP = {
    (".jsx", ".tsx", ".html", ".css"): "frontend-design",
}


def _needs_skill_for_path(path: str) -> str | None:
    for extensions, skill_name in SKILL_MAP.items():
        if any(path.endswith(ext) for ext in extensions):
            return skill_name
    return None


write_like = tool_name in ("write", "edit", "apply_patch", "applypatch")
paths_to_check = []
if file_path:
    paths_to_check.append(file_path)
paths_to_check.extend(patched_files)

if write_like and paths_to_check:
    for p in paths_to_check:
        skill_name = _needs_skill_for_path(p)
        if not skill_name:
            continue
        marker = SESSION_DIR / f"skill-read-{skill_name}"
        if not marker.exists():
            skill_file = f"~/.claude/skills/{skill_name}/SKILL.md"
            block(
                f"Tried to modify '{p}' without reading SKILL.md first. Read {skill_file} before proceeding."
            )

# ─── ENFORCEMENT 2: Track SKILL.md and reference file reads ──────────────────

# Map Skill tool invocations → markers (handles aliases too)
SKILL_TOOL_ALIASES: dict[str, str] = {}
for _exts, _sname in SKILL_MAP.items():
    SKILL_TOOL_ALIASES[_sname] = _sname  # identity

if tool_name == "skill":
    skill_arg = (tool_input.get("skill") or "").strip()
    if skill_arg:
        (SESSION_DIR / f"skill-read-{skill_arg}").touch()
        mapped = SKILL_TOOL_ALIASES.get(skill_arg)
        if mapped and mapped != skill_arg:
            (SESSION_DIR / f"skill-read-{mapped}").touch()
        # Also check if this skill arg is an alias for a known skill
        for known in set(SKILL_MAP.values()):
            if skill_arg.startswith(known) or known.startswith(skill_arg):
                (SESSION_DIR / f"skill-read-{known}").touch()
        log(f"✅ SKILL INVOKED: {skill_arg}")

if tool_name == "read" and file_path:
    # Track SKILL.md reads
    if "SKILL.md" in file_path:
        skill_name = Path(file_path).parent.name
        (SESSION_DIR / f"skill-read-{skill_name}").touch()
        log(f"✅ SKILL READ: {skill_name} ({file_path})")

    # Track reference file reads
    if "plan-template.md" in file_path:
        (SESSION_DIR / "read-plan-template").touch()
        log("✅ REF READ: plan-template.md")

    if "team-execution.md" in file_path:
        (SESSION_DIR / "read-team-execution").touch()
        log("✅ REF READ: team-execution.md")

# ─── ENFORCEMENT 3: Team vs Subagent compliance ──────────────────────────────

if tool_name == "teamcreate":
    team_name = tool_input.get("name") or tool_input.get("team_name") or "unnamed"

    if not (SESSION_DIR / "read-team-execution").exists():
        warn(
            f"Creating team '{team_name}' without reading team-execution.md first. Read ~/.blueprint/references/plan/team-execution.md for delegation strategy."
        )

    (SESSION_DIR / "team-created").touch()
    with open(SESSION_DIR / "teams-created.txt", "a") as f:
        f.write(f"{team_name}\n")
    log(f"✅ TEAM CREATED: {team_name}")

if tool_name == "task":
    has_team = bool(tool_input.get("team_name"))

    if has_team:
        team = tool_input.get("team_name", "?")
        log(f"✅ TEAM WORKER TASK (team={team})")
    else:
        count_file = SESSION_DIR / "standalone-task-count"
        count = int(count_file.read_text()) if count_file.exists() else 0
        count += 1
        count_file.write_text(str(count))

        if count >= 3 and not (SESSION_DIR / "team-created").exists():
            warn(
                f"Standalone Task call #{count} in this session — NO TeamCreate detected. team-execution.md requires TeamCreate for 3+ tasks. You said you'd use teams. Did you lie?"
            )
            log(f"🚨 POTENTIAL LIE: {count} standalone Task calls, 0 TeamCreate calls")
        else:
            log(f"✅ STANDALONE SUBAGENT #{count} (1-2 tasks, acceptable)")

# ─── ENFORCEMENT 4: Track AskUserQuestion calls ──────────────────────────────

if tool_name in ("askuserquestion", "question"):
    (SESSION_DIR / "asked-user").touch()
    log("✅ ASKED USER (AskUserQuestion called)")

# ─── ENFORCEMENT 5: Command checkpoint tracking & prerequisites ──────────────

if tool_name == "bash" and command:
    # Track checkpoints — new format: 🔷 BP: skill [N/TOTAL], legacy: 🏁 [skill:step]
    bp_match = re.search(r"BP:\s+([a-z-]+)\s+\[(\d+)/(\d+)\]", command)
    legacy_match = re.search(r"\[([a-z-]+:[0-9a-z-]+)\]", command)
    if bp_match:
        active_cmd = bp_match.group(1)
        checkpoint = f"{active_cmd}:{bp_match.group(2)}"
        with open(SESSION_DIR / "checkpoints.txt", "a") as f:
            f.write(f"{checkpoint}\n")
        (SESSION_DIR / "active-command").write_text(active_cmd)
        log(f"🔷 BP CHECKPOINT: {checkpoint}")
    elif legacy_match:
        checkpoint = legacy_match.group(1)
        with open(SESSION_DIR / "checkpoints.txt", "a") as f:
            f.write(f"{checkpoint}\n")
        active_cmd = checkpoint.split(":")[0]
        (SESSION_DIR / "active-command").write_text(active_cmd)
        log(f"🔷 BP CHECKPOINT: {checkpoint}")

    # ── Helper: autonomous pipelines that should NOT require AskUserQuestion ──
    def _is_autonomous_pipeline(cmd_name: str) -> bool:
        return cmd_name in ("batch-flow", "flow-auto", "flow-auto-wt")

    # ── Prerequisite: /finish must AskUserQuestion before second gh pr merge ──
    # Exception: autonomous pipelines (flow-auto, flow-auto-wt, batch-flow)
    if "gh pr merge" in command:
        active_cmd_file = SESSION_DIR / "active-command"
        active_cmd = (
            active_cmd_file.read_text().strip() if active_cmd_file.exists() else ""
        )

        if _is_autonomous_pipeline(active_cmd):
            log(f"✅ gh pr merge in autonomous pipeline ({active_cmd}) — allowed")
        elif active_cmd == "finish":
            merge_count_file = SESSION_DIR / "gh-pr-merge-count"
            merge_count = (
                int(merge_count_file.read_text()) if merge_count_file.exists() else 0
            )
            merge_count += 1
            merge_count_file.write_text(str(merge_count))

            if merge_count >= 2 and not (SESSION_DIR / "asked-user").exists():
                warn(
                    f"Second gh pr merge during /finish without AskUserQuestion. Step 6 requires asking the user before merging {STAGING_BRANCH} → main."
                )

    # ── Prerequisite: gh pr create requires /pr, /finish, or pipeline context ──
    if "gh pr create" in command:
        active_cmd_file = SESSION_DIR / "active-command"
        active_cmd = (
            active_cmd_file.read_text().strip() if active_cmd_file.exists() else ""
        )

        if _is_autonomous_pipeline(active_cmd):
            log(f"✅ gh pr create in autonomous pipeline ({active_cmd}) — allowed")
        elif active_cmd not in ("pr", "finish", "hotfix-push"):
            warn(
                f"gh pr create called outside /pr, /finish, or /hotfix-push context. PRs should only be created via these skills. Active context: '{active_cmd or 'none'}'"
            )

# ─── ENFORCEMENT 6: Plan file writes require template read ────────────────────

if write_like and paths_to_check:
    for p in paths_to_check:
        if "blueprint/" in p and p.endswith(".md") and "/backlog/" not in p:
            if not Path(p).exists():
                if not (SESSION_DIR / "read-plan-template").exists():
                    block(
                        "Creating new plan file without reading plan-template.md first.\n"
                        "Read ~/.blueprint/references/plan/plan-template.md before creating plans."
                    )

# ─── ENFORCEMENT 7: Block GitHub workflow creation without staging branch ─────
# claude-pr-reviewer.yml and tests.yml should only exist in projects that use the staging flow.

if write_like and paths_to_check:
    for p in paths_to_check:
        if ".github/workflows/" in p and Path(p).name in ("claude-pr-reviewer.yml", "tests.yml"):
            import subprocess as _sp_wf
            try:
                _has_staging_wf = _sp_wf.run(
                    ["git", "branch", "-a", "--list", f"*{STAGING_BRANCH}*"],
                    capture_output=True, text=True
                ).stdout.strip()
            except Exception:
                _has_staging_wf = ""
            if not _has_staging_wf:
                block(
                    f"Creating '{Path(p).name}' but this project has no {STAGING_BRANCH} branch.\n\n"
                    "GitHub workflows (claude-pr-reviewer.yml, tests.yml) are only for projects using the\n"
                    f"{STAGING_BRANCH} → main PR flow. Without {STAGING_BRANCH}, there's no PR pipeline to run CI on.\n\n"
                    f"If this project should use the {STAGING_BRANCH} flow, create the {STAGING_BRANCH} branch first:\n"
                    f"  git checkout -b {STAGING_BRANCH} && git push -u origin {STAGING_BRANCH}"
                )

# ─── ENFORCEMENT 8: Full test suite must run in parallel (--processes=10) ────

if tool_name == "bash" and command:
    is_test_command = bool(
        re.search(
            r"(vendor/bin/pest|php\s+artisan\s+test|artisan\s+test|herd\s+coverage)",
            command,
        )
    )

    if is_test_command:
        has_filter = bool(re.search(r"--filter", command))
        has_test_file = bool(re.search(r"tests/\S+", command))
        is_full_suite = not has_filter and not has_test_file

        if is_full_suite:
            has_parallel = bool(re.search(r"--parallel", command))
            has_processes_10 = bool(re.search(r"--processes[=\s]+10", command))

            if not has_parallel or not has_processes_10:
                block(
                    "Wrong way to run the full test suite.\n\n"
                    "You MUST run it in parallel with 10 processes — otherwise it takes 40+ minutes:\n\n"
                    "  Full suite:     ./vendor/bin/pest --parallel --processes=10\n"
                    "  With coverage:  herd coverage ./vendor/bin/pest --coverage --parallel --processes=10\n\n"
                    "These are available as shell aliases:\n"
                    "  ptp   →  ./vendor/bin/pest --parallel --processes=10\n"
                    "  tcq   →  herd coverage ./vendor/bin/pest --coverage --parallel --processes=10\n\n"
                    "Targeted tests (always preferred for speed):\n"
                    '  vendor/bin/pest --filter="TestName"\n'
                    "  vendor/bin/pest tests/Feature/Path/ToTest.php\n"
                    '  herd coverage vendor/bin/pest --coverage --filter="TestName"'
                )

# ─── ENFORCEMENT 9: Plan task deletion detection (anti-lying) ────────────────
# Trust but CHECK — detect when tasks are removed from plan files.
# Claude sometimes "lies by omission" by deleting planned tasks instead of
# implementing them. This catches that by tracking task counts.

if write_like and paths_to_check:
    for p in paths_to_check:
        if "blueprint/" in p and p.endswith("-todo.md") and Path(p).exists():
            # Read current file to count existing tasks
            try:
                current_content = Path(p).read_text()
                current_checked = len(re.findall(r"- \[x\]", current_content, re.IGNORECASE))
                current_unchecked = len(re.findall(r"- \[ \]", current_content))
                current_total = current_checked + current_unchecked

                # Get the new content from the edit
                old_string = tool_input.get("old_string", "")
                new_string = tool_input.get("new_string", "")

                if old_string and new_string:
                    # Edit tool — check if tasks are being removed
                    old_tasks = len(re.findall(r"- \[[x ]\]", old_string, re.IGNORECASE))
                    new_tasks = len(re.findall(r"- \[[x ]\]", new_string, re.IGNORECASE))

                    tasks_removed = old_tasks - new_tasks
                    if tasks_removed > 0:
                        # Check if unchecked tasks are being deleted (lying)
                        old_unchecked = len(re.findall(r"- \[ \]", old_string))
                        new_unchecked = len(re.findall(r"- \[ \]", new_string))
                        unchecked_removed = old_unchecked - new_unchecked

                        if unchecked_removed > 0:
                            warn(
                                f"🚨 PLAN TASK DELETION DETECTED: {unchecked_removed} unchecked task(s) removed from '{Path(p).name}'.\n"
                                f"   Before: {old_tasks} tasks ({old_unchecked} unchecked) → After: {new_tasks} tasks ({new_unchecked} unchecked)\n"
                                f"   You MUST implement planned tasks, not delete them. If a task is no longer needed, mark it [x] SKIPPED with reason.\n"
                                f"   Trust but CHECK — this edit looks like lying by omission."
                            )
                            log(f"🚨 TASK DELETION: {unchecked_removed} unchecked tasks removed from {Path(p).name}")

                elif tool_name == "write":
                    # Write tool (full file replacement) — compare task counts
                    new_content = tool_input.get("content", "")
                    if new_content:
                        new_checked = len(re.findall(r"- \[x\]", new_content, re.IGNORECASE))
                        new_unchecked = len(re.findall(r"- \[ \]", new_content))
                        new_total = new_checked + new_unchecked

                        tasks_lost = current_total - new_total
                        if tasks_lost > 0:
                            unchecked_lost = current_unchecked - new_unchecked
                            if unchecked_lost > 0:
                                warn(
                                    f"🚨 PLAN TASK DELETION DETECTED: {tasks_lost} task(s) disappeared from '{Path(p).name}'.\n"
                                    f"   Before: {current_total} tasks ({current_unchecked} unchecked) → After: {new_total} tasks ({new_unchecked} unchecked)\n"
                                    f"   {unchecked_lost} unchecked task(s) were removed — these should be implemented, not deleted.\n"
                                    f"   If tasks are genuinely not needed, mark them [x] SKIPPED with reason."
                                )
                                log(f"🚨 TASK DELETION (write): {tasks_lost} tasks lost, {unchecked_lost} unchecked, from {Path(p).name}")

                # Save task count snapshot for cross-check
                snapshot_file = SESSION_DIR / f"plan-tasks-{Path(p).stem}"
                snapshot_file.write_text(f"{current_total}:{current_checked}:{current_unchecked}")

            except Exception:
                pass  # Don't block on snapshot errors

# ─── ENFORCEMENT 10: Block dangerous commands ────────────────────────────────

if tool_name == "bash" and command:
    # Block migrate:fresh — unless /squash-migrations skill is active
    if re.search(r"migrate:fresh", command):
        squash_active = (SESSION_DIR / "skill-read-squash-migrations").exists()
        if not squash_active:
            block(
                "migrate:fresh is BLOCKED — it drops all tables and destroys data.\n"
                "Use 'php artisan migrate' for normal migrations.\n\n"
                "If you need migrate:fresh for migration consolidation, invoke /squash-migrations first.\n"
                "That skill unlocks migrate:fresh for the session.\n\n"
                "If you truly need it outside that context, ask the user to run it manually."
            )
        else:
            warn(
                "migrate:fresh ALLOWED — /squash-migrations skill is active.\n"
                "This is expected during migration consolidation. Proceed with caution."
            )

    # Block direct push to main — only for repos that use the staging branch flow.
    # Repos without a staging branch (config repos, docs, etc.) can push directly.
    # Matches: git push [flags] <remote> main  OR  bare git push while on main
    # Uses word-level match to avoid false positives in commit messages.
    _is_explicit_main_push = bool(re.search(r"git\s+push(?:\s+-\S+)*\s+\S+\s+main\b", command))
    _is_bare_push = bool(re.search(r"(?:^|&&|;)\s*git\s+push\s*$", command))
    if _is_explicit_main_push or _is_bare_push:
        import subprocess as _sp
        # Use the cd path from the command (e.g. "cd ~/Sites/project && ... && git push")
        # so the remote/branch check runs in the right directory.
        _cd_match = re.search(r"(?:^|&&|;)\s*cd\s+(\S+)", command)
        _cwd = (
            os.path.expandvars(os.path.expanduser(_cd_match.group(1)))
            if _cd_match else os.getcwd()
        )
        try:
            _branch = _sp.run(
                ["git", "branch", "--show-current"],
                capture_output=True, text=True, cwd=_cwd
            ).stdout.strip()
            # Check if repo has the staging branch (local or remote)
            _has_staging = _sp.run(
                ["git", "branch", "-a", "--list", f"*{STAGING_BRANCH}*"],
                capture_output=True, text=True, cwd=_cwd
            ).stdout.strip()
        except Exception:
            _branch = ""
            _has_staging = ""
        _is_main_branch = _branch == "main"
        if _has_staging and (_is_explicit_main_push or _is_main_branch):
            block(
                "Direct push to main is BLOCKED.\n"
                f"Use the branch flow: feature/* → {STAGING_BRANCH} → main (via PR).\n"
                "If this is an emergency, use /hotfix-push which creates a proper PR."
            )

    # Warn on AI signatures in commit commands
    if re.search(r"git\s+commit", command) and re.search(r"Co-Authored-By|Generated.by.AI|Generated.by.Claude", command, re.IGNORECASE):
        block(
            "AI signatures are BLOCKED in commits.\n"
            "Remove any Co-Authored-By, 'Generated by AI', or similar attribution.\n"
            "The commit-msg hook will also reject these, but catching it early."
        )

    # Block AI signatures in PR creation/editing (gh pr create/edit --body)
    # NOTE: "Generated by /flow-auto — Kaisser SDLC" is ALLOWED (owner branding, not AI signature)
    if re.search(r"gh\s+pr\s+(create|edit)", command):
        ai_sig_patterns = [
            r"Generated\s+with\s+\[?Claude",
            r"Generated\s+with\s+Claude\s+Code",
            r"Co-Authored-By.*claude",
            r"Co-Authored-By.*anthropic",
            r"Co-Authored-By.*noreply@anthropic",
            r"🤖\s*Generated",
            r"Generated\s+by\s+AI",
            r"Generated\s+by\s+Claude",
            r"claude\.com/claude-code",
        ]
        for pattern in ai_sig_patterns:
            if re.search(pattern, command, re.IGNORECASE):
                block(
                    "AI signatures are BLOCKED in pull requests.\n"
                    "Remove any 'Generated with Claude Code', 'Co-Authored-By', '🤖 Generated',\n"
                    "or any AI attribution from the PR title and body.\n"
                    f"Matched pattern: {pattern}"
                )

    # Block AI signatures in PR comments (gh pr comment --body)
    # NOTE: "Generated by /flow-auto — Kaisser SDLC" is ALLOWED (owner branding, not AI signature)
    if re.search(r"gh\s+pr\s+comment", command):
        ai_sig_patterns = [
            r"Generated\s+with\s+\[?Claude",
            r"🤖\s*Generated",
            r"Generated\s+by\s+AI",
            r"Generated\s+by\s+Claude",
            r"claude\.com/claude-code",
        ]
        for pattern in ai_sig_patterns:
            if re.search(pattern, command, re.IGNORECASE):
                block(
                    "AI signatures are BLOCKED in PR comments.\n"
                    "Remove any AI attribution from the comment body.\n"
                    f"Matched pattern: {pattern}"
                )

    # ── BREAKING CHANGE footer detection ──
    if re.search(r"git\s+commit", command):
        if "BREAKING CHANGE:" in command:
            log("⚠️ BREAKING CHANGE footer detected in commit")

# ─── ENFORCEMENT 11: @claude review prompt must be comprehensive ─────────────
# Agents sometimes shortcut the review comment to "@claude review this pr".
# The /review skill specifies a full prompt — enforce it.

if tool_name == "bash" and command:
    if re.search(r"gh\s+pr\s+comment", command) and re.search(r"@claude\s+review", command, re.IGNORECASE):
        has_full_prompt = bool(re.search(r"check if we are able to merge", command, re.IGNORECASE))
        if not has_full_prompt:
            block(
                "@claude review comment is too short — agents must use the full prompt.\n\n"
                "Required comment body:\n"
                '  "@claude review this PR and check if we are able to merge. '
                'Analyze the code changes for any issues, security concerns, or improvements needed."\n\n'
                "This ensures the CI reviewer performs a thorough analysis, not a superficial pass."
            )

# ─── ENFORCEMENT 12: Plan-check skip detection ─────────────────────────────
# Warn if /pr is invoked without /plan-check having run first (when a plan exists)

if tool_name == "bash" and command:
    if "gh pr create" in command or re.search(r"gh api repos/\S+/pulls\b.*-X POST", command):
        checkpoints_file = SESSION_DIR / "checkpoints.txt"
        checkpoints = checkpoints_file.read_text() if checkpoints_file.exists() else ""

        had_plan_approved = "plan-approved:" in checkpoints
        had_plan_check = "plan-check:" in checkpoints

        if had_plan_approved and not had_plan_check:
            warn(
                "Creating PR after /plan-approved but /plan-check was never run.\n"
                "/plan-check is the quality gate — it audits code vs plan, catches orphaned tests,\n"
                "and creates the audit commit. Run /plan-check before /pr."
            )

# ─── ENFORCEMENT 13: Unchecked acceptance criteria on PR creation ─────────────
# Warn if creating a PR while the plan file has unchecked acceptance criteria.

if tool_name == "bash" and command:
    if "gh pr create" in command or re.search(r"gh\s+pr\s+edit", command):
        import glob as _glob_ac
        plan_files = _glob_ac.glob("blueprint/*-todo.md")
        for pf in plan_files:
            try:
                content = Path(pf).read_text()
                # Find the Acceptance Criteria section
                ac_match = re.search(
                    r"##\s*Acceptance\s*Criteria\s*\n(.*?)(?=\n##\s|\Z)",
                    content, re.DOTALL | re.IGNORECASE
                )
                if ac_match:
                    ac_section = ac_match.group(1)
                    unchecked = re.findall(r"- \[ \]", ac_section)
                    total = len(re.findall(r"- \[[x ]\]", ac_section, re.IGNORECASE))
                    if unchecked and total > 0:
                        warn(
                            f"PR has {len(unchecked)}/{total} unchecked acceptance criteria in '{Path(pf).name}'.\n"
                            "/plan-check should verify and mark each criterion before creating the PR.\n"
                            "Run /plan-check or manually verify and check off the acceptance criteria."
                        )
            except Exception:
                pass

# ─── ENFORCEMENT 14: flow-auto step enforcement ─────────────────────────────
# Block PR creation during flow-auto if mandatory steps were skipped.
# flow-auto has 8 steps — steps 5 (plan-check) and 7 (review-loop) are mandatory.
# The checkpoint pattern is: 🤖 [flow-auto:N] — tracked via the checkpoint tracker.

if tool_name == "bash" and command:
    checkpoints_file = SESSION_DIR / "checkpoints.txt"
    checkpoints = checkpoints_file.read_text() if checkpoints_file.exists() else ""

    is_flow_auto = "flow-auto:" in checkpoints
    is_pr_creation = bool(
        "gh pr create" in command
        or re.search(r"gh api repos/\S+/pulls\b.*-X POST", command)
    )

    if is_flow_auto and is_pr_creation:
        # Check mandatory steps were completed
        had_step_5 = "flow-auto:5" in checkpoints  # plan-check / audit
        had_step_4 = "flow-auto:4" in checkpoints  # execute

        if had_step_4 and not had_step_5:
            block(
                "flow-auto: Creating PR without running Step 5 (plan-check).\n\n"
                "The flow-auto pipeline requires ALL steps in order:\n"
                "  1. Initialize → 2. Plan → 3. Review → 4. Execute → 5. Plan Check → 6. PR → 7. Review Loop → 8. Report\n\n"
                "You skipped Step 5 (auditing implementation). Run it now:\n"
                '  echo "🤖 [flow-auto:5] auditing implementation"\n'
                "  uv run ~/.blueprint/scripts/sdlc.py context --diffs\n"
                "  # ... compare plan vs implementation, fix mismatches\n"
                "  uv run ~/.blueprint/scripts/sdlc.py sync"
            )

    # Enforce review loop — block final report comment if review wasn't attempted
    is_pr_comment = bool(
        re.search(r"gh pr comment", command)
        and "Final Report" in command
    )
    if is_flow_auto and is_pr_comment:
        had_step_7 = "flow-auto:7" in checkpoints  # review loop
        had_step_6 = "flow-auto:6" in checkpoints  # PR creation

        if had_step_6 and not had_step_7:
            warn(
                "flow-auto: Posting final report without running Step 7 (review loop).\n\n"
                "The review loop is mandatory — trigger @claude review on the PR:\n"
                '  echo "🤖 [flow-auto:7] starting review loop"\n'
                "  gh pr comment $PR_NUM --body \"@claude review this PR\"\n\n"
                "If the GitHub Action is not set up, poll for 5 min then continue.\n"
                "You MUST at least attempt the review step."
            )

# ─── ENFORCEMENT 15: Backlog CLI enforcement — block manual parsing ──────────
# Agents should use `blueprint backlog` CLI, not grep/sed/awk/cat on backlog files.

if tool_name == "bash" and command:
    _backlog_manual = (
        re.search(r"(grep|sed|awk).*blueprint/backlog", command)
        or re.search(r"for\s+\w+\s+in\s+.*blueprint/backlog", command)
        or re.search(r"cat\s+.*blueprint/backlog", command)
    )
    if _backlog_manual:
        block(
            "Use `blueprint backlog` CLI to read backlog files — it handles both YAML and legacy formats correctly.\n\n"
            "Available commands:\n"
            "  blueprint backlog                  # JSON output (default)\n"
            "  blueprint backlog --format table   # Pretty table\n"
            "  blueprint backlog --archive        # Include archived items\n"
            "  blueprint backlog migrate           # Convert old format → YAML\n\n"
            "Never parse backlog files with grep/sed/awk/cat — the CLI is faster and correct."
        )

# ─── All good ────────────────────────────────────────────────────────────────
sys.exit(0)

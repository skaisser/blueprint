#!/bin/bash

# Claude Code Status Line
# Shows: Model | Context Bar % ~Remaining | 2x/1x Promo | Duration | Branch ↑↓ | ±Changes | Folder
# Location: config/statusline.sh → installed to ~/.claude/statusline.sh

input=$(cat)
model_name=$(echo "$input" | jq -r '.model.display_name // "Claude"')
current_dir=$(echo "$input" | jq -r '.workspace.current_dir')
project_dir=$(echo "$input" | jq -r '.workspace.project_dir // ""')

# Folder display — project name or relative path
if [ -n "$project_dir" ] && [ "$current_dir" != "$project_dir" ]; then
    rel_path="${current_dir#$project_dir/}"
    if [ "$rel_path" = "$current_dir" ]; then
        dir_display=$(basename "$current_dir")
    else
        dir_display="$(basename "$project_dir")/$rel_path"
    fi
else
    dir_display=$(basename "$current_dir")
fi

# Context window calculation
pct=0
current=0
size=1
usage=$(echo "$input" | jq '.context_window.current_usage')
if [ "$usage" != "null" ]; then
    input_tokens=$(echo "$usage" | jq '.input_tokens // 0')
    cache_create=$(echo "$usage" | jq '.cache_creation_input_tokens // 0')
    cache_read=$(echo "$usage" | jq '.cache_read_input_tokens // 0')
    current=$((input_tokens + cache_create + cache_read))
    size=$(echo "$input" | jq '.context_window.context_window_size // 1')
    if [ "$size" != "null" ] && [ "$size" -gt 0 ]; then
        pct=$((current * 100 / size))
    fi
fi

# Context bar (10 chars) — thresholds: 0-50 green, 50-65 orange, 65+ red
bar_width=10
filled=$((pct * bar_width / 100))
[ "$filled" -gt "$bar_width" ] && filled=$bar_width
empty=$((bar_width - filled))
bar=""
for ((i=0; i<filled; i++)); do bar+="█"; done
for ((i=0; i<empty; i++)); do bar+="░"; done

if [ "$pct" -le 50 ]; then
    bar_color="\033[32m"  # Green — safe zone
elif [ "$pct" -le 65 ]; then
    bar_color="\033[33m"  # Orange — caution
else
    bar_color="\033[31m"  # Red — wrap up or start new session
fi

# Duration from session stats
duration_ms=$(echo "$input" | jq -r '.cost.total_duration_ms // 0')

# Format duration (Xm or Xh Ym)
duration_display=""
if [ "$duration_ms" != "0" ] && [ "$duration_ms" != "null" ]; then
    total_sec=$((duration_ms / 1000))
    hours=$((total_sec / 3600))
    mins=$(( (total_sec % 3600) / 60 ))
    if [ "$hours" -gt 0 ]; then
        duration_display="${hours}h${mins}m"
    elif [ "$mins" -gt 0 ]; then
        duration_display="${mins}m"
    fi
fi

# Estimated remaining session time based on token burn rate
remaining_display=""
if [ "$duration_ms" != "0" ] && [ "$duration_ms" != "null" ] && [ "$current" -gt 0 ] && [ "$size" -gt 0 ]; then
    total_sec=$((duration_ms / 1000))
    if [ "$total_sec" -gt 60 ]; then
        # tokens per second burn rate
        remaining_tokens=$((size - current))
        if [ "$remaining_tokens" -gt 0 ]; then
            # estimated seconds remaining = remaining_tokens / (current_tokens / elapsed_seconds)
            est_remaining_sec=$((remaining_tokens * total_sec / current))
            est_hours=$((est_remaining_sec / 3600))
            est_mins=$(( (est_remaining_sec % 3600) / 60 ))
            if [ "$est_hours" -gt 0 ]; then
                remaining_display="~${est_hours}h${est_mins}m left"
            elif [ "$est_mins" -gt 0 ]; then
                remaining_display="~${est_mins}m left"
            fi
        fi
    fi
fi

# 200k token threshold warning
exceeds_200k=$(echo "$input" | jq -r '.exceeds_200k_tokens // false')

# Worktree info
wt_name=$(echo "$input" | jq -r '.worktree.name // ""')

# Claude March 2026 promotion: 2x usage outside 8AM-2PM ET weekdays (Mar 13-28)
# After promotion ends, this section simply won't show anything.
promo_display=""
now_epoch=$(date +%s)
promo_start=1773370800   # 2026-03-13 00:00:00 local
promo_end=1774753199     # 2026-03-28 23:59:59 local

if [ "$now_epoch" -ge "$promo_start" ] && [ "$now_epoch" -le "$promo_end" ]; then
    # Get current hour and day-of-week in ET (America/New_York)
    et_hour=$(TZ="America/New_York" date +%H | sed 's/^0//')
    et_dow=$(TZ="America/New_York" date +%u)  # 1=Mon, 7=Sun

    if [ "$et_dow" -le 5 ] && [ "$et_hour" -ge 8 ] && [ "$et_hour" -lt 14 ]; then
        # Peak hours: weekday 8AM-2PM ET
        promo_display="1x"
        promo_color="\033[33m"  # Yellow — peak, normal rate
    else
        # Off-peak: 2x usage
        promo_display="2x"
        promo_color="\033[32m"  # Green — doubled!
    fi
fi

# Git info: branch + code changes (cached for performance)
git_info=""
code_stats=""
git_cache="/tmp/claude_statusline_git_$$"
cache_age=999

if [ -f "$git_cache" ]; then
    cache_age=$(( $(date +%s) - $(stat -f %m "$git_cache" 2>/dev/null || echo 0) ))
fi

if [ "$cache_age" -gt 5 ] && git -C "$current_dir" rev-parse --git-dir > /dev/null 2>&1; then
    git_branch=$(git -C "$current_dir" --no-optional-locks branch --show-current 2>/dev/null)
    if [ -n "$git_branch" ]; then
        # Code written: insertions/deletions across working + staged
        ins=0; del=0
        diff_stat=$(git -C "$current_dir" --no-optional-locks diff --numstat 2>/dev/null)
        staged_stat=$(git -C "$current_dir" --no-optional-locks diff --cached --numstat 2>/dev/null)
        all_stat=$(printf "%s\n%s" "$diff_stat" "$staged_stat")
        if [ -n "$(echo "$all_stat" | tr -d '[:space:]')" ]; then
            while IFS=$'\t' read -r a d _; do
                [ "$a" != "-" ] && ins=$((ins + a)) 2>/dev/null
                [ "$d" != "-" ] && del=$((del + d)) 2>/dev/null
            done <<< "$all_stat"
            if [ "$ins" -gt 0 ] || [ "$del" -gt 0 ]; then
                code_stats="+${ins}/-${del}"
            fi
        fi

        # Ahead/behind
        upstream_info=""
        upstream=$(git -C "$current_dir" --no-optional-locks rev-parse --abbrev-ref @{upstream} 2>/dev/null)
        if [ -n "$upstream" ]; then
            ahead=$(git -C "$current_dir" --no-optional-locks rev-list --count @{upstream}..HEAD 2>/dev/null || echo 0)
            behind=$(git -C "$current_dir" --no-optional-locks rev-list --count HEAD..@{upstream} 2>/dev/null || echo 0)
            [ "$ahead" -gt 0 ] && upstream_info="↑${ahead}"
            [ "$behind" -gt 0 ] && upstream_info="${upstream_info}↓${behind}"
            [ -n "$upstream_info" ] && upstream_info=" ${upstream_info}"
        fi

        git_info="${git_branch}${upstream_info}"
    fi

    # Cache result
    echo "${git_info}|${code_stats}" > "$git_cache" 2>/dev/null
elif [ -f "$git_cache" ]; then
    # Use cached result
    cached=$(cat "$git_cache")
    git_info="${cached%%|*}"
    code_stats="${cached##*|}"
fi

# Colors
reset="\033[0m"
dim="\033[2m"
blue="\033[34m"
green="\033[32m"
cyan="\033[36m"
yellow="\033[33m"
magenta="\033[35m"
red="\033[31m"
bold="\033[1m"

# Build: Model | Bar % ~remaining | 2x/1x | Duration | Branch | ±Code | Folder
output="${magenta}${model_name}${reset}"
output+=" ${dim}│${reset} ${bar_color}${bar}${reset} ${cyan}${pct}%${reset}"

# 200k warning
if [ "$exceeds_200k" = "true" ]; then
    output+=" ${red}${bold}200K+${reset}"
fi

# Estimated remaining time
if [ -n "$remaining_display" ]; then
    output+=" ${dim}${remaining_display}${reset}"
fi

# Promotion indicator (2x/1x)
if [ -n "$promo_display" ]; then
    output+=" ${dim}│${reset} ${promo_color}${bold}${promo_display}${reset}"
fi

# Duration
if [ -n "$duration_display" ]; then
    output+=" ${dim}│${reset} ${yellow}${duration_display}${reset}"
fi

# Git branch
if [ -n "$git_info" ]; then
    # Show worktree name instead of branch if in a worktree
    if [ -n "$wt_name" ]; then
        output+=" ${dim}│${reset} ${green}${wt_name}${reset}"
    else
        output+=" ${dim}│${reset} ${green}${git_info}${reset}"
    fi
fi

# Code changes
if [ -n "$code_stats" ]; then
    output+=" ${yellow}${code_stats}${reset}"
fi

# Folder
output+=" ${dim}│${reset} ${blue}${dir_display}${reset}"

echo -e "$output"

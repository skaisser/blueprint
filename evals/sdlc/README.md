# Blueprint SDLC Eval Framework

This directory contains the autoresearch eval framework for Blueprint's 22 core SDLC skills.

## Overview

Evals validate that each skill behaves correctly and produces expected outputs. The framework supports:

- **Unit evals**: Test individual skill outputs against expected results
- **Integration evals**: Test skill interactions within the SDLC workflow
- **Dashboard**: `dashboard.py` — view eval results across all skills
- **Optimizer**: `optimize.py` — identify and improve underperforming skills

## Core Skills (22)

### Backlog & Planning
- `backlog` — manage backlog items
- `plan` — generate implementation plans
- `plan-review` — review a plan before approval
- `plan-approved` — mark a plan as approved
- `plan-check` — verify plan validity

### Code Review & PRs
- `pr` — create pull requests
- `review` — review code changes
- `address-pr` — address PR feedback
- `finish` — finalize and close work

### Flow & Automation
- `flow` — guided SDLC flow
- `flow-auto` — automated flow execution
- `flow-auto-wt` — automated flow with worktrees
- `batch-flow` — batch multiple flow runs

### Quick Actions
- `quick` — quick one-off tasks
- `hotfix-push` — emergency hotfix workflow
- `resume` — resume interrupted work

### Git Operations
- `commit` — create commits
- `ship` — ship completed work
- `push` — push changes
- `branch` — manage branches
- `complete` — complete a feature branch

### Testing
- `test` — run targeted tests
- `run-tests` — run the full test suite
- `coverage` — check test coverage
- `tdd-review` — review TDD compliance

### Context & Status
- `start` — start a work session
- `context` — load project context
- `status` — show current status

### Meta
- `skill-creator` — create new skills

## Running Evals

```bash
# Run all evals
python evals/sdlc/dashboard.py

# Optimize underperforming skills
python evals/sdlc/optimize.py
```

## Notes

- All references to `blueprint` refer to the Blueprint CLI tool
- Eval files follow the pattern `eval-<skill-name>.md`
- Results are stored in `.blueprint/eval-results/`

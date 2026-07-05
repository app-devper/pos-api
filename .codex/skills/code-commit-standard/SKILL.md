---
name: code-commit-standard
description: Standard workflow for staging and creating clean git commits in this repository. Use when the user asks to commit code, prepare a commit, standardize commit messages, or wants commit hygiene such as semantic grouping, focused staging, and concise conventional-style messages.
---

# Code Commit Standard

Use this skill when the task includes creating commits or preparing changes for commit.

## Goals

- Keep each commit focused and reviewable.
- Stage only files related to the requested change.
- Use concise conventional-style commit messages.
- Respect existing uncommitted work from the user.

## Workflow

1. Inspect `git status --short` and identify unrelated changes before staging anything.
2. Review the diff for the files relevant to the task.
3. Group changes by purpose in this order when possible:
   - `feat`
   - `test`
   - `docs`
   - `refactor`
   - `chore`
4. Stage files per semantic group, not as one broad snapshot.
5. Create one commit per group when the groups are meaningfully separate.
6. Use non-interactive git commands only.

## Message Standard

Prefer this format:

```text
type(scope): short summary
```

Rules:

- Keep the summary concise and specific.
- Use lowercase `type`.
- Use `scope` when it clarifies the area changed, such as `repository`, `middleware`, `auth`, or `api`.
- Use imperative phrasing.
- Avoid trailing periods.

Examples:

```text
feat(repository): add branch-safe sequence increment
test(repository): cover sequence reset edge cases
docs(api): clarify product sync behavior
refactor(middleware): simplify auth token parsing
chore(repo): align local codex commit skill
```

## Staging Rules

- Never stage unrelated dirty files just to make the working tree clean.
- If the repo already contains unrelated edits, leave them unstaged unless the user explicitly asks to include them.
- If a file mixes unrelated changes with the requested work and cannot be cleanly separated, pause and tell the user.
- Do not amend existing commits unless the user explicitly asks.
- Do not use destructive git commands to force a clean state.

## Validation Before Commit

Before committing:

- Run targeted validation for the changed area when practical.
- For Go files touched in this repo, run `gofmt -w` on edited files.
- Prefer focused tests first, then broader tests only when appropriate.

## Repo-Specific Notes

- This repository is a Go POS API. Favor commit scopes that reflect the current layout, such as `repository`, `middleware`, `domain`, `db`, or `featues`.
- Preserve existing package naming, including current typos that are part of the codebase contract.
- Follow the repo commit grouping guidance: `feat -> test -> docs -> refactor -> chore`.

## Response Pattern

When asked to commit:

1. Briefly state which files or semantic groups will be committed.
2. Stage only the intended files.
3. Commit with a concise message that matches the group.
4. Report the commit summary and mention any intentionally excluded dirty files.

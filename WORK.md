# Zap — Work Log

## Current Phase
Stable. No open bugs.

## Current Tasks
| # | Task | Status | Notes |
|---|------|--------|-------|

## Backlog / Feature Ideas
| Priority | Task | Notes |
|----------|------|-------|

## DevLog
### 2026-03-20: Bug fixes + cleanup
Fixed add-mode refresh bug (`cancelEdit()` reset mode before check), editor detection now respects $EDITOR, removed type column, dropped sort-by-type (4 sort modes now: Project, Recent, Name, Path), tightened layout spacing. Edit fields: 4 fields (Name, Project, Path, Description) instead of 5.
Files: update.go, view.go, helpers.go, internal/storage/storage.go

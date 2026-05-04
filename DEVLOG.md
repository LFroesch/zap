## DevLog
### 2026-05-04: Fix help panel clipping
Adjusted the help panel to size against its full border frame instead of treating the entire main content area as inner content. Help paging now uses the same visible body height as the renderer, and a regression test asserts the header remains visible and the full view stays within the terminal height.
Files: internal/ui/help.go, helpers.go, help_test.go, WORK.md

### 2026-05-04: Fix help modal crash
Stopped help-mode rendering from recursing through `mainContentHeight()`, `renderStatusBar()`, and help scroll sizing. Header and footer heights now use fixed single-line sizing, and a regression test covers rendering the help view after `?`.
Files: helpers.go, help_test.go, WORK.md

### 2026-05-04: Full-terminal layout sizing
Removed the hard-coded layout overhead and now size the main content area from the actual rendered header and status bar heights. This fixes cases where the app left unused terminal rows and keeps help, paging, and inline editing aligned with the real viewport.
Files: model.go, helpers.go, view.go, update.go, README.md, WORK.md

### 2026-04-15: Scrollable bounded help view
Reworked help mode so it renders inside the normal app layout instead of overflowing the terminal. The help body now scrolls with `j/k`, `pgup/pgdn`, and `g/G` while keeping the main header and footer visible.
Files: model.go, update.go, view.go, helpers.go, internal/ui/help.go, README.md, WORK.md

### 2026-03-23: Doc suite added
Added CLAUDE.md, agent_spec.md. Updated WORK.md with feature ideas. Added license reference to README.

### 2026-03-20: Bug fixes + cleanup
Fixed add-mode refresh bug (`cancelEdit()` reset mode before check), editor detection now respects $EDITOR, removed type column, dropped sort-by-type (4 sort modes now: Project, Recent, Name, Path), tightened layout spacing. Edit fields: 4 fields (Name, Project, Path, Description) instead of 5.
Files: update.go, view.go, helpers.go, internal/storage/storage.go

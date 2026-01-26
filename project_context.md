# Project Context

This file captures coding preferences, architectural decisions, and project-specific conventions for developers working on this project. This is a living document that evolves with the project.

For user-facing documentation, installation, usage, and features, see [README.md](README.md).

## Development Workflow

- **Iterative Execution**: Work in small, incremental steps rather than large monolithic changes
- **Collaborative Decision-Making**: When multiple implementation approaches are possible, pause to discuss options and decide together before proceeding
- **Continuous Feedback**: Each iteration should be reviewable and testable before moving to the next step
- **Pre-Commit Testing**: Execute all tests before any git commits to ensure code quality and prevent regressions
- **No Auto-Commit**: Do not commit changes unless explicitly asked or prompted by the user
- **Test Gate**: Run all tests before commits (`go test ./...`). Do not commit if tests fail unless they are fixed first
- **Feature Branches**: Work in feature branches for new features, enhancements, or significant changes. Merge to main after review and testing
- **Security**: Never commit environment variables, API keys, passwords, tokens, or other secrets to the repository. Use environment variables, configuration files excluded from git, or secure secret management tools
- **Agent scope**: Agents must not search or traverse the user's home folder unless the user explicitly asks to look at specific subdirectories (e.g. `~/.config`, `~/.cursor`, or other named config directories)

## Coding Preferences

- **Modularity**: Functionality should be modularized with clear separation of concerns, making each component individually unit testable
- **Testing Approach**: 
  - Write tests as we go (alongside implementation)
  - Unit tests for individual concerns/components
  - Integration tests for common code paths that execute across multiple concerns
  - Tests should be executable documentation - clear, descriptive, and demonstrating expected behavior

## Architecture & Implementation

### Component Flow
```
main.go → Scanner.Scan() → Matcher.Group() → TUI (bubbletea) → DiffExecutor
```

**Note**: The legacy `InteractiveCLI` (interactive.go) is kept for reference but is no longer used. The current implementation uses `tui.go` with bubbletea for the interactive interface.

### Key Algorithms

**Prefix Matching (matcher.go)**:
- Uses union-find (disjoint set) algorithm for transitive grouping
- Files grouped if they share common prefix ≥ minPrefixLength
- Matching based on filename only (path ignored)
- Returns only groups with 2+ files

**TUI File Selection (tui.go)**:
- Two-step selection process: first file, then second file
- Navigation automatically skips the first file when selecting the second file
- Group selection shows filenames for better context
- Uses bubbletea for interactive terminal UI
- States: group selection → first file selection → second file selection → diff view

**Diff Execution (diff.go)**:
- Uses system `diff` command (default: `diff -y --width=120`)
- Handles exit codes: 0=identical, 1=different (expected), 2+=error
- Side-by-side format for interactive display

### Important Behaviors

- Scanner: Non-recursive, ignores subdirectories
- Matcher: Filename-only matching (uses `filepath.Base()`)
- TUI: Uses bubbletea with state machine (group → first file → second file → diff)
- DiffExecutor: Non-zero exit for differences is expected, not an error
- Legacy InteractiveCLI: Deprecated but kept for reference/testing

### Quick Development Commands

```bash
go test ./...              # Run all tests
go test -v -run TestName   # Run specific test
go build -o doppel .       # Build binary
./doppel testdata          # Manual test
```

### Component Dependencies

- `main.go` → all components (orchestration)
- `tui.go` → `diff.go` (uses DiffExecutor)
- All components are independent and testable in isolation

## Recent Changes & Notes

### TUI Implementation (Current)
- Replaced interactive CLI with bubbletea-based TUI for better UX
- Two-step file selection: select first file, then second file
- Group selection displays filenames for better context
- Navigation uses arrow keys (↑/↓) or vim keys (j/k)
- Known considerations:
  - Alignment issues with lipgloss styles were addressed by using fixed-width prefixes and applying styles only to text content
  - Window size handling via `tea.WindowSizeMsg` for responsive layout

### Future Enhancements (from README)
- Auto-diff for pairs: Skip file selection when group has exactly 2 files
- Colorized diff output: Add color coding to differentiate diff changes
- Prefix filter flag: Add `--prefix` CLI flag to filter files before grouping

### Dependencies
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling

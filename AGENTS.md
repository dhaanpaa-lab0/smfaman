# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the CLI entry point; commands live in `cmd/` (Cobra + Bubble Tea TUI).
- Core logic is in `pkgs/`:
  - `pkgs/frontend_mgr/` for CDN requests and version logic.
  - `pkgs/frontend_config/` for config parsing and helpers.
  - `pkgs/cache/` for the local cache system.
- `examples/` contains sample configs (e.g., `examples/frontend.yaml`).
- `frontend/` holds vendored Bootstrap/Bootswatch/jQuery assets; treat as third-party source.
- Default project config is `smartfrontend.yaml` at repo root.

## Available Commands
The CLI provides the following commands:
- `init` - Create new configuration file interactively
- `add` - Add library to configuration with version validation
- `delete` (aliases: `del`, `pkgdel`, `d`) - Remove library from configuration
- `upgrade` (alias: `u`) - Upgrade library versions (single or all)
- `clean` (aliases: `rm`, `remove`) - Remove library destination folders
- `install` - Install smfaman binary to ~/bin and update PATH
- `pkgmgr` - Interactive TUI package manager for editing configuration
- `sync` - Download libraries to local filesystem
- `pkgver` - List and browse package versions from CDN
- `get` - Download remote configuration files
- `cache` - Manage local cache (stats, clear, clean)

## Build, Test, and Development Commands
- `go build -o smfaman` builds the CLI binary.
- `go run main.go [command]` runs the CLI locally (e.g., `go run main.go sync --dry-run`).
- `go test ./...` runs the full test suite.
- `go test ./cmd -v -run TestAdd` targets a specific command test.
- `go test ./cmd -v -run TestDelete` tests the delete command.
- `go test ./cmd -v -run TestUpgrade` tests the upgrade command.
- `go test ./cmd -v -run TestClean` tests the clean command.
- `go test ./cmd -v -run TestInstall` tests the install command.
- `go test ./pkgs/frontend_mgr -bench=.` runs benchmarks for CDN logic.

## Testing New Commands
When testing interactive commands:
- Use `--dry-run` flag where available to preview without making changes
- Use non-interactive modes when available (e.g., `add react@18.2.0` vs `add react -i`)
- Test both success and error cases (missing files, invalid versions, etc.)
- Verify config file changes using `cat smartfrontend.yaml`

## Coding Style & Naming Conventions
- Use `gofmt` defaults (tabs for indentation, standard Go formatting).
- Prefer clear, descriptive names; follow Go naming conventions (CamelCase for exported identifiers).
- Keep CLI flags consistent with existing commands (`--interactive`, `--no-cache`, `--force`).

## Testing Guidelines
- Tests live alongside code as `*_test.go` in `cmd/` and `pkgs/`.
- Name tests with `TestXxx` and keep coverage for new CDN behaviors or config logic.
- When adding new CDN functions, add at least one test per provider and a cross-CDN test.

## Commit & Pull Request Guidelines
- Git history uses short, sentence-style summaries; follow that style (e.g., "Add cache support").
- Keep commits focused; avoid mixing unrelated refactors with behavior changes.
- PRs should include: a brief problem/solution description, tests run (or "not run" with reason), and any config changes or CLI output samples when relevant.

## Security & Configuration Tips
- Local config: `~/.smfaman.yaml`; project config: `smartfrontend.yaml`.
- Cache location: `~/.smfaman-cache/`
  - `metadata/` - CDN API responses (24h TTL)
  - `packages/` - Downloaded library files (permanent, shared across projects)
- Avoid committing local configs or cache artifacts; prefer the `examples/` directory for shared templates.
- Package cache is safe to share across projects and significantly speeds up syncing

## Interactive Features
The project includes several interactive TUI (Terminal User Interface) components:
- **Init command** - Interactive config file creation with text inputs
- **Package version selector** - Browse and select versions with search/filter
- **Package manager (pkgmgr)** - Full-featured TUI for editing configuration
  - View/add/edit/delete libraries
  - Edit global settings
  - Interactive version selection
  - Save changes back to config file

All interactive features are built with Bubble Tea and support:
- Arrow key navigation
- Tab/Shift+Tab for field navigation
- Search/filter capabilities
- Keyboard shortcuts for common actions

## Common Development Pitfalls & Best Practices

### Configuration Management
- **Always initialize `config.Libraries` map**: Check for nil before accessing
- **Use helper methods**: Prefer `config.GetLibraryDestination()` over manual path building
- **Validate before saving**: Ensure required fields exist before writing YAML

### Command Implementation
- **Reuse config loaders**: Use shared `loadConfig()` functions to maintain consistency
- **Support dry-run**: Add `--dry-run` flag for destructive operations (upgrade, clean, delete)
- **Provide user feedback**: Show what changed, what was skipped, and suggest next steps
- **Handle aliases properly**: Register command aliases in the `Aliases` field

### Error Handling
- **Wrap errors with context**: Use `fmt.Errorf("action failed: %w", err)` for traceable errors
- **Validate early**: Check file existence, version availability before performing operations
- **User-friendly messages**: Avoid exposing internal errors; suggest solutions

### Testing Strategy
- **Test both interactive and non-interactive modes**: Ensure CLI args work without TUI
- **Mock network calls sparingly**: Most tests use real CDN calls but handle failures gracefully
- **Use `--dry-run` in tests**: Verify logic without side effects
- **Test error paths**: Invalid versions, missing files, network failures

### TUI Development
- **Keep models simple**: Separate view logic from business logic
- **Test without TUI first**: Implement core functionality in non-interactive mode
- **Handle window resize**: TUI should adapt to terminal size changes
- **Provide escape hatches**: Always allow `q`, `Esc`, or `Ctrl+C` to exit

### Performance Considerations
- **Leverage cache**: All CDN API calls should check cache first
- **Avoid redundant fetches**: Batch operations where possible
- **Show progress**: Use progress bars for long-running operations (sync, upgrade all)

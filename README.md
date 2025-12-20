# Smart Frontend Asset Manager (smfaman)

A CLI tool for managing frontend assets from CDNs when you don't need bundling or the full infrastructure of a JavaScript SPA.

**smfaman** helps you manage frontend libraries (Bootstrap, React, jQuery, etc.) by fetching them directly from CDNs, validating integrity hashes, and keeping your frontend dependencies organized without npm or bundlers.

## Features

- 🚀 **Multi-CDN Support**: Fetch from jsDelivr, UNPKG, and CDNJS
- 🔒 **SRI Hash Verification**: Automatic Subresource Integrity hashes
- 📦 **Package Management**: Add, delete, upgrade libraries with version validation
- 🔍 **Interactive TUI**: Browse versions, manage packages with keyboard navigation
- ⚡ **Lightweight**: No Node.js required - pure Go binary
- 📝 **YAML Configuration**: Simple, version-controllable config files
- 💾 **Smart Caching**: Local CDN metadata cache with 24-hour TTL
- 📦 **Package File Caching**: Downloaded files cached globally for faster re-syncing
- 🎨 **Progress Tracking**: Real-time download progress bars
- 🔄 **Incremental Sync**: Only downloads missing or changed files
- 🔧 **CLI & TUI Modes**: Work your way - command line or interactive interface
- 📥 **Auto-Installation**: Self-installing binary with PATH management
- 🧹 **Clean Management**: Remove outdated libraries and clean caches

## Supported CDNs

| CDN | Speed* | Files | Features |
|-----|--------|-------|----------|
| **jsDelivr** | ~57ms | Hierarchical | Stats API, entrypoints, fastest |
| **UNPKG** | ~76ms | Flat list | Complete metadata, file sizes, MIME types |
| **CDNJS** | ~116ms | Flat list | Comprehensive SRI hashes |

*Benchmark based on fetching Bootstrap 5.3.0 metadata

## Installation

### Quick Install (Recommended)

Build and self-install with automatic PATH management:

```bash
git clone https://github.com/yourusername/smfaman
cd smfaman
go build -o smfaman
./smfaman install          # Installs to ~/bin and updates PATH
```

The `install` command will:
- Create `~/bin` directory if needed
- Copy the binary to `~/bin/smfaman`
- Add `~/bin` to your PATH (persists across sessions)
- Auto-detect your shell (bash, zsh, fish, PowerShell)

### Using Go Install

```bash
go install nexus-sds.com/smfaman@latest
# Then optionally run: smfaman install
```

### Manual Installation

```bash
git clone https://github.com/yourusername/smfaman
cd smfaman
go build -o smfaman
sudo mv smfaman /usr/local/bin/
```

## Quick Start

```bash
# 1. Initialize a new frontend config (interactive)
smfaman init

# 2. Add libraries to your configuration
smfaman add bootstrap@5.3.0
smfaman add react --interactive      # Browse versions interactively
smfaman add jquery                    # Uses latest version

# 3. Sync all configured libraries to local directory
smfaman sync

# 4. Manage your libraries
smfaman upgrade --dry-run             # Check for updates
smfaman upgrade                       # Upgrade all libraries
smfaman pkgmgr                        # Interactive package manager

# 5. Clean up when needed
smfaman delete old-library            # Remove from config
smfaman clean                         # Delete downloaded files

# 6. Install globally (optional)
smfaman install                       # Installs to ~/bin

# Work with custom config files
smfaman -f myproject.yaml sync
```

## Commands

### Command Summary

| Command | Description | Aliases |
|---------|-------------|---------|
| `init` | Create new configuration file | - |
| `add` | Add library to configuration | - |
| `delete` | Remove library from configuration | `del`, `pkgdel`, `d` |
| `upgrade` | Upgrade library versions | `u` |
| `clean` | Remove library destination folders | `rm`, `remove` |
| `install` | Install binary to ~/bin | - |
| `pkgmgr` | Interactive package manager | - |
| `sync` | Download libraries to filesystem | - |
| `pkgver` | List package versions | - |
| `get` | Download remote config file | - |
| `cache stats` | Show cache statistics | - |
| `cache clear` | Clear all cache | - |
| `cache clear-packages` | Clear package cache only | - |
| `cache clean` | Remove expired metadata | - |

### `init`
Create a new smart frontend asset configuration file interactively.

```bash
# Interactive mode (default)
smfaman init

# Force overwrite existing config
smfaman init --force
```

Creates `smartfrontend.yaml` in the current directory with:
- Project name
- Destination path with `{library_name}` template
- Default CDN selection (unpkg, cdnjs, jsdelivr)

### `add`
Add a new library to the configuration with version validation.

```bash
# Add with specific version
smfaman add bootstrap@5.3.0

# Interactive version selector
smfaman add react --interactive
smfaman add react -i

# Add with custom CDN
smfaman add jquery@3.7.1 --cdn cdnjs

# Add with specific files only
smfaman add bootstrap --files "dist/css/bootstrap.min.css" --files "dist/js/bootstrap.bundle.min.js"

# Add with custom output path
smfaman add lodash --output "./custom/lodash"

# Force overwrite if library exists
smfaman add react@18.2.0 --force
```

**Features:**
- Validates version exists on CDN before adding
- Supports scoped packages: `@babel/core@7.22.0`
- Uses latest version if not specified
- Interactive mode for browsing all available versions

### `pkgver`
List and browse available versions for a package from CDN.

```bash
# List versions (shows 20 most recent)
smfaman pkgver react

# Show more versions
smfaman pkgver react --limit 50

# Interactive mode with search/filter
smfaman pkgver react --interactive
smfaman pkgver react -i

# Use specific CDN
smfaman pkgver bootstrap --cdn cdnjs

# Bypass cache
smfaman pkgver jquery --no-cache
```

**Interactive mode:**
- Browse all versions with arrow keys
- Search/filter with `/`
- Shows which version is latest
- Press Enter to select (displays helpful command)

### `delete`
Remove a library from the configuration file.

```bash
# Remove a library from config
smfaman delete react
smfaman del bootstrap
smfaman pkgdel jquery
smfaman d lodash
```

**Note:** This command only removes the library from the configuration file. It does NOT delete downloaded files from your filesystem.

### `upgrade`
Upgrade library versions to newer releases.

```bash
# Upgrade a specific library to latest version
smfaman upgrade react

# Upgrade to a specific version
smfaman upgrade react@18.3.0

# Upgrade all libraries to latest versions
smfaman upgrade

# Interactive version selection
smfaman upgrade bootstrap --interactive
smfaman u jquery -i

# Preview changes without modifying config
smfaman upgrade --dry-run
```

**Features:**
- Checks CDN for latest available versions
- Can upgrade individual libraries or all at once
- Interactive mode for version selection
- Dry-run mode to preview changes

### `clean`
Remove destination folders for all libraries in the configuration.

```bash
# Remove all library folders (with confirmation)
smfaman clean

# Preview what would be deleted
smfaman clean --dry-run

# Remove without confirmation prompt
smfaman clean --force

# Clean with custom config file
smfaman clean -f myproject.yaml
```

**Safety Features:**
- Prompts for confirmation before deleting
- Only deletes directories that exist
- Shows what will be deleted before proceeding

### `install`
Install smfaman binary to user's bin directory and update PATH.

```bash
# Install to ~/bin
smfaman install

# Overwrite if already installed
smfaman install --force
```

**Features:**
- Creates `~/bin` directory if needed
- Copies binary to `~/bin`
- Automatically updates PATH in shell config
- Persistent across terminal sessions
- Supports bash, zsh, fish, PowerShell

### `pkgmgr`
Interactive TUI package manager for editing frontend configuration.

```bash
# Launch interactive package manager
smfaman pkgmgr
```

**Features:**
- View all libraries in configuration
- Add new libraries interactively
- Edit library settings (version, CDN, files, output path)
- Delete libraries from configuration
- Edit global settings (project name, destination, default CDN)
- Save changes back to config file

**Navigation:**
- Arrow keys / Tab: Navigate between items
- Enter: Edit selected library
- `a`: Add new library
- `v` or `i`: Select version interactively
- `d`: Delete selected library
- `g`: Edit global settings
- `s`: Save and quit
- `q` / Esc: Quit without saving

### `sync`
Download libraries defined in the configuration file.

```bash
# Sync all libraries
smfaman sync

# Force re-download all files
smfaman sync --force

# Preview what would be downloaded
smfaman sync --dry-run

# Disable package file caching (download directly)
smfaman sync --no-package-cache

# Use custom config
smfaman -f myproject.yaml sync
```

**Features:**
- Smart incremental sync (only downloads missing files)
- Real-time progress bars for each download
- Package file caching (reuses downloaded files across projects)
- Respects library-specific file filters
- Creates destination directories automatically
- Uses cached CDN metadata for speed

**Package Caching:**
Downloaded library files are cached in `~/.smfaman-cache/packages/` and reused across syncs and projects. This dramatically speeds up syncing when:
- Re-syncing after deleting local files
- Using the same libraries across multiple projects
- Switching between library versions

**Progress Display:**
```
Syncing libraries... [3/15 files]

Library: jquery@3.7.1
File:    dist/jquery.min.js

[████████████████████░░░░░░░░░░░░░░░░░░░░] 52.5%
```

### `get`
Download a frontend config from a remote HTTP server.

```bash
# Download config from URL
smfaman get https://example.com/frontend.yaml

# Download to custom file
smfaman get https://example.com/config.yaml -f myproject.yaml

# Force overwrite existing file
smfaman get https://example.com/frontend.yaml --force

# Set timeout (default: 30s)
smfaman get https://slow-server.com/config.yaml --timeout 60
```

**Features:**
- Validates YAML structure before saving
- Checks required fields (destination, libraries)
- Shows config summary after download
- Suggests next steps (review, sync)

### `cache`
Manage local cache for CDN metadata and package files.

```bash
# Show cache statistics (metadata and packages)
smfaman cache stats

# Clear all cached data (metadata and packages)
smfaman cache clear

# Clear only package files (keep metadata)
smfaman cache clear-packages

# Remove only expired metadata entries
smfaman cache clean
```

**Cache Details:**
- Location: `~/.smfaman-cache/`
- Two cache types:
  - **Metadata cache**: CDN API responses (24-hour TTL)
  - **Package cache**: Downloaded library files (no expiration)
- Automatic cleanup of expired metadata
- Speeds up repeated operations and cross-project syncing

**Package Cache Benefits:**
The package cache stores downloaded library files permanently, so:
- Switching between projects using the same libraries is instant
- Re-syncing after cleaning local files doesn't re-download
- Testing different versions is faster (cached versions reused)
- Saves bandwidth and CDN API calls

## Configuration

The default configuration file is `smartfrontend.yaml`. You can specify a different file using the `-f` flag.

### Example Configuration

```yaml
destination: "./public/libs/{library_name}"
project_name: "my-awesome-project"
cdn: "unpkg"  # Default CDN for all libraries

libraries:
  jquery:
    version: "3.7.1"

  bootstrap:
    version: "5.3.0"
    cdn: "cdnjs"  # Override global CDN for this library
    files:
      - "css/bootstrap.min.css"
      - "js/bootstrap.bundle.min.js"

  react:
    version: "18.2.0"
    files:
      - "umd/react.production.min.js"
    output_path: "./custom/react"  # Custom output directory
```

### Configuration Fields

**Global Fields:**
- `destination` (required): Output path template, use `{library_name}` placeholder
- `project_name` (optional): Project identifier
- `cdn` (optional): Default CDN (unpkg, cdnjs, jsdelivr)

**Library Fields:**
- `version` (required): Specific version to download
- `cdn` (optional): Override global CDN for this library
- `files` (optional): Specific files to download (supports patterns)
- `output_path` (optional): Custom output path (overrides destination template)

## Global Configuration

Application settings can be configured in `~/.smfaman.yaml`:

```yaml
default_cdn: jsdelivr
cache_duration: 24h
verify_ssl: true
```

## Key Advantages

### Why use smfaman instead of npm?

**Perfect for:**
- Static sites with minimal JavaScript needs
- HTML/CSS projects that just need a few libraries
- Teaching environments where simplicity matters
- Projects where bundling overhead isn't justified
- Rapid prototyping without build toolchain setup

**Benefits:**
- ⚡ **No build step**: Download and use immediately
- 🪶 **No node_modules**: Keep your project lightweight (GBs → KBs)
- 🔒 **Version locked**: YAML config is your lockfile
- 🌐 **CDN-ready**: Downloaded files can be served directly or re-hosted
- 📦 **Clean separation**: Frontend assets separate from backend code
- 🔄 **Easy sharing**: One YAML file defines all dependencies

## Usage Examples

### Example 1: Simple Bootstrap Project

```bash
# Initialize config interactively
smfaman init

# Add Bootstrap CSS framework
smfaman add bootstrap@5.3.0

# Sync to local directory
smfaman sync
```

### Example 2: React Application

```bash
# Initialize
smfaman init

# Add React and ReactDOM interactively
smfaman add react -i
smfaman add react-dom -i

# Sync all libraries
smfaman sync
```

### Example 3: Exploring Versions

```bash
# Browse jQuery versions interactively
smfaman pkgver jquery -i

# Check latest React version
smfaman pkgver react --limit 5

# Compare versions across CDNs
smfaman pkgver bootstrap --cdn unpkg
smfaman pkgver bootstrap --cdn cdnjs
smfaman pkgver bootstrap --cdn jsdelivr
```

### Example 4: Team Configuration Sharing

```bash
# On machine 1: Create and configure
smfaman init
smfaman add bootstrap@5.3.0
smfaman add jquery@3.7.1

# Upload smartfrontend.yaml to your server
# On machine 2: Download and sync
smfaman get https://yourserver.com/smartfrontend.yaml
smfaman sync
```

### Example 5: Custom File Selection

```yaml
# Only download minified files
libraries:
  lodash:
    version: "4.17.21"
    files:
      - "lodash.min.js"  # Exact match

  bootstrap:
    version: "5.3.0"
    files:
      - "dist/css/"  # Prefix match - all files in dist/css/
```

## API Integration

The tool integrates with three CDN APIs:

- **UNPKG**: `https://unpkg.com/{library}@{version}/?meta`
- **CDNJS**: `https://api.cdnjs.com/libraries/{library}/{version}`
- **jsDelivr**: `https://data.jsdelivr.com/v1/packages/npm/{library}@{version}`

Each CDN provides different metadata:
- **UNPKG**: File paths, sizes, MIME types, and integrity hashes
- **CDNJS**: File lists with comprehensive SRI hash mappings
- **jsDelivr**: Hierarchical file tree with default entry points

**All API calls are cached locally for 24 hours to improve performance.**

## Development

### Building

```bash
go build -o smfaman
```

### Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Test specific package
go test ./pkgs/frontend_mgr -v

# Test specific command
go test ./cmd -v -run TestAdd

# Run benchmarks
go test ./pkgs/frontend_mgr -bench=.
```

### Project Structure

```
smfaman/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command and config
│   ├── init.go            # Initialize config file
│   ├── init_tui.go        # Bubble Tea UI for init
│   ├── add.go             # Add library command
│   ├── add_test.go        # Add command tests
│   ├── delete.go          # Delete library command
│   ├── delete_test.go     # Delete command tests
│   ├── upgrade.go         # Upgrade library command
│   ├── upgrade_test.go    # Upgrade command tests
│   ├── clean.go           # Clean library folders
│   ├── clean_test.go      # Clean command tests
│   ├── install.go         # Install binary to ~/bin
│   ├── install_test.go    # Install command tests
│   ├── pkgmgr.go          # Interactive package manager
│   ├── pkgmgr_tui.go      # TUI for package manager
│   ├── sync.go            # Sync libraries command
│   ├── sync_test.go       # Sync command tests
│   ├── pkgver.go          # List package versions
│   ├── pkgver_tui.go      # Interactive version selector
│   ├── get.go             # Download remote config
│   ├── get_test.go        # Get command tests
│   └── cache.go           # Cache management commands
├── pkgs/
│   ├── frontend_mgr/      # CDN API integration
│   │   ├── requests.go    # HTTP client functions
│   │   ├── responses.go   # Response structures
│   │   ├── versions.go    # Version fetching/sorting
│   │   └── *_test.go      # Test files
│   ├── frontend_config/   # Configuration management
│   │   ├── cfg.go         # Config structs and methods
│   │   └── cfg_test.go    # Config tests
│   └── cache/             # Cache management
│       ├── cache.go       # Cache implementation
│       └── cache_test.go  # Cache tests
├── frontend/              # Vendored frontend assets (used for testing and is ephemeral)
│   ├── bootstrap/         # Bootstrap library files (used for testing and is ephemeral)
│   ├── bootswatch/        # Bootswatch theme files (used for testing and is ephemeral)
│   └── jquery/            # jQuery library files (used for testing and is ephemeral)
├── examples/
│   └── frontend.yaml      # Example configuration
├── main.go                # Entry point
├── go.mod                 # Go modules
├── README.md              # This file
├── CLAUDE.md              # Development guide for Claude Code
└── AGENTS.md              # Repository guidelines
```

## Requirements

- Go 1.21 or higher
- Internet connection (for CDN access)

## License

MIT License - Copyright © 2025 Daniel Haanpaa

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Contributing

Contributions are welcome from both humans and AI agents (Especially Claude Code)! Please feel free to submit a Pull Request.

## Roadmap

- [x] Implement library version resolution (latest, semver ranges)
- [x] Interactive mode for browsing available files
- [x] Local cache for CDN metadata
- [x] Progress bars for downloads
- [x] Add upgrade command to check for library updates
- [x] Interactive package manager TUI
- [x] Install command for binary installation
- [x] Delete/clean commands for library management
- [ ] Support for GitHub releases as a source
- [ ] Parallel downloads for faster syncing
- [ ] Generate HTML import tags with SRI hashes
- [ ] Integrity verification during download

## Inspiration
- Microsoft LibMan: https://github.com/microsoft/libman

## Notes
- This project is not affiliated with any CDN providers.

## Project Goals
- To provide a simple CLI tool for managing frontend assets from CDNs.
- To be lightweight and fast.
- To be easy to use and extend.
- This is not a replacement for bundlers like Webpack or Parcel.
- This is not a replacement for package managers like npm or Yarn.
- This is not a replacement for vite
- To be awesome
- To spark joy


## Troubleshooting

### Common Issues

**Library not found:**
```bash
# Check if package exists on CDN
smfaman pkgver library-name

# Try different CDN
smfaman add library-name --cdn jsdelivr
smfaman add library-name --cdn cdnjs
```

**Cache issues:**
```bash
# Clear cache and retry
smfaman cache clear
smfaman add library-name
```

**Install command not updating PATH:**
```bash
# Manually reload shell configuration
source ~/.bashrc    # or ~/.zshrc, ~/.bash_profile
# Or restart your terminal
```

**Permission denied during install:**
```bash
# Install creates ~/bin which doesn't need sudo
# If you get permission errors, check directory permissions
ls -la ~/bin
```

## FAQ

**Q: Does this replace npm/yarn?**
A: No, smfaman is for simple projects that only need CDN libraries without bundling. Use npm for projects with complex build pipelines.

**Q: Can I use private/scoped packages?**
A: Yes, scoped packages work: `smfaman add @babel/core@7.22.0`

**Q: Where are files downloaded?**
A: To the destination specified in your config file (default: `./frontend/{library_name}`)

**Q: Can I download specific files only?**
A: Yes, use the `--files` flag or specify `files:` in your config YAML.

**Q: How do I update all libraries?**
A: Run `smfaman upgrade` to upgrade all libraries to their latest versions.

**Q: Is the cache safe to delete?**
A: Yes, run `smfaman cache clear`. It only contains CDN metadata and will be rebuilt as needed.

**Q: Can I use this in CI/CD?**
A: Absolutely! The config file is version-controllable, and `smfaman sync` downloads all dependencies.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/yourusername/smfaman).

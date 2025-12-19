# Smart Frontend Asset Manager (smfaman)

A CLI tool for managing frontend assets from CDNs when you don't need bundling or the full infrastructure of a JavaScript SPA.

**smfaman** helps you manage frontend libraries (Bootstrap, React, jQuery, etc.) by fetching them directly from CDNs, validating integrity hashes, and keeping your frontend dependencies organized without npm or bundlers.

## Features

- 🚀 Fetch frontend libraries from multiple CDNs
- 🔒 Automatic SRI (Subresource Integrity) hash verification
- 📦 Support for three major CDNs: jsDelivr, UNPKG, and CDNJS
- 🔍 Browse available files and versions
- ⚡ Fast and lightweight - no Node.js required
- 📝 Simple YAML configuration

## Supported CDNs

| CDN | Speed* | Files | Features |
|-----|--------|-------|----------|
| **jsDelivr** | ~57ms | Hierarchical | Stats API, entrypoints, fastest |
| **UNPKG** | ~76ms | Flat list | Complete metadata, file sizes, MIME types |
| **CDNJS** | ~116ms | Flat list | Comprehensive SRI hashes |

*Benchmark based on fetching Bootstrap 5.3.0 metadata

## Installation

### From Source

```bash
git clone https://github.com/yourusername/smfaman
cd smfaman
go build -o smfaman
sudo mv smfaman /usr/local/bin/
```

### Using Go Install

```bash
go install nexus-sds.com/smfaman@latest
```

## Quick Start

```bash
# Initialize a new frontend config
smfaman init

# Add a library to your configuration
smfaman add bootstrap@5.3.0

# Sync all configured libraries to local directory
smfaman sync

# Use custom config file
smfaman -f myproject.yaml sync
```

## Commands

### `init`
Create a new smart frontend asset configuration file.

```bash
smfaman init
```

Creates `smartfrontend.yaml` in the current directory.

### `add`
Add a new library to the configuration.

```bash
smfaman add <library>@<version>
smfaman add bootstrap@5.3.0
smfaman add react@18.2.0
```

### `sync`
Download libraries defined in the configuration file but not present locally.

```bash
smfaman sync
smfaman -f custom-config.yaml sync
```

## Configuration

The default configuration file is `smartfrontend.yaml`. You can specify a different file using the `-f` flag.

### Example Configuration

```yaml
libraries:
  - name: bootstrap
    version: 5.3.0
    cdn: jsdelivr
    files:
      - dist/css/bootstrap.min.css
      - dist/js/bootstrap.bundle.min.js

  - name: react
    version: 18.2.0
    cdn: unpkg
    files:
      - umd/react.production.min.js
      - umd/react-dom.production.min.js

  - name: jquery
    version: 3.7.1
    cdn: cdnjs
    files:
      - jquery.min.js

output:
  directory: ./public/vendor
  integrity: true  # Generate SRI hashes
```

## Global Configuration

Application settings can be configured in `~/.smfaman.yaml`:

```yaml
default_cdn: jsdelivr
cache_duration: 24h
verify_ssl: true
```

## Usage Examples

### Example 1: Simple Bootstrap Project

```bash
# Initialize config
smfaman init

# Add Bootstrap CSS framework
smfaman add bootstrap@5.3.0

# Sync to local directory
smfaman sync
```

### Example 2: React Application

```bash
# Add React and ReactDOM
smfaman add react@18.2.0
smfaman add react-dom@18.2.0

# Sync all libraries
smfaman sync
```

### Example 3: Using Multiple CDNs

```yaml
libraries:
  - name: bootstrap
    version: 5.3.0
    cdn: jsdelivr  # Fastest for Bootstrap

  - name: lodash
    version: 4.17.21
    cdn: unpkg     # Good metadata

  - name: jquery
    version: 3.7.1
    cdn: cdnjs     # Reliable jQuery host
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

# Run benchmarks
go test ./pkgs/frontend_mgr -bench=.
```

### Project Structure

```
smfaman/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command and config
│   ├── init.go            # Initialize config file
│   ├── add.go             # Add library command
│   └── sync.go            # Sync libraries command
├── pkgs/
│   ├── frontend_mgr/      # CDN API integration
│   │   ├── requests.go    # HTTP client functions
│   │   ├── responses.go   # Response structures
│   │   └── *_test.go      # Test files
│   └── frontend_config/   # Configuration management
├── main.go                # Entry point
└── go.mod                 # Go modules
```

## Requirements

- Go 1.25.4 or higher
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

- [ ] Implement library version resolution (latest, semver ranges)
- [ ] Add update command to check for library updates
- [ ] Support for GitHub releases as a source
- [ ] Parallel downloads for faster syncing
- [ ] Generate HTML import tags with SRI hashes
- [ ] Interactive mode for browsing available files
- [ ] Local cache for CDN metadata

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/yourusername/smfaman).

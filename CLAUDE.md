# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**smfaman** (Smart Frontend Asset Manager) is a CLI tool for managing frontend assets from CDNs (jsDelivr, UNPKG, CDNJS) when you don't need bundling or the full infrastructure of a JavaScript SPA.

Module path: `nexus-sds.com/smfaman`

## Build & Test Commands

```bash
# Build the project
go build -o smfaman

# Build to a specific location
go build -o /tmp/smfaman

# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run specific test by name pattern
go test ./pkgs/frontend_mgr -v -run "TestFetchUnpkg"

# Run tests in a specific package
go test ./pkgs/frontend_mgr -v

# Run specific command tests
go test ./cmd -v -run "TestAdd"
go test ./cmd -v -run "TestSync"

# Run benchmarks
go test ./pkgs/frontend_mgr -bench=. -benchtime=3x

# Run the CLI directly
go run main.go [command]

# Test specific commands
go run main.go init
go run main.go add react --interactive
go run main.go pkgver jquery -i
go run main.go sync --dry-run
```

## Architecture

### CLI Structure (Cobra-based)

The project uses **Cobra** for CLI framework with commands in `cmd/`:
- `root.go` - Base command and configuration initialization
- `init.go` + `init_tui.go` - Interactive config file creation (Bubble Tea TUI)
- `add.go` + `add_test.go` - Add library to configuration with version validation
- `sync.go` + `sync_test.go` - Downloads libraries with progress bars
- `pkgver.go` + `pkgver_tui.go` - List/browse package versions (interactive TUI)
- `get.go` + `get_test.go` - Download remote config files
- `cache.go` - Cache management (stats, clear, clean)

Configuration is managed via **Viper**:
- Default config: `$HOME/.smfaman.yaml`
- Frontend config (via `-f` flag): `smartfrontend.yaml` (default)

### Package Structure

**`pkgs/frontend_mgr/`** - CDN API integration layer
- `requests.go` - HTTP client functions for fetching from CDNs (with caching)
- `responses.go` - Response structs for all three CDN APIs
- `versions.go` - Version fetching and semantic version sorting
- `*_test.go` - Test files

**`pkgs/frontend_config/`** - Configuration management
- `cfg.go` - FrontendConfig and LibraryConfig structs with YAML serialization
- Methods: GetLibraryDestination, GetLibraryVersions, GetLibraryCDN, etc.

**`pkgs/cache/`** - Local caching system
- `cache.go` - Cache manager with TTL support (default: 24 hours)
- Location: `~/.smfaman-cache/`
- SHA256-based cache keys

### CDN API Integration

Three CDN providers are supported with dedicated functions in `pkgs/frontend_mgr/requests.go`:

1. **UNPKG** - `FetchUnpkgMeta(libraryName, version string)`
   - Endpoint: `https://unpkg.com/{library}@{version}/?meta`
   - Returns: Flat file list with sizes, types, and integrity hashes
   - Also: `FetchUnpkgVersions(libraryName string)` for version listing

2. **CDNJS** - `FetchCdnjsVersion(libraryName, version string)`
   - Endpoint: `https://api.cdnjs.com/libraries/{library}/{version}`
   - Returns: File lists with SRI hashes (map[string]string)
   - Also: `FetchCdnjsVersions(libraryName string)` for version listing

3. **jsDelivr** - `FetchJsdelivrPackage(libraryName, version string)`
   - Endpoint: `https://data.jsdelivr.com/v1/packages/npm/{library}@{version}`
   - Returns: Hierarchical file tree with recursive `JsdelivrFile` structs
   - Also: `FetchJsdelivrVersions(libraryName string)` for version listing

**All CDN request functions use the cache manager automatically.**

### Response Structure Differences

**Key architectural difference**: jsDelivr uses a recursive/hierarchical structure (`Files []JsdelivrFile` can contain nested `Files`), while UNPKG and CDNJS use flat file arrays. When traversing jsDelivr responses, you must recursively walk the file tree.

**UNPKG** provides the most metadata per file (size, type, integrity, path).
**CDNJS** provides separate `Files` array and `SRI` map for integrity lookups.
**jsDelivr** includes related API endpoints via the `Links` struct (stats, entrypoints).

### Version Management

`pkgs/frontend_mgr/versions.go`:
- `SortVersions(versions []string)` - Sorts versions using semantic versioning (hashicorp/go-version)
- Handles pre-releases, build metadata, and invalid versions
- Returns versions in descending order (newest first)

### Caching System

All CDN API calls are automatically cached:
- Cache manager created in `pkgs/cache/cache.go`
- Default TTL: 24 hours
- Cache keys generated with `GenerateKey(components ...string)`
- Commands can disable cache with `--no-cache` flag
- Cache location: `~/.smfaman-cache/`

To integrate caching in new CDN functions:
```go
cacheKey := cache.GenerateKey("cdn-name", packageName, version)
var result ResponseType
found, err := cacheManager.Get(cacheKey, &result)
if err == nil && found {
    return &result, nil
}
// ... fetch from CDN ...
cacheManager.Set(cacheKey, result)
```

### Interactive TUI Components

The project uses **Bubble Tea** for interactive interfaces:

1. **Init Command** (`cmd/init_tui.go`)
   - Text inputs for project name, destination, CDN selection
   - Navigation with Tab/Shift+Tab
   - Validation before saving

2. **Package Version Selector** (`cmd/pkgver_tui.go`)
   - List interface with search/filter
   - Highlights latest version
   - Used by both `pkgver --interactive` and `add --interactive`

3. **Sync Progress** (`cmd/sync.go`)
   - Real-time progress bars
   - Shows current file being downloaded
   - Overall progress counter

### Testing Patterns

Test files demonstrate the proper usage patterns:
- `requests_test.go` - Basic functionality tests for each CDN
- `cdn_test.go` - Cross-CDN validation with bootstrap/bootswatch
- `versions_test.go` - Version fetching and sorting tests
- `add_test.go` - Add command tests (package parsing, config loading)
- `sync_test.go` - Sync command tests (file filtering, task building)
- `get_test.go` - HTTP download and validation tests
- `cache_test.go` - Cache operations with TTL

When adding new CDN functions, test with multiple libraries (e.g., react, bootstrap, bootswatch) to ensure compatibility across different package structures.

## Command Implementation Patterns

### Package Parsing (cmd/add.go:166-188)

Handles both regular and scoped packages:
```go
func parsePackageSpec(spec string) (name, version string) {
    // Handles: react@18.2.0, @babel/core@7.22.0
    if strings.HasPrefix(spec, "@") {
        // Find second @ for scoped packages
        idx := strings.Index(spec[1:], "@")
        if idx == -1 {
            return spec, ""
        }
        return spec[:idx+1], spec[idx+2:]
    }
    // Regular packages
    parts := strings.SplitN(spec, "@", 2)
    // ...
}
```

### Config Loading (cmd/add.go:261-286)

Standard pattern for loading frontend config:
```go
func loadConfig(path string) (*frontend_config.FrontendConfig, error) {
    // Check existence
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, fmt.Errorf("config file '%s' does not exist", path)
    }
    // Read and unmarshal
    data, _ := os.ReadFile(path)
    var config frontend_config.FrontendConfig
    yaml.Unmarshal(data, &config)
    // Initialize Libraries map if nil
    if config.Libraries == nil {
        config.Libraries = make(map[string]frontend_config.LibraryConfig)
    }
    return &config, nil
}
```

### File Filtering (cmd/sync.go:244-259)

Pattern matching for file selection:
```go
func filterFiles(files []CDNFile, patterns []string) []CDNFile {
    // Supports exact match: "dist/jquery.min.js"
    // Supports prefix match: "dist/" matches all files in dist/
    for _, file := range files {
        for _, pattern := range patterns {
            if file.Path == pattern || strings.HasPrefix(file.Path, pattern) {
                filtered = append(filtered, file)
            }
        }
    }
}
```

### Recursive File Collection (cmd/sync.go:222-242)

For jsDelivr's hierarchical structure:
```go
func collectJsdelivrFiles(libName, version string, jsFiles []frontend_mgr.JsdelivrFile, basePath string) []CDNFile {
    for _, f := range jsFiles {
        path := filepath.Join(basePath, f.Name)
        if f.Type == "file" {
            files = append(files, CDNFile{...})
        } else if f.Type == "directory" && len(f.Files) > 0 {
            // Recursive call for subdirectories
            files = append(files, collectJsdelivrFiles(libName, version, f.Files, path)...)
        }
    }
}
```

## Development Notes

### Import Path

Always use the full module path for internal imports:
```go
import "nexus-sds.com/smfaman/cmd"
import "nexus-sds.com/smfaman/pkgs/frontend_mgr"
import "nexus-sds.com/smfaman/pkgs/frontend_config"
import "nexus-sds.com/smfaman/pkgs/cache"
```

### Error Handling Pattern

CDN request functions follow this pattern:
- Return `(*ResponseType, error)`
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Check HTTP status codes and include response body in errors
- Always defer `resp.Body.Close()`

### Cobra Command Pattern

Commands are initialized in their `init()` function:
```go
func init() {
    rootCmd.AddCommand(commandCmd)
    // Add flags here
    commandCmd.Flags().StringVar(&flagVar, "flag-name", "default", "description")
}
```

The `FrontendConfig` variable in `cmd/root.go` is accessible to all commands as a persistent flag value.

### Bubble Tea TUI Pattern

For interactive interfaces:
```go
type myModel struct {
    // State fields
}

func (m myModel) Init() tea.Cmd { return nil }

func (m myModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle key presses
    }
    return m, nil
}

func (m myModel) View() string {
    // Render UI
}
```

### Configuration Structure

Frontend config YAML structure:
```yaml
destination: "./public/libs/{library_name}"  # Required, supports {library_name} template
project_name: "my-project"                   # Optional
cdn: "unpkg"                                 # Optional global default

libraries:
  library-name:
    version: "1.0.0"                         # Required
    cdn: "cdnjs"                             # Optional, overrides global
    files:                                   # Optional, filter which files to download
      - "dist/file.min.js"
    output_path: "./custom/path"             # Optional, overrides destination
```

## Common Operations

### Adding a New Command

1. Create `cmd/commandname.go` with cobra command structure
2. Add to `rootCmd` in `init()` function
3. Implement command logic in `Run` function
4. Create `cmd/commandname_test.go` with tests
5. Update README.md and CLAUDE.md

### Adding a New CDN Function

1. Add function to `pkgs/frontend_mgr/requests.go`
2. Create corresponding response struct in `responses.go`
3. Integrate cache: check cache first, set cache after fetch
4. Add tests in `pkgs/frontend_mgr/*_test.go`
5. Test with multiple libraries (react, bootstrap, bootswatch)

### Adding a New Config Field

1. Update structs in `pkgs/frontend_config/cfg.go`
2. Add YAML tags
3. Create helper methods if needed
4. Add tests in `cfg_test.go`
5. Update example in `examples/frontend.yaml`

## Key Dependencies

- **Cobra**: CLI framework (`github.com/spf13/cobra`)
- **Viper**: Configuration management (`github.com/spf13/viper`)
- **Bubble Tea**: Terminal UI (`github.com/charmbracelet/bubbletea`)
- **Bubbles**: TUI components (`github.com/charmbracelet/bubbles`)
- **Lipgloss**: TUI styling (`github.com/charmbracelet/lipgloss`)
- **go-version**: Semantic versioning (`github.com/hashicorp/go-version`)
- **yaml.v3**: YAML parsing (`gopkg.in/yaml.v3`)

## Debugging Tips

### Testing Interactive Commands

Use dry-run or non-interactive modes:
```bash
smfaman sync --dry-run
smfaman add react@18.2.0  # Non-interactive
```

### Cache Issues

Clear cache during development:
```bash
smfaman cache clear
```

### Testing Network Calls

Tests that make network calls should:
- Use `testing.Short()` to skip in short mode
- Handle network errors gracefully with `t.Skipf()`
- Use `t.Logf()` for informational output

## Project Status

**Implemented:**
- ✅ Interactive config initialization (Bubble Tea)
- ✅ Add library with version validation
- ✅ Interactive version selector
- ✅ Package version listing
- ✅ Download remote configs
- ✅ Sync with progress bars
- ✅ Local caching (24h TTL)
- ✅ Cache management commands
- ✅ Support for all three CDNs
- ✅ Semantic version sorting
- ✅ Scoped package support
- ✅ File filtering

**Future Enhancements:**
- Update command (check for newer versions)
- Parallel downloads
- SRI hash generation
- GitHub releases support

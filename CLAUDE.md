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

# Run benchmarks
go test ./pkgs/frontend_mgr -bench=. -benchtime=3x

# Run the CLI directly
go run main.go [command]
```

## Architecture

### CLI Structure (Cobra-based)

The project uses **Cobra** for CLI framework with commands in `cmd/`:
- `root.go` - Base command and configuration initialization
- `init.go` - Creates new frontend asset config file
- `add.go` - Adds library to configuration
- `sync.go` - Downloads libraries defined in config but not present locally

Configuration is managed via **Viper**:
- Default config: `$HOME/.smfaman.yaml`
- Frontend config (via `-f` flag): `smartfrontend.yaml` (default)

### Package Structure

**`pkgs/frontend_mgr/`** - CDN API integration layer
- `requests.go` - HTTP client functions for fetching from CDNs
- `responses.go` - Response structs for all three CDN APIs
- `*_test.go` - Test files

**`pkgs/frontend_config/`** - Configuration management (placeholder)

### CDN API Integration

Three CDN providers are supported with dedicated functions in `pkgs/frontend_mgr/requests.go`:

1. **UNPKG** - `FetchUnpkgMeta(libraryName, version string)`
   - Endpoint: `https://unpkg.com/{library}@{version}/?meta`
   - Returns: Flat file list with sizes, types, and integrity hashes

2. **CDNJS** - `FetchCdnjsVersion(libraryName, version string)`
   - Endpoint: `https://api.cdnjs.com/libraries/{library}/{version}`
   - Returns: File lists with SRI hashes (map[string]string)

3. **jsDelivr** - `FetchJsdelivrPackage(libraryName, version string)`
   - Endpoint: `https://data.jsdelivr.com/v1/packages/npm/{library}@{version}`
   - Returns: Hierarchical file tree with recursive `JsdelivrFile` structs

### Response Structure Differences

**Key architectural difference**: jsDelivr uses a recursive/hierarchical structure (`Files []JsdelivrFile` can contain nested `Files`), while UNPKG and CDNJS use flat file arrays. When traversing jsDelivr responses, you must recursively walk the file tree.

**UNPKG** provides the most metadata per file (size, type, integrity, path).
**CDNJS** provides separate `Files` array and `SRI` map for integrity lookups.
**jsDelivr** includes related API endpoints via the `Links` struct (stats, entrypoints).

### Testing Patterns

Test files demonstrate the proper usage patterns:
- `requests_test.go` - Basic functionality tests for each CDN
- `cdn_test.go` - Cross-CDN validation with bootstrap/bootswatch

When adding new CDN functions, test with multiple libraries (e.g., react, bootstrap, bootswatch) to ensure compatibility across different package structures.

## Development Notes

### Import Path

Always use the full module path for internal imports:
```go
import "nexus-sds.com/smfaman/cmd"
import "nexus-sds.com/smfaman/pkgs/frontend_mgr"
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
}
```

The `FrontendConfig` variable in `cmd/root.go` is accessible to all commands as a persistent flag value.

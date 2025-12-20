# CDN Comparison Guide

This document provides detailed information about the three supported CDNs in smfaman and their differences.

## Supported CDNs

smfaman supports three Content Delivery Networks (CDNs) for downloading frontend libraries:

1. **UNPKG** - `cdn: unpkg`
2. **CDNJS** - `cdn: cdnjs`
3. **jsDelivr** - `cdn: jsdelivr`

## Quick Comparison

| Feature | UNPKG | CDNJS | jsDelivr |
|---------|-------|-------|----------|
| **Package Source** | Full npm packages | Curated distribution files | Full npm packages |
| **File Count** | High | Low | High |
| **Source Files** | ✅ Yes | ❌ No | ✅ Yes |
| **Build Files** | ✅ Yes | ❌ No | ✅ Yes |
| **Documentation** | ✅ Yes | ❌ No | ✅ Yes |
| **Distribution Only** | ❌ No | ✅ Yes | ❌ No |
| **Update Frequency** | Real-time from npm | Manual curation | Real-time from npm |
| **Response Type** | Flat file list | Flat file list | Hierarchical tree |

## File Count Comparison

Based on testing with common libraries (Bootstrap 5.3.8, Bootswatch 5.3.8, jQuery 3.7.1):

| Library | UNPKG | CDNJS | jsDelivr |
|---------|-------|-------|----------|
| **Bootstrap 5.3.8** | 219 files | 136 files | 219 files |
| **Bootswatch 5.3.8** | 211 files | 156 files | 211 files |
| **jQuery 3.7.1** | 125 files | 6 files | 125 files |
| **Total** | **555 files** | **298 files** | **555 files** |

## Detailed CDN Information

### UNPKG

**Website:** https://unpkg.com
**API Endpoint:** `https://unpkg.com/{package}@{version}/?meta`

**Characteristics:**
- Mirrors the **entire npm package** exactly as published
- Includes all files: source code, tests, documentation, build scripts, etc.
- Real-time updates when new versions are published to npm
- Type field contains **MIME types** (e.g., `text/javascript`, `text/css`, `application/json`)
- Returns a flat list of all files in the package

**Best for:**
- Development environments where you need access to source files
- Projects that need documentation or examples from the package
- When you want the complete package structure
- Debugging with source maps and unminified code

**Example files included:**
- `dist/bootstrap.min.js` (production)
- `dist/bootstrap.js` (unminified)
- `js/src/` (source files)
- `package.json`, `README.md`, `LICENSE`
- Build configuration files

### CDNJS

**Website:** https://cdnjs.com
**API Endpoint:** `https://api.cdnjs.com/libraries/{library}/{version}`

**Characteristics:**
- Hosts **only curated distribution files** ready for production use
- Manually curated and reviewed by the CDNJS team
- Does not include source code, tests, or documentation
- Provides both `files` and `rawFiles` arrays (typically identical)
- Includes SRI (Subresource Integrity) hashes for security
- May lag behind npm releases due to manual curation process

**Best for:**
- Production websites that only need minified/compiled files
- Projects prioritizing file size and bandwidth
- When you need SRI hashes for security
- Simpler project structure with fewer files

**Example files included:**
- `css/bootstrap.min.css` (minified CSS)
- `css/bootstrap.css` (unminified CSS)
- `js/bootstrap.bundle.min.js` (minified JS)
- `*.map` (source maps)

**Example files NOT included:**
- Source files (`js/src/`, `scss/`)
- Documentation (`README.md`, `docs/`)
- Package metadata (`package.json`)
- Build configuration

### jsDelivr

**Website:** https://www.jsdelivr.com
**API Endpoint:** `https://data.jsdelivr.com/v1/packages/npm/{package}@{version}`

**Characteristics:**
- Mirrors the **entire npm package** like UNPKG
- Returns files in a **hierarchical/recursive tree structure**
- Real-time updates from npm
- Includes file hashes for integrity verification
- Multi-CDN infrastructure with automatic failover
- Provides additional metadata via `links` (stats, entrypoints)

**Best for:**
- Production sites that want CDN reliability with full package access
- Projects needing both source and distribution files
- When you want automatic CDN failover
- Access to package statistics and metadata

**Example files included:**
- Same comprehensive file set as UNPKG
- Complete package structure with nested directories
- All source files, documentation, and build files

## API Response Structure Differences

### UNPKG Response
```json
{
  "files": [
    {
      "path": "/dist/jquery.min.js",
      "type": "text/javascript",  // MIME type, not "file"
      "size": 89476,
      "integrity": "sha384-..."
    }
  ]
}
```
- **Flat array** of files
- Type field is **MIME type** (e.g., `text/javascript`, `text/css`)

### CDNJS Response
```json
{
  "files": ["jquery.js", "jquery.min.js"],
  "rawFiles": ["jquery.js", "jquery.min.js"],
  "sri": {
    "jquery.min.js": "sha512-..."
  }
}
```
- **Flat array** of file paths
- Separate SRI hash map
- Both `files` and `rawFiles` arrays (usually identical)

### jsDelivr Response
```json
{
  "files": [
    {
      "name": "dist",
      "type": "directory",
      "files": [
        {
          "name": "jquery.min.js",
          "type": "file",
          "size": 89476,
          "hash": "..."
        }
      ]
    }
  ]
}
```
- **Hierarchical tree** structure
- Requires recursive traversal
- Type field is `"file"` or `"directory"`

## Choosing the Right CDN

### Use **UNPKG** when:
- ✅ You need source files for debugging
- ✅ You want documentation and examples
- ✅ You prefer a simple API response (flat list)
- ✅ You need the latest npm versions immediately

### Use **CDNJS** when:
- ✅ You only need production-ready distribution files
- ✅ You want a smaller, cleaner project structure
- ✅ You need SRI hashes for security
- ✅ Bandwidth and file count are concerns
- ✅ You trust manually curated packages

### Use **jsDelivr** when:
- ✅ You need high availability with multi-CDN failover
- ✅ You want both source and distribution files
- ✅ You need package statistics and metadata
- ✅ You prefer the latest npm versions immediately
- ✅ You want file integrity hashes

## Configuration Examples

### Using UNPKG (default)
```yaml
destination: ./frontend/{library_name}
cdn: unpkg
libraries:
  bootstrap:
    version: 5.3.8
```

### Using CDNJS
```yaml
destination: ./frontend/{library_name}
cdn: cdnjs
libraries:
  bootstrap:
    version: 5.3.8
```

### Using jsDelivr
```yaml
destination: ./frontend/{library_name}
cdn: jsdelivr
libraries:
  bootstrap:
    version: 5.3.8
```

### Per-Library CDN Override
```yaml
destination: ./frontend/{library_name}
cdn: unpkg  # Global default
libraries:
  bootstrap:
    version: 5.3.8
    cdn: cdnjs  # Override for this library
  jquery:
    version: 3.7.1
    cdn: jsdelivr  # Override for this library
```

## File Filtering

Since UNPKG and jsDelivr include many files, you may want to filter them:

```yaml
destination: ./frontend/{library_name}
cdn: unpkg
libraries:
  bootstrap:
    version: 5.3.8
    files:
      - dist/css/  # Only download CSS files
      - dist/js/   # Only download JS files
```

With CDNJS, filtering is often unnecessary since it only includes distribution files.

## Performance Considerations

| CDN | File Count | Download Size | Sync Time |
|-----|-----------|---------------|-----------|
| **UNPKG** | High (555+) | Large | Slower |
| **CDNJS** | Low (298) | Small | Faster |
| **jsDelivr** | High (555+) | Large | Slower |

**Note:** smfaman's package cache significantly improves sync times on subsequent runs.

## Version Availability

- **UNPKG**: All npm versions available immediately
- **CDNJS**: Only curated versions (may lag behind npm)
- **jsDelivr**: All npm versions available immediately

To check available versions:
```bash
smfaman pkgver bootstrap -i
```

## Known Issues and Limitations

### UNPKG
- Type field contains MIME types, not "file"/"directory" (fixed in smfaman)
- No built-in SRI hash support in API

### CDNJS
- Fewer files (distribution only)
- May not have latest versions immediately
- Some packages may not be available

### jsDelivr
- Hierarchical API requires recursive parsing
- More complex response structure

## Migration Between CDNs

To switch CDNs, simply update your config and re-sync:

```bash
# Edit smartfrontend.yaml, change cdn: unpkg to cdn: cdnjs
smfaman sync --force
```

The `--force` flag ensures all files are re-downloaded from the new CDN.

## Testing CDN Differences

To compare CDNs for your libraries:

```bash
# Test with UNPKG
smfaman -f test-unpkg.yaml sync --dry-run

# Test with CDNJS
smfaman -f test-cdnjs.yaml sync --dry-run

# Test with jsDelivr
smfaman -f test-jsdelivr.yaml sync --dry-run
```

## Additional Resources

- UNPKG Documentation: https://unpkg.com
- CDNJS Documentation: https://cdnjs.com
- jsDelivr Documentation: https://www.jsdelivr.com
- npm Registry: https://www.npmjs.com

## Version History

- **2024-12**: Updated with accurate file counts and CDN comparison data
- **2024-12**: Fixed UNPKG Type field bug (MIME type vs "file")

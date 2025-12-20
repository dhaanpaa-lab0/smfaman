package frontend_mgr

// UnpkgMetaResponse represents the response from https://unpkg.com/{library_name}@{version}/?meta
type UnpkgMetaResponse struct {
	Package string      `json:"package"`
	Version string      `json:"version"`
	Prefix  string      `json:"prefix"`
	Files   []UnpkgFile `json:"files"`
}

// UnpkgFile represents a file entry in the UNPKG meta response
type UnpkgFile struct {
	Path      string `json:"path"`
	Size      int    `json:"size"`
	Type      string `json:"type"`
	Integrity string `json:"integrity"`
}

// CdnjsVersionResponse represents the response from https://api.cdnjs.com/libraries/{library_name}/{version}
type CdnjsVersionResponse struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	RawFiles []string          `json:"rawFiles"`
	Files    []string          `json:"files"`
	SRI      map[string]string `json:"sri"` // Map of filename to SRI hash
}

// JsdelivrPackageResponse represents the response from https://data.jsdelivr.com/v1/packages/npm/{library_name}@{version}
type JsdelivrPackageResponse struct {
	Type    string            `json:"type"`    // Package type (e.g., "npm")
	Name    string            `json:"name"`    // Package name
	Version string            `json:"version"` // Package version
	Default string            `json:"default"` // Default/main file path
	Files   []JsdelivrFile    `json:"files"`   // File tree structure
	Links   JsdelivrLinks     `json:"links"`   // Related API endpoints
}

// JsdelivrFile represents a file or directory entry in the jsDelivr response
type JsdelivrFile struct {
	Type  string           `json:"type"`           // "file" or "directory"
	Name  string           `json:"name"`           // File or directory name
	Hash  string           `json:"hash,omitempty"` // File hash (only for files)
	Size  int              `json:"size,omitempty"` // File size in bytes (only for files)
	Files []JsdelivrFile   `json:"files,omitempty"` // Nested files (only for directories)
}

// JsdelivrLinks contains URLs to related jsDelivr API endpoints
type JsdelivrLinks struct {
	Stats       string `json:"stats"`       // URL to package stats endpoint
	Entrypoints string `json:"entrypoints"` // URL to package entrypoints endpoint
}

// CdnjsLibraryResponse represents the response from https://api.cdnjs.com/libraries/{library}
// This endpoint returns library information including all available versions
type CdnjsLibraryResponse struct {
	Name        string   `json:"name"`
	Latest      string   `json:"latest"`      // URL to latest version
	Version     string   `json:"version"`     // Latest version number
	Description string   `json:"description"` // Package description
	Homepage    string   `json:"homepage"`    // Project homepage URL
	Repository  struct {
		Type string `json:"type"` // Repository type (e.g., "git")
		URL  string `json:"url"`  // Repository URL
	} `json:"repository"`
	Versions []string `json:"versions"` // All available versions
}

// JsdelivrVersionsResponse represents the response from https://data.jsdelivr.com/v1/packages/npm/{library}
// This endpoint returns package information including available versions
type JsdelivrVersionsResponse struct {
	Type     string                 `json:"type"`     // Package type (e.g., "npm")
	Name     string                 `json:"name"`     // Package name
	Tags     map[string]string      `json:"tags"`     // Version tags (e.g., "latest": "1.2.3")
	Versions []JsdelivrVersionInfo  `json:"versions"` // Available versions
	Links    JsdelivrVersionsLinks  `json:"links"`    // Related API endpoints
}

// JsdelivrVersionInfo represents version information in jsDelivr response
type JsdelivrVersionInfo struct {
	Version string                `json:"version"` // Version number
	Links   JsdelivrVersionLinks  `json:"links"`   // Links to version-specific endpoints
}

// JsdelivrVersionLinks contains URLs to version-specific jsDelivr endpoints
type JsdelivrVersionLinks struct {
	Self string `json:"self"` // URL to this version's endpoint
	Stats string `json:"stats,omitempty"` // URL to version stats
}

// JsdelivrVersionsLinks contains URLs to package-level jsDelivr endpoints
type JsdelivrVersionsLinks struct {
	Self string `json:"self"` // URL to this package's endpoint
}

// UnpkgPackageResponse represents the response from https://registry.npmjs.org/{package}
// UNPKG doesn't have its own versions API, so we use npm registry
type UnpkgPackageResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	DistTags    map[string]string `json:"dist-tags"` // Version tags (e.g., "latest": "1.2.3")
	Versions    map[string]struct {
		Version string `json:"version"`
	} `json:"versions"` // Map of version number to version info
}

// CdnjsSearchResponse represents the response from https://api.cdnjs.com/libraries?search={query}
type CdnjsSearchResponse struct {
	Results []CdnjsSearchResult `json:"results"`
	Total   int                 `json:"total"`
}

// CdnjsSearchResult represents a single search result from CDNJS
type CdnjsSearchResult struct {
	Name        string `json:"name"`
	Latest      string `json:"latest"`      // URL to latest version
	Description string `json:"description"` // Package description
	Version     string `json:"version"`     // Latest version number
	Homepage    string `json:"homepage"`    // Project homepage URL
	Keywords    []string `json:"keywords,omitempty"` // Package keywords
}

// NpmSearchResponse represents the response from npm registry search
// Used for UNPKG and jsDelivr package searches
type NpmSearchResponse struct {
	Objects []NpmSearchObject `json:"objects"`
	Total   int               `json:"total"`
	Time    string            `json:"time"`
}

// NpmSearchObject represents a single search result from npm registry
type NpmSearchObject struct {
	Package NpmPackageInfo `json:"package"`
	Score   NpmScore       `json:"score"`
}

// NpmPackageInfo contains package information from npm search
type NpmPackageInfo struct {
	Name        string            `json:"name"`
	Scope       string            `json:"scope,omitempty"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords,omitempty"`
	Date        string            `json:"date"`
	Links       NpmPackageLinks   `json:"links"`
	Publisher   NpmPublisher      `json:"publisher"`
}

// NpmPackageLinks contains URLs for npm package
type NpmPackageLinks struct {
	Npm        string `json:"npm"`
	Homepage   string `json:"homepage,omitempty"`
	Repository string `json:"repository,omitempty"`
	Bugs       string `json:"bugs,omitempty"`
}

// NpmPublisher contains publisher information
type NpmPublisher struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// NpmScore contains scoring information for search results
type NpmScore struct {
	Final  float64 `json:"final"`
	Detail struct {
		Quality     float64 `json:"quality"`
		Popularity  float64 `json:"popularity"`
		Maintenance float64 `json:"maintenance"`
	} `json:"detail"`
}

// SearchResult is a unified search result structure across all CDNs
type SearchResult struct {
	Name        string
	Version     string
	Description string
	Homepage    string
	Keywords    []string
	CDN         string // Which CDN this result came from
}

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

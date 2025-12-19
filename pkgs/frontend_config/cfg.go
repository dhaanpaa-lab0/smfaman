package frontend_config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CDN represents supported CDN providers
type CDN string

const (
	// CDNUnpkg represents the UNPKG CDN (https://unpkg.com)
	CDNUnpkg CDN = "unpkg"

	// CDNCdnjs represents the CDNJS CDN (https://cdnjs.com)
	CDNCdnjs CDN = "cdnjs"

	// CDNJsdelivr represents the jsDelivr CDN (https://www.jsdelivr.com)
	CDNJsdelivr CDN = "jsdelivr"
)

// FrontendConfig represents the top-level configuration for frontend asset management
type FrontendConfig struct {
	// Destination is the output path template for downloaded libraries
	// Supports {library_name} placeholder (e.g., "./frontend/{library_name}")
	Destination string `yaml:"destination"`

	// ProjectName is an identifier for the project
	ProjectName string `yaml:"project_name"`

	// CDN specifies the default CDN to use for all libraries
	// Valid values: "unpkg", "cdnjs", "jsdelivr"
	// Individual libraries can override this with their own CDN setting
	CDN CDN `yaml:"cdn,omitempty"`

	// Libraries is a map where the key is the library name (e.g., "jquery", "bootstrap")
	// and the value contains the library configuration
	Libraries map[string]LibraryConfig `yaml:"libraries"`
}

// LibraryConfig represents configuration for a single library
type LibraryConfig struct {
	// Version specifies the library version to fetch (e.g., "3.5.1", "4.5.2")
	Version string `yaml:"version"`

	// CDN specifies which CDN to use: "unpkg", "cdnjs", or "jsdelivr"
	// If empty, the global CDN setting from FrontendConfig will be used
	CDN CDN `yaml:"cdn,omitempty"`

	// Files specifies which files to download from the library
	// If empty, all files or a default set will be downloaded
	Files []string `yaml:"files,omitempty"`

	// OutputPath allows overriding the global Destination for this specific library
	// If empty, the global Destination template is used
	OutputPath string `yaml:"output_path,omitempty"`
}

// GetLibraryDestination generates an absolute destination path for a library
// by applying the library name to the path template and resolving it to an absolute path.
// It uses the library's OutputPath if specified, otherwise falls back to the global Destination.
func (fc *FrontendConfig) GetLibraryDestination(libraryName string, libConfig LibraryConfig) (string, error) {
	// Determine which path template to use
	pathTemplate := fc.Destination
	if libConfig.OutputPath != "" {
		pathTemplate = libConfig.OutputPath
	}

	if pathTemplate == "" {
		return "", fmt.Errorf("no destination path configured for library %s", libraryName)
	}

	// Replace {library_name} placeholder with actual library name
	resolvedPath := strings.ReplaceAll(pathTemplate, "{library_name}", libraryName)

	// Convert to absolute path
	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %s: %w", resolvedPath, err)
	}

	return absPath, nil
}

// GetLibraryVersions returns a map of library names to their versions
func (fc *FrontendConfig) GetLibraryVersions() map[string]string {
	versions := make(map[string]string, len(fc.Libraries))

	for libraryName, libConfig := range fc.Libraries {
		versions[libraryName] = libConfig.Version
	}

	return versions
}

// GetLibraryDestinations returns a map of library names to their absolute destination paths
func (fc *FrontendConfig) GetLibraryDestinations() (map[string]string, error) {
	destinations := make(map[string]string, len(fc.Libraries))

	for libraryName, libConfig := range fc.Libraries {
		destPath, err := fc.GetLibraryDestination(libraryName, libConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get destination for library %s: %w", libraryName, err)
		}
		destinations[libraryName] = destPath
	}

	return destinations, nil
}

// IsValidCDN checks if a CDN value is one of the supported CDNs
func IsValidCDN(cdn CDN) bool {
	switch cdn {
	case CDNUnpkg, CDNCdnjs, CDNJsdelivr:
		return true
	default:
		return false
	}
}

// GetLibraryCDN returns the effective CDN for a library, considering both
// the library-specific CDN and the global CDN setting
func (fc *FrontendConfig) GetLibraryCDN(libConfig LibraryConfig) CDN {
	// Use library-specific CDN if specified
	if libConfig.CDN != "" {
		return libConfig.CDN
	}

	// Fall back to global CDN
	return fc.CDN
}

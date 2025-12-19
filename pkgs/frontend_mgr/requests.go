package frontend_mgr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/hashicorp/go-version"
	"nexus-sds.com/smfaman/pkgs/cache"
)

var (
	// CacheManager is the global cache manager instance
	CacheManager *cache.Manager

	// CacheEnabled controls whether caching is enabled globally
	CacheEnabled = true
)

func init() {
	// Initialize cache manager with default settings
	var err error
	CacheManager, err = cache.NewManager(CacheEnabled, 24*time.Hour)
	if err != nil {
		// If cache initialization fails, disable caching
		CacheEnabled = false
	}
}

// SetCacheEnabled enables or disables caching
func SetCacheEnabled(enabled bool) error {
	CacheEnabled = enabled
	var err error
	CacheManager, err = cache.NewManager(enabled, 24*time.Hour)
	return err
}

// FetchUnpkgMeta fetches package metadata from UNPKG CDN
// Endpoint: https://unpkg.com/{library_name}@{version}/?meta
func FetchUnpkgMeta(libraryName, version string) (*UnpkgMetaResponse, error) {
	// Check cache first
	cacheKey := cache.GenerateKey("unpkg", "meta", libraryName, version)
	var result UnpkgMetaResponse
	if found, _ := CacheManager.Get(cacheKey, &result); found {
		return &result, nil
	}

	url := fmt.Sprintf("https://unpkg.com/%s@%s/?meta", libraryName, version)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from UNPKG: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("UNPKG API returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode UNPKG response: %w", err)
	}

	// Store in cache
	CacheManager.Set(cacheKey, &result)

	return &result, nil
}

// FetchCdnjsVersion fetches version-specific package data from CDNJS
// Endpoint: https://api.cdnjs.com/libraries/{library_name}/{version}
func FetchCdnjsVersion(libraryName, version string) (*CdnjsVersionResponse, error) {
	// Check cache first
	cacheKey := cache.GenerateKey("cdnjs", "version", libraryName, version)
	var result CdnjsVersionResponse
	if found, _ := CacheManager.Get(cacheKey, &result); found {
		return &result, nil
	}

	url := fmt.Sprintf("https://api.cdnjs.com/libraries/%s/%s", libraryName, version)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from CDNJS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CDNJS API returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode CDNJS response: %w", err)
	}

	// Store in cache
	CacheManager.Set(cacheKey, &result)

	return &result, nil
}

// FetchJsdelivrPackage fetches package metadata from jsDelivr CDN
// Endpoint: https://data.jsdelivr.com/v1/packages/npm/{library_name}@{version}
func FetchJsdelivrPackage(libraryName, version string) (*JsdelivrPackageResponse, error) {
	// Check cache first
	cacheKey := cache.GenerateKey("jsdelivr", "package", libraryName, version)
	var result JsdelivrPackageResponse
	if found, _ := CacheManager.Get(cacheKey, &result); found {
		return &result, nil
	}

	url := fmt.Sprintf("https://data.jsdelivr.com/v1/packages/npm/%s@%s", libraryName, version)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from jsDelivr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jsDelivr API returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode jsDelivr response: %w", err)
	}

	// Store in cache
	CacheManager.Set(cacheKey, &result)

	return &result, nil
}

// FetchCdnjsVersions fetches all available versions for a package from CDNJS
// Endpoint: https://api.cdnjs.com/libraries/{library_name}
func FetchCdnjsVersions(libraryName string) (*CdnjsLibraryResponse, error) {
	// Check cache first
	cacheKey := cache.GenerateKey("cdnjs", "versions", libraryName)
	var result CdnjsLibraryResponse
	if found, _ := CacheManager.Get(cacheKey, &result); found {
		return &result, nil
	}

	url := fmt.Sprintf("https://api.cdnjs.com/libraries/%s", libraryName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from CDNJS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CDNJS API returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode CDNJS response: %w", err)
	}

	// Store in cache
	CacheManager.Set(cacheKey, &result)

	return &result, nil
}

// FetchJsdelivrVersions fetches all available versions for a package from jsDelivr
// Endpoint: https://data.jsdelivr.com/v1/packages/npm/{library_name}
func FetchJsdelivrVersions(libraryName string) (*JsdelivrVersionsResponse, error) {
	// Check cache first
	cacheKey := cache.GenerateKey("jsdelivr", "versions", libraryName)
	var result JsdelivrVersionsResponse
	if found, _ := CacheManager.Get(cacheKey, &result); found {
		return &result, nil
	}

	url := fmt.Sprintf("https://data.jsdelivr.com/v1/packages/npm/%s", libraryName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from jsDelivr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jsDelivr API returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode jsDelivr response: %w", err)
	}

	// Store in cache
	CacheManager.Set(cacheKey, &result)

	return &result, nil
}

// FetchUnpkgVersions fetches all available versions for a package from npm registry
// UNPKG doesn't have its own versions API, so we use the npm registry
// Endpoint: https://registry.npmjs.org/{library_name}
func FetchUnpkgVersions(libraryName string) (*UnpkgPackageResponse, error) {
	// Check cache first
	cacheKey := cache.GenerateKey("unpkg", "versions", libraryName)
	var result UnpkgPackageResponse
	if found, _ := CacheManager.Get(cacheKey, &result); found {
		return &result, nil
	}

	url := fmt.Sprintf("https://registry.npmjs.org/%s", libraryName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from npm registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("npm registry API returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode npm registry response: %w", err)
	}

	// Store in cache
	CacheManager.Set(cacheKey, &result)

	return &result, nil
}

// SortVersions sorts version strings in descending order (newest first)
// Uses semantic versioning for proper sorting
func SortVersions(versions []string) []string {
	sorted := make([]*version.Version, 0, len(versions))

	for _, v := range versions {
		ver, err := version.NewVersion(v)
		if err != nil {
			// If parsing fails, skip this version
			continue
		}
		sorted = append(sorted, ver)
	}

	// Sort in descending order
	sort.Sort(sort.Reverse(version.Collection(sorted)))

	// Convert back to strings
	result := make([]string, len(sorted))
	for i, v := range sorted {
		result[i] = v.Original()
	}

	return result
}

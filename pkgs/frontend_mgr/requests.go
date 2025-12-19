package frontend_mgr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FetchUnpkgMeta fetches package metadata from UNPKG CDN
// Endpoint: https://unpkg.com/{library_name}@{version}/?meta
func FetchUnpkgMeta(libraryName, version string) (*UnpkgMetaResponse, error) {
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

	var result UnpkgMetaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode UNPKG response: %w", err)
	}

	return &result, nil
}

// FetchCdnjsVersion fetches version-specific package data from CDNJS
// Endpoint: https://api.cdnjs.com/libraries/{library_name}/{version}
func FetchCdnjsVersion(libraryName, version string) (*CdnjsVersionResponse, error) {
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

	var result CdnjsVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode CDNJS response: %w", err)
	}

	return &result, nil
}

// FetchJsdelivrPackage fetches package metadata from jsDelivr CDN
// Endpoint: https://data.jsdelivr.com/v1/packages/npm/{library_name}@{version}
func FetchJsdelivrPackage(libraryName, version string) (*JsdelivrPackageResponse, error) {
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

	var result JsdelivrPackageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode jsDelivr response: %w", err)
	}

	return &result, nil
}

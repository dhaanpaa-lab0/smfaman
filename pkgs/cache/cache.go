package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultTTL is the default time-to-live for cache entries (24 hours)
	DefaultTTL = 24 * time.Hour

	// CacheDirName is the name of the cache directory
	CacheDirName = ".smfaman-cache"

	// MetadataDirName is the subdirectory for metadata cache
	MetadataDirName = "metadata"

	// PackagesDirName is the subdirectory for package file cache
	PackagesDirName = "packages"
)

// Entry represents a cached CDN response
type Entry struct {
	Key       string          `json:"key"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	TTL       time.Duration   `json:"ttl"`
}

// Manager handles cache operations
type Manager struct {
	cacheDir     string
	metadataDir  string
	packagesDir  string
	ttl          time.Duration
	enabled      bool
	packageCache bool
}

// NewManager creates a new cache manager
func NewManager(enabled bool, ttl time.Duration) (*Manager, error) {
	if ttl == 0 {
		ttl = DefaultTTL
	}

	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}

	m := &Manager{
		cacheDir:     cacheDir,
		metadataDir:  filepath.Join(cacheDir, MetadataDirName),
		packagesDir:  filepath.Join(cacheDir, PackagesDirName),
		ttl:          ttl,
		enabled:      enabled,
		packageCache: true, // Enable package caching by default
	}

	if enabled {
		// Ensure cache directories exist
		if err := os.MkdirAll(m.metadataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create metadata cache directory: %w", err)
		}
		if err := os.MkdirAll(m.packagesDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create packages cache directory: %w", err)
		}
	}

	return m, nil
}

// SetPackageCacheEnabled enables or disables package caching
func (m *Manager) SetPackageCacheEnabled(enabled bool) {
	m.packageCache = enabled
}

// Get retrieves a cached entry if it exists and is not expired
func (m *Manager) Get(key string, result interface{}) (bool, error) {
	if !m.enabled {
		return false, nil
	}

	filePath := m.getFilePath(key)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, nil
	}

	// Read cache file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse cache entry
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Check if expired
	if time.Since(entry.Timestamp) > entry.TTL {
		// Remove expired entry
		os.Remove(filePath)
		return false, nil
	}

	// Unmarshal data into result
	if err := json.Unmarshal(entry.Data, result); err != nil {
		return false, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return true, nil
}

// Set stores data in the cache
func (m *Manager) Set(key string, data interface{}) error {
	if !m.enabled {
		return nil
	}

	// Marshal data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create cache entry
	entry := Entry{
		Key:       key,
		Data:      dataBytes,
		Timestamp: time.Now(),
		TTL:       m.ttl,
	}

	// Marshal entry
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Write to file
	filePath := m.getFilePath(key)
	if err := os.WriteFile(filePath, entryBytes, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Clear removes all cached entries (both metadata and packages)
func (m *Manager) Clear() error {
	if !m.enabled {
		return nil
	}

	// Remove cache directory
	if err := os.RemoveAll(m.cacheDir); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	// Recreate cache directories
	if err := os.MkdirAll(m.metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate metadata cache directory: %w", err)
	}
	if err := os.MkdirAll(m.packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate packages cache directory: %w", err)
	}

	return nil
}

// ClearPackages removes all cached package files
func (m *Manager) ClearPackages() error {
	if !m.enabled {
		return nil
	}

	// Remove packages directory
	if err := os.RemoveAll(m.packagesDir); err != nil {
		return fmt.Errorf("failed to clear package cache: %w", err)
	}

	// Recreate packages directory
	if err := os.MkdirAll(m.packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate package cache directory: %w", err)
	}

	return nil
}

// ClearExpired removes all expired metadata cache entries
// Package cache files are not expired (they're kept indefinitely)
func (m *Manager) ClearExpired() (int, error) {
	if !m.enabled {
		return 0, nil
	}

	entries, err := os.ReadDir(m.metadataDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read metadata cache directory: %w", err)
	}

	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(m.metadataDir, entry.Name())

		// Read and parse entry
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var cacheEntry Entry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		// Check if expired
		if time.Since(cacheEntry.Timestamp) > cacheEntry.TTL {
			if err := os.Remove(filePath); err == nil {
				removed++
			}
		}
	}

	return removed, nil
}

// Stats returns cache statistics
func (m *Manager) Stats() (CacheStats, error) {
	stats := CacheStats{
		Enabled:       m.enabled,
		CacheDir:      m.cacheDir,
		TTL:           m.ttl,
		PackageCache:  m.packageCache,
	}

	if !m.enabled {
		return stats, nil
	}

	// Count metadata cache entries
	entries, err := os.ReadDir(m.metadataDir)
	if err != nil {
		return stats, fmt.Errorf("failed to read metadata cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(m.metadataDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		stats.MetadataEntries++
		stats.MetadataSize += info.Size()

		// Read and parse entry to check expiration
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var cacheEntry Entry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		if time.Since(cacheEntry.Timestamp) > cacheEntry.TTL {
			stats.ExpiredEntries++
		}
	}

	// Count package cache entries
	if m.packageCache {
		stats.PackageFiles, stats.PackageSize = countPackageFiles(m.packagesDir)
	}

	stats.TotalSize = stats.MetadataSize + stats.PackageSize

	return stats, nil
}

// countPackageFiles recursively counts files and total size in package cache
func countPackageFiles(dir string) (int, int64) {
	var count int
	var size int64

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		count++
		size += info.Size()
		return nil
	})

	return count, size
}

// CacheStats represents cache statistics
type CacheStats struct {
	Enabled         bool
	CacheDir        string
	TTL             time.Duration
	PackageCache    bool
	MetadataEntries int
	ExpiredEntries  int
	MetadataSize    int64
	PackageFiles    int
	PackageSize     int64
	TotalSize       int64
}

// getFilePath returns the file path for a metadata cache key
func (m *Manager) getFilePath(key string) string {
	// Use SHA256 hash of key as filename to avoid filesystem issues
	hash := sha256.Sum256([]byte(key))
	filename := fmt.Sprintf("%x.json", hash)
	return filepath.Join(m.metadataDir, filename)
}

// getPackageFilePath returns the file path for a package cache entry
// Key format: "{cdn}/{library}/{version}/{filepath}"
func (m *Manager) getPackageFilePath(cdn, library, version, filePath string) string {
	// Create directory structure: packages/{cdn}/{library}/{version}/{filepath}
	return filepath.Join(m.packagesDir, cdn, library, version, filePath)
}

// GetPackageFile retrieves a cached package file
// Returns the file data, whether it was found, and any error
func (m *Manager) GetPackageFile(cdn, library, version, filePath string) ([]byte, bool, error) {
	if !m.enabled || !m.packageCache {
		return nil, false, nil
	}

	cachePath := m.getPackageFilePath(cdn, library, version, filePath)

	// Check if file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, false, nil
	}

	// Read file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read cached package file: %w", err)
	}

	return data, true, nil
}

// SetPackageFile stores a package file in the cache
func (m *Manager) SetPackageFile(cdn, library, version, filePath string, data []byte) error {
	if !m.enabled || !m.packageCache {
		return nil
	}

	cachePath := m.getPackageFilePath(cdn, library, version, filePath)

	// Create directory structure
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package file to cache: %w", err)
	}

	return nil
}

// getCacheDir returns the cache directory path
func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, CacheDirName), nil
}

// GenerateKey generates a cache key from components
func GenerateKey(components ...string) string {
	return fmt.Sprintf("%v", components)
}

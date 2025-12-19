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
	cacheDir string
	ttl      time.Duration
	enabled  bool
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
		cacheDir: cacheDir,
		ttl:      ttl,
		enabled:  enabled,
	}

	if enabled {
		// Ensure cache directory exists
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}
	}

	return m, nil
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

// Clear removes all cached entries
func (m *Manager) Clear() error {
	if !m.enabled {
		return nil
	}

	// Remove cache directory
	if err := os.RemoveAll(m.cacheDir); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	// Recreate cache directory
	if err := os.MkdirAll(m.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate cache directory: %w", err)
	}

	return nil
}

// ClearExpired removes all expired cache entries
func (m *Manager) ClearExpired() (int, error) {
	if !m.enabled {
		return 0, nil
	}

	entries, err := os.ReadDir(m.cacheDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(m.cacheDir, entry.Name())

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
		Enabled:  m.enabled,
		CacheDir: m.cacheDir,
		TTL:      m.ttl,
	}

	if !m.enabled {
		return stats, nil
	}

	entries, err := os.ReadDir(m.cacheDir)
	if err != nil {
		return stats, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(m.cacheDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		stats.TotalEntries++
		stats.TotalSize += info.Size()

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

	return stats, nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	Enabled        bool
	CacheDir       string
	TTL            time.Duration
	TotalEntries   int
	ExpiredEntries int
	TotalSize      int64
}

// getFilePath returns the file path for a cache key
func (m *Manager) getFilePath(key string) string {
	// Use SHA256 hash of key as filename to avoid filesystem issues
	hash := sha256.Sum256([]byte(key))
	filename := fmt.Sprintf("%x.json", hash)
	return filepath.Join(m.cacheDir, filename)
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

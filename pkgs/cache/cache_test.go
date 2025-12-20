package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager(true, DefaultTTL)
	if err != nil {
		t.Fatalf("failed to create cache manager: %v", err)
	}

	if !manager.enabled {
		t.Error("expected cache to be enabled")
	}

	if manager.ttl != DefaultTTL {
		t.Errorf("expected TTL %v, got %v", DefaultTTL, manager.ttl)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	manager, err := NewManager(true, DefaultTTL)
	if err != nil {
		t.Fatalf("failed to create cache manager: %v", err)
	}
	defer manager.Clear()

	type TestData struct {
		Name    string
		Version string
		Count   int
	}

	// Set data
	testData := TestData{
		Name:    "test-package",
		Version: "1.0.0",
		Count:   42,
	}

	key := "test-key"
	if err := manager.Set(key, testData); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Get data
	var retrieved TestData
	found, err := manager.Get(key, &retrieved)
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	if !found {
		t.Error("expected to find cached data")
	}

	if retrieved != testData {
		t.Errorf("expected %+v, got %+v", testData, retrieved)
	}
}

func TestCacheExpiration(t *testing.T) {
	// Create manager with 1 second TTL
	manager, err := NewManager(true, 1*time.Second)
	if err != nil {
		t.Fatalf("failed to create cache manager: %v", err)
	}
	defer manager.Clear()

	key := "expiring-key"
	testData := map[string]string{"test": "data"}

	// Set data
	if err := manager.Set(key, testData); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Should find it immediately
	var retrieved map[string]string
	found, err := manager.Get(key, &retrieved)
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	if !found {
		t.Error("expected to find cached data")
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Should not find it after expiration
	found, err = manager.Get(key, &retrieved)
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	if found {
		t.Error("expected cache entry to be expired and not found")
	}
}

func TestCacheDisabled(t *testing.T) {
	manager, err := NewManager(false, DefaultTTL)
	if err != nil {
		t.Fatalf("failed to create cache manager: %v", err)
	}

	key := "disabled-key"
	testData := "test data"

	// Set should not error but also not store
	if err := manager.Set(key, testData); err != nil {
		t.Fatalf("set should not error when disabled: %v", err)
	}

	// Get should return false without error
	var retrieved string
	found, err := manager.Get(key, &retrieved)
	if err != nil {
		t.Fatalf("get should not error when disabled: %v", err)
	}

	if found {
		t.Error("expected not to find data when cache is disabled")
	}
}

func TestCacheClear(t *testing.T) {
	manager, err := NewManager(true, DefaultTTL)
	if err != nil {
		t.Fatalf("failed to create cache manager: %v", err)
	}

	// Add some entries
	for i := 0; i < 5; i++ {
		key := GenerateKey("test", string(rune(i)))
		if err := manager.Set(key, i); err != nil {
			t.Fatalf("failed to set cache: %v", err)
		}
	}

	// Verify entries exist
	stats, err := manager.Stats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.MetadataEntries != 5 {
		t.Errorf("expected 5 entries, got %d", stats.MetadataEntries)
	}

	// Clear cache
	if err := manager.Clear(); err != nil {
		t.Fatalf("failed to clear cache: %v", err)
	}

	// Verify entries are gone
	stats, err = manager.Stats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.MetadataEntries != 0 {
		t.Errorf("expected 0 entries after clear, got %d", stats.MetadataEntries)
	}
}

func TestClearExpired(t *testing.T) {
	manager, err := NewManager(true, 1*time.Second)
	if err != nil {
		t.Fatalf("failed to create cache manager: %v", err)
	}
	defer manager.Clear()

	// Add some entries
	for i := 0; i < 3; i++ {
		key := GenerateKey("test", string(rune(i)))
		if err := manager.Set(key, i); err != nil {
			t.Fatalf("failed to set cache: %v", err)
		}
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Add one fresh entry
	if err := manager.Set("fresh", "data"); err != nil {
		t.Fatalf("failed to set fresh cache: %v", err)
	}

	// Clear expired
	removed, err := manager.ClearExpired()
	if err != nil {
		t.Fatalf("failed to clear expired: %v", err)
	}

	if removed != 3 {
		t.Errorf("expected to remove 3 expired entries, removed %d", removed)
	}

	// Verify fresh entry still exists
	stats, err := manager.Stats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.MetadataEntries != 1 {
		t.Errorf("expected 1 entry remaining, got %d", stats.MetadataEntries)
	}
}

func TestGenerateKey(t *testing.T) {
	key1 := GenerateKey("cdn", "package", "version")
	key2 := GenerateKey("cdn", "package", "version")
	key3 := GenerateKey("different", "package", "version")

	if key1 != key2 {
		t.Error("same components should generate same key")
	}

	if key1 == key3 {
		t.Error("different components should generate different keys")
	}
}

func TestGetFilePath(t *testing.T) {
	manager, err := NewManager(true, DefaultTTL)
	if err != nil {
		t.Fatalf("failed to create cache manager: %v", err)
	}
	defer manager.Clear()

	key := "test-key"
	path := manager.getFilePath(key)

	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}

	if filepath.Ext(path) != ".json" {
		t.Error("expected .json extension")
	}
}

func TestGetCacheDir(t *testing.T) {
	dir, err := getCacheDir()
	if err != nil {
		t.Fatalf("failed to get cache dir: %v", err)
	}

	if !filepath.IsAbs(dir) {
		t.Error("expected absolute path")
	}

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, CacheDirName)

	if dir != expectedDir {
		t.Errorf("expected cache dir %q, got %q", expectedDir, dir)
	}
}

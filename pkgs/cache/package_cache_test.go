package cache

import (
	"testing"
)

func TestPackageCache(t *testing.T) {
	// Create cache manager (uses default cache directory)
	manager, err := NewManager(true, DefaultTTL)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	// Clean up at the end
	defer manager.ClearPackages()

	// Test data
	cdn := "unpkg"
	library := "react"
	version := "18.2.0"
	filePath := "umd/react.production.min.js"
	fileData := []byte("// React production build\nconsole.log('React');\n")

	// Test SetPackageFile
	t.Run("SetPackageFile", func(t *testing.T) {
		err := manager.SetPackageFile(cdn, library, version, filePath, fileData)
		if err != nil {
			t.Errorf("SetPackageFile failed: %v", err)
		}

		// Verify file was created by checking if it can be retrieved
		_, found, _ := manager.GetPackageFile(cdn, library, version, filePath)
		if !found {
			t.Error("Package file was not stored in cache")
		}
	})

	// Test GetPackageFile
	t.Run("GetPackageFile", func(t *testing.T) {
		data, found, err := manager.GetPackageFile(cdn, library, version, filePath)
		if err != nil {
			t.Errorf("GetPackageFile failed: %v", err)
		}
		if !found {
			t.Error("Package file not found in cache")
		}
		if string(data) != string(fileData) {
			t.Errorf("Retrieved data doesn't match. Got %q, want %q", string(data), string(fileData))
		}
	})

	// Test GetPackageFile for non-existent file
	t.Run("GetPackageFile_NotFound", func(t *testing.T) {
		data, found, err := manager.GetPackageFile(cdn, library, version, "nonexistent.js")
		if err != nil {
			t.Errorf("GetPackageFile failed: %v", err)
		}
		if found {
			t.Error("Expected file not to be found")
		}
		if data != nil {
			t.Error("Expected nil data for non-existent file")
		}
	})

	// Test Stats includes package cache
	t.Run("Stats", func(t *testing.T) {
		stats, err := manager.Stats()
		if err != nil {
			t.Fatalf("Stats failed: %v", err)
		}

		if !stats.PackageCache {
			t.Error("Package cache should be enabled")
		}

		if stats.PackageFiles < 1 {
			t.Errorf("Expected at least 1 package file, got %d", stats.PackageFiles)
		}

		if stats.PackageSize == 0 {
			t.Error("Expected non-zero package size")
		}
	})

	// Test ClearPackages
	t.Run("ClearPackages", func(t *testing.T) {
		err := manager.ClearPackages()
		if err != nil {
			t.Errorf("ClearPackages failed: %v", err)
		}

		// Verify file was removed
		data, found, _ := manager.GetPackageFile(cdn, library, version, filePath)
		if found {
			t.Error("Package file should have been cleared")
		}
		if data != nil {
			t.Error("Expected nil data after clear")
		}

		// Verify stats show empty cache
		stats, _ := manager.Stats()
		if stats.PackageFiles != 0 {
			t.Errorf("Expected 0 package files after clear, got %d", stats.PackageFiles)
		}
	})
}

func TestPackageCacheDisabled(t *testing.T) {
	// Create cache manager
	manager, err := NewManager(true, DefaultTTL)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	// Disable package cache
	manager.SetPackageCacheEnabled(false)

	// Test data
	cdn := "unpkg"
	library := "react"
	version := "18.2.0"
	filePath := "umd/react.production.min.js"
	fileData := []byte("// React production build\n")

	// SetPackageFile should not create file
	err = manager.SetPackageFile(cdn, library, version, filePath, fileData)
	if err != nil {
		t.Errorf("SetPackageFile failed: %v", err)
	}

	// GetPackageFile should return not found
	data, found, err := manager.GetPackageFile(cdn, library, version, filePath)
	if err != nil {
		t.Errorf("GetPackageFile failed: %v", err)
	}
	if found {
		t.Error("Expected file not to be cached when package cache is disabled")
	}
	if data != nil {
		t.Error("Expected nil data when package cache is disabled")
	}
}

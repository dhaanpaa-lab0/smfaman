package frontend_mgr

import (
	"testing"
)

func TestBootstrapOnAllCDNs(t *testing.T) {
	version := "5.3.0"

	t.Run("UNPKG Bootstrap", func(t *testing.T) {
		result, err := FetchUnpkgMeta("bootstrap", version)
		if err != nil {
			t.Fatalf("Failed to fetch bootstrap from UNPKG: %v", err)
		}
		t.Logf("✓ UNPKG - Package: %s, Version: %s, Files: %d", result.Package, result.Version, len(result.Files))

		if result.Package != "bootstrap" {
			t.Errorf("Expected package 'bootstrap', got '%s'", result.Package)
		}
		if len(result.Files) == 0 {
			t.Error("Expected files array to have items")
		}
	})

	t.Run("CDNJS Bootstrap", func(t *testing.T) {
		result, err := FetchCdnjsVersion("bootstrap", version)
		if err != nil {
			t.Fatalf("Failed to fetch bootstrap from CDNJS: %v", err)
		}
		t.Logf("✓ CDNJS - Name: %s, Version: %s, Files: %d", result.Name, result.Version, len(result.Files))

		if result.Name != "bootstrap" {
			t.Errorf("Expected name 'bootstrap', got '%s'", result.Name)
		}
		if len(result.Files) == 0 {
			t.Error("Expected files array to have items")
		}
	})

	t.Run("jsDelivr Bootstrap", func(t *testing.T) {
		result, err := FetchJsdelivrPackage("bootstrap", version)
		if err != nil {
			t.Fatalf("Failed to fetch bootstrap from jsDelivr: %v", err)
		}
		t.Logf("✓ jsDelivr - Name: %s, Version: %s, Files: %d", result.Name, result.Version, len(result.Files))

		if result.Name != "bootstrap" {
			t.Errorf("Expected name 'bootstrap', got '%s'", result.Name)
		}
		if len(result.Files) == 0 {
			t.Error("Expected files array to have items")
		}
	})
}

func TestBootswatchOnAllCDNs(t *testing.T) {
	version := "5.3.0"

	t.Run("UNPKG Bootswatch", func(t *testing.T) {
		result, err := FetchUnpkgMeta("bootswatch", version)
		if err != nil {
			t.Fatalf("Failed to fetch bootswatch from UNPKG: %v", err)
		}
		t.Logf("✓ UNPKG - Package: %s, Version: %s, Files: %d", result.Package, result.Version, len(result.Files))

		if result.Package != "bootswatch" {
			t.Errorf("Expected package 'bootswatch', got '%s'", result.Package)
		}
		if len(result.Files) == 0 {
			t.Error("Expected files array to have items")
		}
	})

	t.Run("CDNJS Bootswatch", func(t *testing.T) {
		result, err := FetchCdnjsVersion("bootswatch", version)
		if err != nil {
			t.Fatalf("Failed to fetch bootswatch from CDNJS: %v", err)
		}
		t.Logf("✓ CDNJS - Name: %s, Version: %s, Files: %d", result.Name, result.Version, len(result.Files))

		if result.Name != "bootswatch" {
			t.Errorf("Expected name 'bootswatch', got '%s'", result.Name)
		}
		if len(result.Files) == 0 {
			t.Error("Expected files array to have items")
		}
	})

	t.Run("jsDelivr Bootswatch", func(t *testing.T) {
		result, err := FetchJsdelivrPackage("bootswatch", version)
		if err != nil {
			t.Fatalf("Failed to fetch bootswatch from jsDelivr: %v", err)
		}
		t.Logf("✓ jsDelivr - Name: %s, Version: %s, Files: %d", result.Name, result.Version, len(result.Files))

		if result.Name != "bootswatch" {
			t.Errorf("Expected name 'bootswatch', got '%s'", result.Name)
		}
		if len(result.Files) == 0 {
			t.Error("Expected files array to have items")
		}
	})
}

// TestShowBootstrapFiles demonstrates accessing file information from each CDN
func TestShowBootstrapFiles(t *testing.T) {
	version := "5.3.0"

	t.Run("Show Bootstrap CSS files", func(t *testing.T) {
		// Test UNPKG
		unpkgResult, err := FetchUnpkgMeta("bootstrap", version)
		if err != nil {
			t.Fatalf("UNPKG failed: %v", err)
		}

		cssFiles := 0
		for _, file := range unpkgResult.Files {
			if len(file.Path) > 4 && file.Path[len(file.Path)-4:] == ".css" {
				cssFiles++
				if cssFiles <= 3 {
					t.Logf("  UNPKG CSS: %s (%d bytes)", file.Path, file.Size)
				}
			}
		}
		t.Logf("UNPKG: Found %d CSS files", cssFiles)

		// Test CDNJS
		cdnjsResult, err := FetchCdnjsVersion("bootstrap", version)
		if err != nil {
			t.Fatalf("CDNJS failed: %v", err)
		}

		cssFilesCount := 0
		for _, file := range cdnjsResult.Files {
			if len(file) > 4 && file[len(file)-4:] == ".css" {
				cssFilesCount++
				if cssFilesCount <= 3 {
					sri := cdnjsResult.SRI[file]
					t.Logf("  CDNJS CSS: %s (SRI: %s...)", file, sri[:20])
				}
			}
		}
		t.Logf("CDNJS: Found %d CSS files", cssFilesCount)

		// Test jsDelivr
		jsdelivrResult, err := FetchJsdelivrPackage("bootstrap", version)
		if err != nil {
			t.Fatalf("jsDelivr failed: %v", err)
		}

		t.Logf("jsDelivr: Default file: %s", jsdelivrResult.Default)
		t.Logf("jsDelivr: Total directories/files at root: %d", len(jsdelivrResult.Files))
	})
}

// Benchmark tests to compare CDN response times
func BenchmarkCDNs(b *testing.B) {
	version := "5.3.0"

	b.Run("UNPKG", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := FetchUnpkgMeta("bootstrap", version)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CDNJS", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := FetchCdnjsVersion("bootstrap", version)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("jsDelivr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := FetchJsdelivrPackage("bootstrap", version)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

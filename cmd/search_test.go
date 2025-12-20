package cmd

import (
	"testing"

	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

func TestPerformSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	tests := []struct {
		name      string
		query     string
		cdn       string
		limit     int
		wantError bool
		minResults int
	}{
		{
			name:       "Search all CDNs for react",
			query:      "react",
			cdn:        "all",
			limit:      10,
			wantError:  false,
			minResults: 1,
		},
		{
			name:       "Search CDNJS only",
			query:      "jquery",
			cdn:        "cdnjs",
			limit:      5,
			wantError:  false,
			minResults: 1,
		},
		{
			name:       "Search npm registry",
			query:      "vue",
			cdn:        "npm",
			limit:      5,
			wantError:  false,
			minResults: 1,
		},
		{
			name:       "Invalid CDN",
			query:      "react",
			cdn:        "invalid",
			limit:      10,
			wantError:  true,
			minResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := performSearch(tt.query, tt.cdn, tt.limit)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Skipf("Network error (expected in some environments): %v", err)
				return
			}

			if len(results) < tt.minResults {
				t.Errorf("Expected at least %d results, got %d", tt.minResults, len(results))
			}

			// Verify result structure
			for i, r := range results {
				if r.Name == "" {
					t.Errorf("Result %d has empty Name", i)
				}
				if r.Version == "" {
					t.Errorf("Result %d has empty Version", i)
				}
				if r.CDN == "" {
					t.Errorf("Result %d has empty CDN", i)
				}
			}

			t.Logf("Found %d results for query '%s' on %s", len(results), tt.query, tt.cdn)
		})
	}
}

func TestSearchCDNs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	t.Run("SearchCdnjs", func(t *testing.T) {
		results, err := frontend_mgr.SearchCdnjs("bootstrap", 5)
		if err != nil {
			t.Skipf("Network error: %v", err)
			return
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'bootstrap'")
		}

		// Check first result
		if len(results) > 0 {
			r := results[0]
			if r.Name == "" {
				t.Error("Result has empty name")
			}
			if r.Version == "" {
				t.Error("Result has empty version")
			}
			if r.CDN != "cdnjs" {
				t.Errorf("Expected CDN to be 'cdnjs', got '%s'", r.CDN)
			}
			t.Logf("First result: %s@%s - %s", r.Name, r.Version, r.Description)
		}
	})

	t.Run("SearchNpm", func(t *testing.T) {
		results, err := frontend_mgr.SearchNpm("lodash", 5)
		if err != nil {
			t.Skipf("Network error: %v", err)
			return
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'lodash'")
		}

		// Check first result
		if len(results) > 0 {
			r := results[0]
			if r.Name == "" {
				t.Error("Result has empty name")
			}
			if r.Version == "" {
				t.Error("Result has empty version")
			}
			if r.CDN != "npm" {
				t.Errorf("Expected CDN to be 'npm', got '%s'", r.CDN)
			}
			t.Logf("First result: %s@%s - %s", r.Name, r.Version, r.Description)
		}
	})

	t.Run("SearchAllCDNs", func(t *testing.T) {
		results, err := frontend_mgr.SearchAllCDNs("jquery", 10)
		if err != nil {
			t.Skipf("Network error: %v", err)
			return
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'jquery'")
		}

		// Should have deduplicated results
		seen := make(map[string]bool)
		for _, r := range results {
			if seen[r.Name] {
				t.Errorf("Duplicate package name found: %s", r.Name)
			}
			seen[r.Name] = true
		}

		t.Logf("Found %d unique packages across all CDNs", len(results))
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "Short string",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "Exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "Long string",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "Very short maxLen",
			input:  "hello",
			maxLen: 3,
			want:   "hel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestSearchResultItem(t *testing.T) {
	result := frontend_mgr.SearchResult{
		Name:        "test-package",
		Version:     "1.0.0",
		Description: "A test package",
		Keywords:    []string{"test", "example"},
		CDN:         "npm",
	}

	item := searchResultItem{
		result: result,
		index:  0,
	}

	// Test FilterValue includes name, description, and keywords
	filterValue := item.FilterValue()
	if filterValue == "" {
		t.Error("FilterValue should not be empty")
	}

	// Should contain name
	if !contains(filterValue, "test-package") {
		t.Error("FilterValue should contain package name")
	}

	// Should contain description
	if !contains(filterValue, "A test package") {
		t.Error("FilterValue should contain description")
	}

	// Should contain keywords
	if !contains(filterValue, "test") && !contains(filterValue, "example") {
		t.Error("FilterValue should contain keywords")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package frontend_config

import (
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetLibraryDestination(t *testing.T) {
	tests := []struct {
		name          string
		config        FrontendConfig
		libraryName   string
		libConfig     LibraryConfig
		expectedPath  string
		shouldContain string
		shouldError   bool
	}{
		{
			name: "uses global destination template",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
			},
			libraryName: "jquery",
			libConfig:   LibraryConfig{Version: "3.5.1"},
			shouldContain: filepath.Join("frontend", "jquery"),
			shouldError:   false,
		},
		{
			name: "uses library-specific output path",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
			},
			libraryName: "bootstrap",
			libConfig: LibraryConfig{
				Version:    "4.5.2",
				OutputPath: "./custom/{library_name}",
			},
			shouldContain: filepath.Join("custom", "bootstrap"),
			shouldError:   false,
		},
		{
			name: "handles library name without placeholder",
			config: FrontendConfig{
				Destination: "./static/libs",
			},
			libraryName:   "react",
			libConfig:     LibraryConfig{Version: "18.2.0"},
			shouldContain: filepath.Join("static", "libs"),
			shouldError:   false,
		},
		{
			name: "errors when no destination configured",
			config: FrontendConfig{
				Destination: "",
			},
			libraryName: "vue",
			libConfig:   LibraryConfig{Version: "3.0.0"},
			shouldError: true,
		},
		{
			name: "returns absolute path",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
			},
			libraryName: "lodash",
			libConfig:   LibraryConfig{Version: "4.17.21"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.GetLibraryDestination(tt.libraryName, tt.libConfig)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check if result is absolute
			if !filepath.IsAbs(result) {
				t.Errorf("expected absolute path, got: %s", result)
			}

			// Check if result contains expected substring
			if tt.shouldContain != "" {
				if !contains(result, tt.shouldContain) {
					t.Errorf("expected path to contain %q, got: %s", tt.shouldContain, result)
				}
			}
		})
	}
}

func contains(path, substring string) bool {
	// Normalize paths for comparison
	normalizedPath := filepath.ToSlash(path)
	normalizedSubstring := filepath.ToSlash(substring)

	return len(normalizedPath) >= len(normalizedSubstring) &&
		normalizedPath[len(normalizedPath)-len(normalizedSubstring):] == normalizedSubstring ||
		filepath.Base(path) == filepath.Base(substring)
}

func TestGetLibraryVersions(t *testing.T) {
	tests := []struct {
		name     string
		config   FrontendConfig
		expected map[string]string
	}{
		{
			name: "returns library versions map",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
				Libraries: map[string]LibraryConfig{
					"jquery":    {Version: "3.5.1"},
					"bootstrap": {Version: "4.5.2"},
					"react":     {Version: "18.2.0"},
				},
			},
			expected: map[string]string{
				"jquery":    "3.5.1",
				"bootstrap": "4.5.2",
				"react":     "18.2.0",
			},
		},
		{
			name: "returns empty map for no libraries",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
				Libraries:   map[string]LibraryConfig{},
			},
			expected: map[string]string{},
		},
		{
			name: "handles nil libraries map",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
				Libraries:   nil,
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetLibraryVersions()

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d libraries, got %d", len(tt.expected), len(result))
				return
			}

			for name, expectedVersion := range tt.expected {
				if actualVersion, ok := result[name]; !ok {
					t.Errorf("expected library %q not found in result", name)
				} else if actualVersion != expectedVersion {
					t.Errorf("for library %q: expected version %q, got %q", name, expectedVersion, actualVersion)
				}
			}
		})
	}
}

func TestIsValidCDN(t *testing.T) {
	tests := []struct {
		name     string
		cdn      CDN
		expected bool
	}{
		{"valid unpkg", CDNUnpkg, true},
		{"valid cdnjs", CDNCdnjs, true},
		{"valid jsdelivr", CDNJsdelivr, true},
		{"invalid empty", CDN(""), false},
		{"invalid random", CDN("random"), false},
		{"invalid case", CDN("UNPKG"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidCDN(tt.cdn)
			if result != tt.expected {
				t.Errorf("IsValidCDN(%q) = %v, expected %v", tt.cdn, result, tt.expected)
			}
		})
	}
}

func TestGetLibraryCDN(t *testing.T) {
	tests := []struct {
		name      string
		config    FrontendConfig
		libConfig LibraryConfig
		expected  CDN
	}{
		{
			name: "uses library-specific CDN",
			config: FrontendConfig{
				CDN: CDNUnpkg,
			},
			libConfig: LibraryConfig{
				CDN: CDNCdnjs,
			},
			expected: CDNCdnjs,
		},
		{
			name: "falls back to global CDN",
			config: FrontendConfig{
				CDN: CDNJsdelivr,
			},
			libConfig: LibraryConfig{
				CDN: "",
			},
			expected: CDNJsdelivr,
		},
		{
			name: "returns empty when both are empty",
			config: FrontendConfig{
				CDN: "",
			},
			libConfig: LibraryConfig{
				CDN: "",
			},
			expected: "",
		},
		{
			name: "library CDN overrides global",
			config: FrontendConfig{
				CDN: CDNUnpkg,
			},
			libConfig: LibraryConfig{
				CDN: CDNJsdelivr,
			},
			expected: CDNJsdelivr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetLibraryCDN(tt.libConfig)
			if result != tt.expected {
				t.Errorf("GetLibraryCDN() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestYAMLSerialization(t *testing.T) {
	yamlData := `
destination: "./frontend/{library_name}"
project_name: "test-project"
cdn: "unpkg"
libraries:
  jquery:
    version: "3.5.1"
  bootstrap:
    version: "4.5.2"
    cdn: "cdnjs"
`

	var config FrontendConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	// Verify top-level fields
	if config.Destination != "./frontend/{library_name}" {
		t.Errorf("expected destination './frontend/{library_name}', got %q", config.Destination)
	}
	if config.ProjectName != "test-project" {
		t.Errorf("expected project_name 'test-project', got %q", config.ProjectName)
	}
	if config.CDN != CDNUnpkg {
		t.Errorf("expected CDN 'unpkg', got %q", config.CDN)
	}

	// Verify libraries
	if len(config.Libraries) != 2 {
		t.Fatalf("expected 2 libraries, got %d", len(config.Libraries))
	}

	// Check jquery
	jquery, ok := config.Libraries["jquery"]
	if !ok {
		t.Fatal("jquery not found in libraries")
	}
	if jquery.Version != "3.5.1" {
		t.Errorf("expected jquery version '3.5.1', got %q", jquery.Version)
	}
	if jquery.CDN != "" {
		t.Errorf("expected jquery CDN to be empty, got %q", jquery.CDN)
	}

	// Check bootstrap
	bootstrap, ok := config.Libraries["bootstrap"]
	if !ok {
		t.Fatal("bootstrap not found in libraries")
	}
	if bootstrap.Version != "4.5.2" {
		t.Errorf("expected bootstrap version '4.5.2', got %q", bootstrap.Version)
	}
	if bootstrap.CDN != CDNCdnjs {
		t.Errorf("expected bootstrap CDN 'cdnjs', got %q", bootstrap.CDN)
	}

	// Test GetLibraryCDN to verify effective CDN resolution
	jqueryCDN := config.GetLibraryCDN(jquery)
	if jqueryCDN != CDNUnpkg {
		t.Errorf("expected effective CDN for jquery to be 'unpkg', got %q", jqueryCDN)
	}

	bootstrapCDN := config.GetLibraryCDN(bootstrap)
	if bootstrapCDN != CDNCdnjs {
		t.Errorf("expected effective CDN for bootstrap to be 'cdnjs', got %q", bootstrapCDN)
	}

	// Test marshalling back to YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		t.Fatalf("failed to marshal config to YAML: %v", err)
	}

	// Unmarshal again to verify round-trip
	var config2 FrontendConfig
	err = yaml.Unmarshal(data, &config2)
	if err != nil {
		t.Fatalf("failed to unmarshal marshalled YAML: %v", err)
	}

	if config2.CDN != config.CDN {
		t.Errorf("round-trip failed: CDN changed from %q to %q", config.CDN, config2.CDN)
	}
}

func TestGetLibraryDestinations(t *testing.T) {
	tests := []struct {
		name              string
		config            FrontendConfig
		expectedLibraries []string
		shouldError       bool
	}{
		{
			name: "returns destination paths for all libraries",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
				Libraries: map[string]LibraryConfig{
					"jquery":    {Version: "3.5.1"},
					"bootstrap": {Version: "4.5.2"},
					"react":     {Version: "18.2.0"},
				},
			},
			expectedLibraries: []string{"jquery", "bootstrap", "react"},
			shouldError:       false,
		},
		{
			name: "handles library-specific output paths",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
				Libraries: map[string]LibraryConfig{
					"jquery": {Version: "3.5.1"},
					"custom": {
						Version:    "1.0.0",
						OutputPath: "./custom-libs/{library_name}",
					},
				},
			},
			expectedLibraries: []string{"jquery", "custom"},
			shouldError:       false,
		},
		{
			name: "returns empty map for no libraries",
			config: FrontendConfig{
				Destination: "./frontend/{library_name}",
				Libraries:   map[string]LibraryConfig{},
			},
			expectedLibraries: []string{},
			shouldError:       false,
		},
		{
			name: "errors when no destination configured",
			config: FrontendConfig{
				Destination: "",
				Libraries: map[string]LibraryConfig{
					"jquery": {Version: "3.5.1"},
				},
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.GetLibraryDestinations()

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expectedLibraries) {
				t.Errorf("expected %d libraries, got %d", len(tt.expectedLibraries), len(result))
				return
			}

			for _, libraryName := range tt.expectedLibraries {
				destPath, ok := result[libraryName]
				if !ok {
					t.Errorf("expected library %q not found in result", libraryName)
					continue
				}

				// Verify path is absolute
				if !filepath.IsAbs(destPath) {
					t.Errorf("expected absolute path for library %q, got: %s", libraryName, destPath)
				}

				// Verify path contains library name (unless it's a custom path without placeholder)
				if tt.config.Libraries[libraryName].OutputPath == "" {
					if !contains(destPath, libraryName) {
						t.Errorf("expected path for %q to contain library name, got: %s", libraryName, destPath)
					}
				}
			}
		})
	}
}

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		name            string
		spec            string
		expectedName    string
		expectedVersion string
	}{
		{
			name:            "package with version",
			spec:            "react@18.2.0",
			expectedName:    "react",
			expectedVersion: "18.2.0",
		},
		{
			name:            "package without version",
			spec:            "react",
			expectedName:    "react",
			expectedVersion: "",
		},
		{
			name:            "scoped package with version",
			spec:            "@babel/core@7.22.0",
			expectedName:    "@babel/core",
			expectedVersion: "7.22.0",
		},
		{
			name:            "scoped package without version",
			spec:            "@babel/core",
			expectedName:    "@babel/core",
			expectedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, version := parsePackageSpec(tt.spec)
			if name != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, name)
			}
			if version != tt.expectedVersion {
				t.Errorf("expected version %q, got %q", tt.expectedVersion, version)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// Test with non-existent file
	_, err := loadConfig(configPath)
	if err == nil {
		t.Error("expected error when loading non-existent config")
	}

	// Create valid config
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend/{library_name}",
		ProjectName: "test-project",
		CDN:         frontend_config.CDNUnpkg,
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.7.1"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	// Test loading valid config
	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if config.Destination != testConfig.Destination {
		t.Errorf("expected destination %q, got %q", testConfig.Destination, config.Destination)
	}

	if config.ProjectName != testConfig.ProjectName {
		t.Errorf("expected project name %q, got %q", testConfig.ProjectName, config.ProjectName)
	}

	if len(config.Libraries) != 1 {
		t.Errorf("expected 1 library, got %d", len(config.Libraries))
	}
}

func TestLoadConfigWithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)

	_, err := loadConfig(configPath)
	if err == nil {
		t.Error("expected error when loading invalid YAML")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "save-test.yaml")

	testConfig := frontend_config.FrontendConfig{
		Destination: "./public/libs",
		ProjectName: "save-test",
		CDN:         frontend_config.CDNCdnjs,
		Libraries: map[string]frontend_config.LibraryConfig{
			"bootstrap": {
				Version: "5.3.0",
				Files:   []string{"css/bootstrap.min.css"},
			},
		},
	}

	// Save config
	err := saveConfig(configPath, &testConfig)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load and verify content
	data, _ := os.ReadFile(configPath)
	var loadedConfig frontend_config.FrontendConfig
	yaml.Unmarshal(data, &loadedConfig)

	if loadedConfig.ProjectName != testConfig.ProjectName {
		t.Errorf("expected project name %q, got %q", testConfig.ProjectName, loadedConfig.ProjectName)
	}

	if len(loadedConfig.Libraries["bootstrap"].Files) != 1 {
		t.Error("expected bootstrap files to be preserved")
	}
}

func TestDetermineCDNForAdd(t *testing.T) {
	tests := []struct {
		name        string
		flagCDN     string
		configCDN   frontend_config.CDN
		expectedCDN frontend_config.CDN
	}{
		{
			name:        "flag takes precedence",
			flagCDN:     "cdnjs",
			configCDN:   frontend_config.CDNUnpkg,
			expectedCDN: frontend_config.CDNCdnjs,
		},
		{
			name:        "config default when no flag",
			flagCDN:     "",
			configCDN:   frontend_config.CDNJsdelivr,
			expectedCDN: frontend_config.CDNJsdelivr,
		},
		{
			name:        "unpkg fallback",
			flagCDN:     "",
			configCDN:   "",
			expectedCDN: frontend_config.CDNUnpkg,
		},
		{
			name:        "invalid flag uses config",
			flagCDN:     "invalid",
			configCDN:   frontend_config.CDNCdnjs,
			expectedCDN: frontend_config.CDNCdnjs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flag variable
			addCDN = tt.flagCDN
			defer func() { addCDN = "" }() // Reset after test

			config := &frontend_config.FrontendConfig{
				CDN:       tt.configCDN,
				Libraries: make(map[string]frontend_config.LibraryConfig),
			}

			result := determineCDNForAdd(config)
			if result != tt.expectedCDN {
				t.Errorf("expected CDN %q, got %q", tt.expectedCDN, result)
			}
		})
	}
}

func TestAddLibraryToConfigWithExistingLibrary(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "existing.yaml")

	// Create config with existing library
	initialConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.5.1"},
		},
	}

	data, _ := yaml.Marshal(&initialConfig)
	os.WriteFile(configPath, data, 0644)

	// Set global config path
	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Try to add without force (should fail)
	addForce = false
	err := addLibraryToConfig("jquery@3.7.1")
	if err == nil {
		t.Error("expected error when adding existing library without --force")
	}

	// Try with force (should succeed but we'll skip actual API calls)
	// This test focuses on the force flag logic
	addForce = false // Reset
}

func TestAddLibraryToConfigNonExistentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	err := addLibraryToConfig("react@18.2.0")
	if err == nil {
		t.Error("expected error when config file doesn't exist")
	}
}

func TestLoadConfigInitializesLibrariesMap(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "no-libs.yaml")

	// Create config without libraries field
	configData := `
destination: "./frontend"
project_name: "test"
`
	os.WriteFile(configPath, []byte(configData), 0644)

	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Should initialize empty map
	if config.Libraries == nil {
		t.Error("expected Libraries map to be initialized")
	}

	if len(config.Libraries) != 0 {
		t.Errorf("expected empty Libraries map, got %d entries", len(config.Libraries))
	}
}

func TestSaveConfigPreservesAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "full-config.yaml")

	// Create config with all optional fields
	testConfig := frontend_config.FrontendConfig{
		Destination: "./public/{library_name}",
		ProjectName: "comprehensive-test",
		CDN:         frontend_config.CDNJsdelivr,
		Libraries: map[string]frontend_config.LibraryConfig{
			"bootstrap": {
				Version:    "5.3.0",
				CDN:        frontend_config.CDNCdnjs,
				Files:      []string{"css/bootstrap.min.css", "js/bootstrap.bundle.min.js"},
				OutputPath: "./custom/bootstrap",
			},
			"jquery": {
				Version: "3.7.1",
			},
		},
	}

	// Save
	err := saveConfig(configPath, &testConfig)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load and verify all fields
	loaded, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Destination != testConfig.Destination {
		t.Errorf("destination mismatch: expected %q, got %q", testConfig.Destination, loaded.Destination)
	}

	if loaded.ProjectName != testConfig.ProjectName {
		t.Errorf("project name mismatch: expected %q, got %q", testConfig.ProjectName, loaded.ProjectName)
	}

	if loaded.CDN != testConfig.CDN {
		t.Errorf("CDN mismatch: expected %q, got %q", testConfig.CDN, loaded.CDN)
	}

	bootstrap := loaded.Libraries["bootstrap"]
	if bootstrap.Version != "5.3.0" {
		t.Errorf("bootstrap version mismatch: expected %q, got %q", "5.3.0", bootstrap.Version)
	}

	if bootstrap.CDN != frontend_config.CDNCdnjs {
		t.Errorf("bootstrap CDN mismatch: expected %q, got %q", frontend_config.CDNCdnjs, bootstrap.CDN)
	}

	if len(bootstrap.Files) != 2 {
		t.Errorf("bootstrap files count mismatch: expected 2, got %d", len(bootstrap.Files))
	}

	if bootstrap.OutputPath != "./custom/bootstrap" {
		t.Errorf("bootstrap output path mismatch: expected %q, got %q", "./custom/bootstrap", bootstrap.OutputPath)
	}
}

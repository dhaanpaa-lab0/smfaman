package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

func TestDeleteLibraryFromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "delete-test.yaml")

	// Create config with multiple libraries
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		ProjectName: "delete-test",
		CDN:         frontend_config.CDNUnpkg,
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery":    {Version: "3.7.1"},
			"bootstrap": {Version: "5.3.0"},
			"react":     {Version: "18.2.0"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	// Set global config path
	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Delete one library
	err := deleteLibraryFromConfig("jquery")
	if err != nil {
		t.Fatalf("failed to delete library: %v", err)
	}

	// Verify library was deleted
	config, err := loadConfigForDelete(configPath)
	if err != nil {
		t.Fatalf("failed to load config after delete: %v", err)
	}

	if _, exists := config.Libraries["jquery"]; exists {
		t.Error("jquery should have been deleted from config")
	}

	// Verify other libraries still exist
	if _, exists := config.Libraries["bootstrap"]; !exists {
		t.Error("bootstrap should still exist in config")
	}

	if _, exists := config.Libraries["react"]; !exists {
		t.Error("react should still exist in config")
	}

	if len(config.Libraries) != 2 {
		t.Errorf("expected 2 libraries remaining, got %d", len(config.Libraries))
	}
}

func TestDeleteLibraryFromConfigNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "delete-nonexistent.yaml")

	// Create config without the library we'll try to delete
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.7.1"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Try to delete non-existent library
	err := deleteLibraryFromConfig("bootstrap")
	if err == nil {
		t.Error("expected error when deleting non-existent library")
	}
}

func TestDeleteLibraryFromConfigMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	err := deleteLibraryFromConfig("react")
	if err == nil {
		t.Error("expected error when config file doesn't exist")
	}
}

func TestDeleteLibraryPreservesConfigStructure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "preserve-structure.yaml")

	// Create config with all fields populated
	testConfig := frontend_config.FrontendConfig{
		Destination: "./public/{library_name}",
		ProjectName: "preservation-test",
		CDN:         frontend_config.CDNCdnjs,
		Libraries: map[string]frontend_config.LibraryConfig{
			"bootstrap": {
				Version:    "5.3.0",
				CDN:        frontend_config.CDNJsdelivr,
				Files:      []string{"css/bootstrap.min.css"},
				OutputPath: "./custom/bootstrap",
			},
			"jquery": {
				Version: "3.7.1",
			},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Delete jquery
	err := deleteLibraryFromConfig("jquery")
	if err != nil {
		t.Fatalf("failed to delete library: %v", err)
	}

	// Load and verify all other fields are preserved
	config, err := loadConfigForDelete(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if config.Destination != testConfig.Destination {
		t.Errorf("destination was not preserved: expected %q, got %q", testConfig.Destination, config.Destination)
	}

	if config.ProjectName != testConfig.ProjectName {
		t.Errorf("project name was not preserved: expected %q, got %q", testConfig.ProjectName, config.ProjectName)
	}

	if config.CDN != testConfig.CDN {
		t.Errorf("CDN was not preserved: expected %q, got %q", testConfig.CDN, config.CDN)
	}

	// Verify bootstrap and its properties are intact
	bootstrap, exists := config.Libraries["bootstrap"]
	if !exists {
		t.Fatal("bootstrap library was deleted unexpectedly")
	}

	if bootstrap.Version != "5.3.0" {
		t.Errorf("bootstrap version changed: expected %q, got %q", "5.3.0", bootstrap.Version)
	}

	if bootstrap.CDN != frontend_config.CDNJsdelivr {
		t.Errorf("bootstrap CDN changed: expected %q, got %q", frontend_config.CDNJsdelivr, bootstrap.CDN)
	}

	if len(bootstrap.Files) != 1 || bootstrap.Files[0] != "css/bootstrap.min.css" {
		t.Error("bootstrap files were not preserved correctly")
	}

	if bootstrap.OutputPath != "./custom/bootstrap" {
		t.Errorf("bootstrap output path changed: expected %q, got %q", "./custom/bootstrap", bootstrap.OutputPath)
	}
}

func TestLoadConfigForDelete(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "load-delete-test.yaml")

	// Test with non-existent file
	_, err := loadConfigForDelete(configPath)
	if err == nil {
		t.Error("expected error when loading non-existent config")
	}

	// Create valid config
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend/{library_name}",
		ProjectName: "load-test",
		CDN:         frontend_config.CDNUnpkg,
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.7.1"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	// Test loading valid config
	config, err := loadConfigForDelete(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if config.Destination != testConfig.Destination {
		t.Errorf("expected destination %q, got %q", testConfig.Destination, config.Destination)
	}

	if len(config.Libraries) != 1 {
		t.Errorf("expected 1 library, got %d", len(config.Libraries))
	}
}

func TestLoadConfigForDeleteInitializesLibrariesMap(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "no-libs-delete.yaml")

	// Create config without libraries field
	configData := `
destination: "./frontend"
project_name: "test"
`
	os.WriteFile(configPath, []byte(configData), 0644)

	config, err := loadConfigForDelete(configPath)
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

func TestLoadConfigForDeleteWithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-delete.yaml")

	// Write invalid YAML
	os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)

	_, err := loadConfigForDelete(configPath)
	if err == nil {
		t.Error("expected error when loading invalid YAML")
	}
}

func TestSaveConfigForDelete(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "save-delete-test.yaml")

	testConfig := frontend_config.FrontendConfig{
		Destination: "./public/libs",
		ProjectName: "save-delete-test",
		CDN:         frontend_config.CDNCdnjs,
		Libraries: map[string]frontend_config.LibraryConfig{
			"bootstrap": {
				Version: "5.3.0",
				Files:   []string{"css/bootstrap.min.css"},
			},
		},
	}

	// Save config
	err := saveConfigForDelete(configPath, &testConfig)
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

func TestDeleteLastLibrary(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "delete-last.yaml")

	// Create config with only one library
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.7.1"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Delete the only library
	err := deleteLibraryFromConfig("jquery")
	if err != nil {
		t.Fatalf("failed to delete last library: %v", err)
	}

	// Verify libraries map is empty
	config, err := loadConfigForDelete(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(config.Libraries) != 0 {
		t.Errorf("expected 0 libraries, got %d", len(config.Libraries))
	}

	// Libraries map should still exist, just be empty
	if config.Libraries == nil {
		t.Error("Libraries map should exist even when empty")
	}
}

func TestDeleteScopedPackage(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "scoped-delete.yaml")

	// Create config with scoped package
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		Libraries: map[string]frontend_config.LibraryConfig{
			"@babel/core": {Version: "7.22.0"},
			"jquery":      {Version: "3.7.1"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Delete scoped package
	err := deleteLibraryFromConfig("@babel/core")
	if err != nil {
		t.Fatalf("failed to delete scoped package: %v", err)
	}

	// Verify scoped package was deleted
	config, err := loadConfigForDelete(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if _, exists := config.Libraries["@babel/core"]; exists {
		t.Error("@babel/core should have been deleted from config")
	}

	// Verify regular package still exists
	if _, exists := config.Libraries["jquery"]; !exists {
		t.Error("jquery should still exist in config")
	}
}

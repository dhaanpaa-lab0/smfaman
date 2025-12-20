package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

func TestUpgradeSpecificLibraryNotInConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "upgrade-not-found.yaml")

	// Create config without the library we'll try to upgrade
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.5.1"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Try to upgrade non-existent library
	err := upgradeSpecificLibrary("bootstrap")
	if err == nil {
		t.Error("expected error when upgrading library not in config")
	}
}

func TestUpgradeSpecificLibraryMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	err := upgradeSpecificLibrary("react")
	if err == nil {
		t.Error("expected error when config file doesn't exist")
	}
}

func TestUpgradeAllLibrariesEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty-libs.yaml")

	// Create config with no libraries
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		Libraries:   make(map[string]frontend_config.LibraryConfig),
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Should handle gracefully
	err := upgradeAllLibraries()
	if err != nil {
		t.Fatalf("should not error with empty libraries: %v", err)
	}
}

func TestUpgradeAllLibrariesMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	err := upgradeAllLibraries()
	if err == nil {
		t.Error("expected error when config file doesn't exist")
	}
}

func TestLoadConfigForUpgrade(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "load-upgrade-test.yaml")

	// Test with non-existent file
	_, err := loadConfigForUpgrade(configPath)
	if err == nil {
		t.Error("expected error when loading non-existent config")
	}

	// Create valid config
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend/{library_name}",
		ProjectName: "upgrade-test",
		CDN:         frontend_config.CDNUnpkg,
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.5.1"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	// Test loading valid config
	config, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if config.Destination != testConfig.Destination {
		t.Errorf("expected destination %q, got %q", testConfig.Destination, config.Destination)
	}

	if len(config.Libraries) != 1 {
		t.Errorf("expected 1 library, got %d", len(config.Libraries))
	}

	if config.Libraries["jquery"].Version != "3.5.1" {
		t.Errorf("expected jquery version '3.5.1', got %q", config.Libraries["jquery"].Version)
	}
}

func TestSaveConfigForUpgrade(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "save-upgrade-test.yaml")

	testConfig := frontend_config.FrontendConfig{
		Destination: "./public/libs",
		ProjectName: "save-upgrade-test",
		CDN:         frontend_config.CDNCdnjs,
		Libraries: map[string]frontend_config.LibraryConfig{
			"bootstrap": {
				Version: "5.3.2",
				Files:   []string{"css/bootstrap.min.css"},
			},
		},
	}

	// Save config
	err := saveConfigForUpgrade(configPath, &testConfig)
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

	if loadedConfig.Libraries["bootstrap"].Version != "5.3.2" {
		t.Errorf("expected bootstrap version '5.3.2', got %q", loadedConfig.Libraries["bootstrap"].Version)
	}
}

func TestUpgradePreservesLibraryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "preserve-config.yaml")

	// Create config with library that has extra config
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		Libraries: map[string]frontend_config.LibraryConfig{
			"bootstrap": {
				Version:    "5.2.0",
				CDN:        frontend_config.CDNJsdelivr,
				Files:      []string{"css/bootstrap.min.css", "js/bootstrap.bundle.min.js"},
				OutputPath: "./custom/bootstrap",
			},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	// Load config
	config, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Simulate upgrade by changing version
	libConfig := config.Libraries["bootstrap"]
	libConfig.Version = "5.3.0"
	config.Libraries["bootstrap"] = libConfig

	// Save config
	err = saveConfigForUpgrade(configPath, config)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load again and verify all fields preserved
	reloadedConfig, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	bootstrap := reloadedConfig.Libraries["bootstrap"]

	if bootstrap.Version != "5.3.0" {
		t.Errorf("version not updated: expected '5.3.0', got %q", bootstrap.Version)
	}

	if bootstrap.CDN != frontend_config.CDNJsdelivr {
		t.Errorf("CDN not preserved: expected %q, got %q", frontend_config.CDNJsdelivr, bootstrap.CDN)
	}

	if len(bootstrap.Files) != 2 {
		t.Errorf("files not preserved: expected 2, got %d", len(bootstrap.Files))
	}

	if bootstrap.OutputPath != "./custom/bootstrap" {
		t.Errorf("output path not preserved: expected './custom/bootstrap', got %q", bootstrap.OutputPath)
	}
}

func TestUpgradeScopedPackage(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "scoped-upgrade.yaml")

	// Create config with scoped package
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		CDN:         frontend_config.CDNUnpkg,
		Libraries: map[string]frontend_config.LibraryConfig{
			"@babel/core": {Version: "7.20.0"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	// Load config
	config, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify scoped package loaded correctly
	if _, exists := config.Libraries["@babel/core"]; !exists {
		t.Error("scoped package not found in loaded config")
	}

	// Simulate upgrade
	libConfig := config.Libraries["@babel/core"]
	libConfig.Version = "7.22.0"
	config.Libraries["@babel/core"] = libConfig

	// Save config
	err = saveConfigForUpgrade(configPath, config)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Reload and verify
	reloadedConfig, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	babelCore, exists := reloadedConfig.Libraries["@babel/core"]
	if !exists {
		t.Fatal("scoped package not found after save/reload")
	}

	if babelCore.Version != "7.22.0" {
		t.Errorf("version not updated: expected '7.22.0', got %q", babelCore.Version)
	}
}

func TestUpgradeWithDifferentCDNs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "multi-cdn.yaml")

	// Create config with libraries from different CDNs
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		CDN:         frontend_config.CDNUnpkg, // Default CDN
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {
				Version: "3.5.1",
				// Uses default CDN (unpkg)
			},
			"bootstrap": {
				Version: "5.2.0",
				CDN:     frontend_config.CDNCdnjs, // Override with cdnjs
			},
			"react": {
				Version: "18.0.0",
				CDN:     frontend_config.CDNJsdelivr, // Override with jsdelivr
			},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	// Load config
	config, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify each library's CDN is correct using GetLibraryCDN
	if cdn := config.GetLibraryCDN(config.Libraries["jquery"]); cdn != frontend_config.CDNUnpkg {
		t.Errorf("jquery should use default CDN (unpkg), got %q", cdn)
	}

	if cdn := config.GetLibraryCDN(config.Libraries["bootstrap"]); cdn != frontend_config.CDNCdnjs {
		t.Errorf("bootstrap should use cdnjs, got %q", cdn)
	}

	if cdn := config.GetLibraryCDN(config.Libraries["react"]); cdn != frontend_config.CDNJsdelivr {
		t.Errorf("react should use jsdelivr, got %q", cdn)
	}

	// Simulate upgrade and verify CDN preferences are preserved
	for libName, libConfig := range config.Libraries {
		libConfig.Version = "new-version"
		config.Libraries[libName] = libConfig
	}

	// Save and reload
	err = saveConfigForUpgrade(configPath, config)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	reloadedConfig, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	// Verify CDN settings preserved
	if reloadedConfig.Libraries["jquery"].CDN != "" {
		t.Error("jquery should not have explicit CDN set (uses default)")
	}

	if reloadedConfig.Libraries["bootstrap"].CDN != frontend_config.CDNCdnjs {
		t.Errorf("bootstrap CDN not preserved: expected %q, got %q", frontend_config.CDNCdnjs, reloadedConfig.Libraries["bootstrap"].CDN)
	}

	if reloadedConfig.Libraries["react"].CDN != frontend_config.CDNJsdelivr {
		t.Errorf("react CDN not preserved: expected %q, got %q", frontend_config.CDNJsdelivr, reloadedConfig.Libraries["react"].CDN)
	}
}

func TestUpgradeMultipleLibraries(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "multi-lib-upgrade.yaml")

	// Create config with multiple libraries
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend",
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery":    {Version: "3.5.1"},
			"bootstrap": {Version: "5.2.0"},
			"react":     {Version: "18.0.0"},
			"lodash":    {Version: "4.17.20"},
		},
	}

	data, _ := yaml.Marshal(&testConfig)
	os.WriteFile(configPath, data, 0644)

	// Load config
	config, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Simulate upgrading some libraries
	jquery := config.Libraries["jquery"]
	jquery.Version = "3.7.1"
	config.Libraries["jquery"] = jquery

	react := config.Libraries["react"]
	react.Version = "18.2.0"
	config.Libraries["react"] = react

	// bootstrap and lodash remain unchanged

	// Save config
	err = saveConfigForUpgrade(configPath, config)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Reload and verify
	reloadedConfig, err := loadConfigForUpgrade(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	// Check upgraded libraries
	if reloadedConfig.Libraries["jquery"].Version != "3.7.1" {
		t.Errorf("jquery not upgraded: expected '3.7.1', got %q", reloadedConfig.Libraries["jquery"].Version)
	}

	if reloadedConfig.Libraries["react"].Version != "18.2.0" {
		t.Errorf("react not upgraded: expected '18.2.0', got %q", reloadedConfig.Libraries["react"].Version)
	}

	// Check unchanged libraries
	if reloadedConfig.Libraries["bootstrap"].Version != "5.2.0" {
		t.Errorf("bootstrap should not change: expected '5.2.0', got %q", reloadedConfig.Libraries["bootstrap"].Version)
	}

	if reloadedConfig.Libraries["lodash"].Version != "4.17.20" {
		t.Errorf("lodash should not change: expected '4.17.20', got %q", reloadedConfig.Libraries["lodash"].Version)
	}

	// Verify all 4 libraries still exist
	if len(reloadedConfig.Libraries) != 4 {
		t.Errorf("expected 4 libraries, got %d", len(reloadedConfig.Libraries))
	}
}

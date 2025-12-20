package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

func TestClean(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create test config file
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	config := frontend_config.FrontendConfig{
		Destination: filepath.Join(tmpDir, "libs", "{library_name}"),
		ProjectName: "test-project",
		CDN:         "unpkg",
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {
				Version: "3.7.1",
			},
			"bootstrap": {
				Version: "5.3.0",
			},
			"react": {
				Version:    "18.2.0",
				OutputPath: filepath.Join(tmpDir, "custom", "react"),
			},
		},
	}

	// Write config to file
	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create destination directories
	jqueryDir := filepath.Join(tmpDir, "libs", "jquery")
	bootstrapDir := filepath.Join(tmpDir, "libs", "bootstrap")
	reactDir := filepath.Join(tmpDir, "custom", "react")

	for _, dir := range []string{jqueryDir, bootstrapDir, reactDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		// Create a dummy file in each directory
		dummyFile := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Verify directories exist before clean
	for _, dir := range []string{jqueryDir, bootstrapDir, reactDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Fatalf("Directory should exist before clean: %s", dir)
		}
	}

	// Set config path
	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Set force flag to skip confirmation
	oldForce := cleanForce
	cleanForce = true
	defer func() { cleanForce = oldForce }()

	// Run clean
	if err := runClean(); err != nil {
		t.Fatalf("runClean failed: %v", err)
	}

	// Verify directories were deleted
	for _, dir := range []string{jqueryDir, bootstrapDir, reactDir} {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("Directory should be deleted: %s", dir)
		}
	}
}

func TestCleanDryRun(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create test config file
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	config := frontend_config.FrontendConfig{
		Destination: filepath.Join(tmpDir, "libs", "{library_name}"),
		ProjectName: "test-project",
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {
				Version: "3.7.1",
			},
		},
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create destination directory
	jqueryDir := filepath.Join(tmpDir, "libs", "jquery")
	if err := os.MkdirAll(jqueryDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Set config path
	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Set dry-run flag
	oldDryRun := cleanDryRun
	cleanDryRun = true
	defer func() { cleanDryRun = oldDryRun }()

	// Run clean in dry-run mode
	if err := runClean(); err != nil {
		t.Fatalf("runClean failed: %v", err)
	}

	// Verify directory still exists (not deleted in dry-run)
	if _, err := os.Stat(jqueryDir); os.IsNotExist(err) {
		t.Errorf("Directory should still exist in dry-run mode: %s", jqueryDir)
	}
}

func TestCleanNoLibraries(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create test config file with no libraries
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	config := frontend_config.FrontendConfig{
		Destination: filepath.Join(tmpDir, "libs", "{library_name}"),
		ProjectName: "test-project",
		Libraries:   map[string]frontend_config.LibraryConfig{},
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set config path
	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Run clean - should succeed with no action
	if err := runClean(); err != nil {
		t.Fatalf("runClean should succeed with no libraries: %v", err)
	}
}

func TestCleanNonExistentDirs(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create test config file
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	config := frontend_config.FrontendConfig{
		Destination: filepath.Join(tmpDir, "libs", "{library_name}"),
		ProjectName: "test-project",
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {
				Version: "3.7.1",
			},
		},
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Don't create the destination directory

	// Set config path
	oldConfig := FrontendConfig
	FrontendConfig = configPath
	defer func() { FrontendConfig = oldConfig }()

	// Set force flag
	oldForce := cleanForce
	cleanForce = true
	defer func() { cleanForce = oldForce }()

	// Run clean - should succeed even though directory doesn't exist
	if err := runClean(); err != nil {
		t.Fatalf("runClean should succeed even with non-existent directories: %v", err)
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count    int
		singular string
		plural   string
		expected string
	}{
		{0, "y", "ies", "ies"},
		{1, "y", "ies", "y"},
		{2, "y", "ies", "ies"},
		{10, "y", "ies", "ies"},
	}

	for _, tt := range tests {
		result := pluralize(tt.count, tt.singular, tt.plural)
		if result != tt.expected {
			t.Errorf("pluralize(%d, %q, %q) = %q, want %q", tt.count, tt.singular, tt.plural, result, tt.expected)
		}
	}
}

func TestGetActionVerb(t *testing.T) {
	tests := []struct {
		dryRun   bool
		expected string
	}{
		{true, "deleted (dry run)"},
		{false, "deleted"},
	}

	for _, tt := range tests {
		result := getActionVerb(tt.dryRun)
		if result != tt.expected {
			t.Errorf("getActionVerb(%v) = %q, want %q", tt.dryRun, result, tt.expected)
		}
	}
}

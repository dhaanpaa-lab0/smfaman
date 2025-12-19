package cmd

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

func TestInitConfigCreation(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "/tmp/test-frontend-config.yaml"
	defer os.Remove(tmpFile)

	// Test data
	projectName := "test-project"
	destination := "./public/libs/{library_name}"
	cdn := frontend_config.CDNJsdelivr

	// Create config
	config := frontend_config.FrontendConfig{
		ProjectName: projectName,
		Destination: destination,
		CDN:         cdn,
		Libraries:   make(map[string]frontend_config.LibraryConfig),
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	// Write to file
	err = os.WriteFile(tmpFile, data, 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Read back and verify
	var readConfig frontend_config.FrontendConfig
	fileData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(fileData, &readConfig)
	if err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	// Verify fields
	if readConfig.ProjectName != projectName {
		t.Errorf("expected project name %q, got %q", projectName, readConfig.ProjectName)
	}
	if readConfig.Destination != destination {
		t.Errorf("expected destination %q, got %q", destination, readConfig.Destination)
	}
	if readConfig.CDN != cdn {
		t.Errorf("expected CDN %q, got %q", cdn, readConfig.CDN)
	}
	if readConfig.Libraries == nil {
		t.Error("expected libraries map to be initialized, got nil")
	}
}

func TestNewInitModel(t *testing.T) {
	configFile := "test-config.yaml"
	model := newInitModel(configFile)

	// Verify model initialization
	if model.configFile != configFile {
		t.Errorf("expected config file %q, got %q", configFile, model.configFile)
	}

	if len(model.inputs) != fieldCount {
		t.Errorf("expected %d inputs, got %d", fieldCount, len(model.inputs))
	}

	if len(model.cdnOptions) != 3 {
		t.Errorf("expected 3 CDN options, got %d", len(model.cdnOptions))
	}

	expectedCDNs := []string{"unpkg", "cdnjs", "jsdelivr"}
	for i, expected := range expectedCDNs {
		if model.cdnOptions[i] != expected {
			t.Errorf("expected CDN option[%d] to be %q, got %q", i, expected, model.cdnOptions[i])
		}
	}

	// Verify default placeholders
	if model.inputs[fieldProjectName].Placeholder != "my-project" {
		t.Errorf("unexpected project name placeholder: %q", model.inputs[fieldProjectName].Placeholder)
	}

	if model.inputs[fieldDestination].Placeholder != "./frontend/{library_name}" {
		t.Errorf("unexpected destination placeholder: %q", model.inputs[fieldDestination].Placeholder)
	}
}

func TestForceOverwrite(t *testing.T) {
	// Create a temporary file
	tmpFile := "/tmp/test-force-config.yaml"
	initialContent := "existing: content\n"
	err := os.WriteFile(tmpFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("test file should exist")
	}

	// Test overwrite with new content
	newProjectName := "force-test-project"
	newDestination := "./forced/{library_name}"
	newCDN := frontend_config.CDNUnpkg

	config := frontend_config.FrontendConfig{
		ProjectName: newProjectName,
		Destination: newDestination,
		CDN:         newCDN,
		Libraries:   make(map[string]frontend_config.LibraryConfig),
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	// Overwrite the file (simulating --force behavior)
	err = os.WriteFile(tmpFile, data, 0644)
	if err != nil {
		t.Fatalf("failed to overwrite config file: %v", err)
	}

	// Read back and verify it was overwritten
	var readConfig frontend_config.FrontendConfig
	fileData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Verify it's not the old content
	if string(fileData) == initialContent {
		t.Error("file was not overwritten")
	}

	err = yaml.Unmarshal(fileData, &readConfig)
	if err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	// Verify new content
	if readConfig.ProjectName != newProjectName {
		t.Errorf("expected project name %q, got %q", newProjectName, readConfig.ProjectName)
	}
	if readConfig.Destination != newDestination {
		t.Errorf("expected destination %q, got %q", newDestination, readConfig.Destination)
	}
	if readConfig.CDN != newCDN {
		t.Errorf("expected CDN %q, got %q", newCDN, readConfig.CDN)
	}
}

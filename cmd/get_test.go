package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

func TestDownloadAndSaveConfig(t *testing.T) {
	// Create a test config
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend/{library_name}",
		ProjectName: "test-project",
		CDN:         frontend_config.CDNUnpkg,
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.5.1"},
			"react":  {Version: "18.2.0", CDN: frontend_config.CDNCdnjs},
		},
	}

	// Marshal to YAML
	configData, err := yaml.Marshal(&testConfig)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		w.Write(configData)
	}))
	defer server.Close()

	// Create temp file path
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "downloaded-config.yaml")

	// Test download
	err = downloadAndSaveConfig(server.URL, targetPath)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Fatal("downloaded file does not exist")
	}

	// Read and verify content
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}

	var downloadedConfig frontend_config.FrontendConfig
	if err := yaml.Unmarshal(data, &downloadedConfig); err != nil {
		t.Fatalf("failed to unmarshal downloaded config: %v", err)
	}

	// Verify config content
	if downloadedConfig.ProjectName != testConfig.ProjectName {
		t.Errorf("project name mismatch: expected %s, got %s", testConfig.ProjectName, downloadedConfig.ProjectName)
	}

	if downloadedConfig.Destination != testConfig.Destination {
		t.Errorf("destination mismatch: expected %s, got %s", testConfig.Destination, downloadedConfig.Destination)
	}

	if len(downloadedConfig.Libraries) != len(testConfig.Libraries) {
		t.Errorf("libraries count mismatch: expected %d, got %d", len(testConfig.Libraries), len(downloadedConfig.Libraries))
	}
}

func TestDownloadAndSaveConfigWithExistingFile(t *testing.T) {
	// Create test server with valid config
	testConfig := frontend_config.FrontendConfig{
		Destination: "./frontend/{library_name}",
		Libraries:   map[string]frontend_config.LibraryConfig{},
	}

	configData, _ := yaml.Marshal(&testConfig)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(configData)
	}))
	defer server.Close()

	// Create existing file
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "existing-config.yaml")
	os.WriteFile(targetPath, []byte("existing content"), 0644)

	// Test download without force (should fail)
	getForce = false
	err := downloadAndSaveConfig(server.URL, targetPath)
	if err == nil {
		t.Error("expected error when file exists without --force, got nil")
	}

	// Test download with force (should succeed)
	getForce = true
	err = downloadAndSaveConfig(server.URL, targetPath)
	if err != nil {
		t.Errorf("download with force failed: %v", err)
	}

	// Reset force flag
	getForce = false
}

func TestDownloadAndSaveConfigInvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "config.yaml")

	tests := []struct {
		name string
		url  string
	}{
		{"invalid scheme", "ftp://example.com/config.yaml"},
		{"malformed url", "ht!tp://invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := downloadAndSaveConfig(tt.url, targetPath)
			if err == nil {
				t.Error("expected error for invalid URL, got nil")
			}
		})
	}
}

func TestDownloadAndSaveConfigInvalidYAML(t *testing.T) {
	// Create test server with invalid YAML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write([]byte("invalid: yaml: content: ["))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "config.yaml")

	err := downloadAndSaveConfig(server.URL, targetPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestDownloadAndSaveConfigMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		config interface{}
	}{
		{
			name: "missing destination",
			config: map[string]interface{}{
				"project_name": "test",
				"libraries":    map[string]interface{}{},
			},
		},
		{
			name: "missing libraries",
			config: map[string]interface{}{
				"destination":  "./frontend",
				"project_name": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configData, _ := yaml.Marshal(tt.config)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/yaml")
				w.Write(configData)
			}))
			defer server.Close()

			tmpDir := t.TempDir()
			targetPath := filepath.Join(tmpDir, "config.yaml")

			err := downloadAndSaveConfig(server.URL, targetPath)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestDownloadAndSaveConfigHTTPError(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "config.yaml")

	err := downloadAndSaveConfig(server.URL, targetPath)
	if err == nil {
		t.Error("expected error for HTTP 404, got nil")
	}
}

func TestDownloadAndSaveConfigValidation(t *testing.T) {
	// Create valid config with all optional fields
	testConfig := frontend_config.FrontendConfig{
		Destination: "./public/libs/{library_name}",
		ProjectName: "full-featured-project",
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

	configData, _ := yaml.Marshal(&testConfig)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(configData)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "full-config.yaml")

	err := downloadAndSaveConfig(server.URL, targetPath)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	// Verify downloaded content
	data, _ := os.ReadFile(targetPath)
	var downloadedConfig frontend_config.FrontendConfig
	yaml.Unmarshal(data, &downloadedConfig)

	// Verify complex library config
	bootstrap := downloadedConfig.Libraries["bootstrap"]
	if len(bootstrap.Files) != 2 {
		t.Errorf("expected 2 files for bootstrap, got %d", len(bootstrap.Files))
	}
	if bootstrap.OutputPath != "./custom/bootstrap" {
		t.Errorf("expected custom output path, got %s", bootstrap.OutputPath)
	}
}

// createTestConfigYAML creates a sample config for testing purposes
func createTestConfigYAML() string {
	config := frontend_config.FrontendConfig{
		Destination: "./frontend/{library_name}",
		ProjectName: "example-project",
		CDN:         frontend_config.CDNUnpkg,
		Libraries: map[string]frontend_config.LibraryConfig{
			"jquery": {Version: "3.7.1"},
			"react":  {Version: "18.2.0"},
		},
	}

	data, _ := yaml.Marshal(&config)
	return string(data)
}

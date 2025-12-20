package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetBinaryName(t *testing.T) {
	name := getBinaryName()

	if runtime.GOOS == "windows" {
		if name != "smfaman.exe" {
			t.Errorf("Expected smfaman.exe on Windows, got %s", name)
		}
	} else {
		if name != "smfaman" {
			t.Errorf("Expected smfaman on Unix, got %s", name)
		}
	}
}

func TestDetectShell(t *testing.T) {
	// Save original SHELL env
	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)

	tests := []struct {
		shellPath string
		expected  string
	}{
		{"/bin/bash", "bash"},
		{"/usr/bin/zsh", "zsh"},
		{"/usr/local/bin/fish", "fish"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		os.Setenv("SHELL", tt.shellPath)
		result := detectShell()
		if result != tt.expected {
			t.Errorf("detectShell() with SHELL=%s: got %s, want %s", tt.shellPath, result, tt.expected)
		}
	}
}

func TestGetShellConfig(t *testing.T) {
	homeDir := t.TempDir()
	binDir := filepath.Join(homeDir, "bin")

	tests := []struct {
		shell              string
		expectedConfigName string
		checkExportLine    bool
	}{
		{"bash", ".bashrc", true},
		{"zsh", ".zshrc", true},
		{"fish", filepath.Join(".config", "fish", "config.fish"), true},
		{"ksh", ".kshrc", true},
		{"unknown", ".profile", true},
	}

	for _, tt := range tests {
		configFile, exportLine := getShellConfig(tt.shell, binDir)

		// Check that config file path is reasonable
		if !strings.Contains(configFile, tt.expectedConfigName) && tt.shell != "bash" {
			// bash is special on macOS
			if !(tt.shell == "bash" && runtime.GOOS == "darwin") {
				t.Errorf("getShellConfig(%s): config file %s doesn't contain %s", tt.shell, configFile, tt.expectedConfigName)
			}
		}

		// Check that export line is not empty
		if tt.checkExportLine && exportLine == "" {
			t.Errorf("getShellConfig(%s): export line is empty", tt.shell)
		}

		// Check that export line contains PATH
		if tt.checkExportLine && !strings.Contains(exportLine, "PATH") {
			t.Errorf("getShellConfig(%s): export line doesn't contain PATH: %s", tt.shell, exportLine)
		}

		// Check that export line contains bin directory
		if tt.checkExportLine && !strings.Contains(exportLine, binDir) {
			t.Errorf("getShellConfig(%s): export line doesn't contain binDir: %s", tt.shell, exportLine)
		}

		// Check shell-specific syntax
		switch tt.shell {
		case "fish":
			if !strings.Contains(exportLine, "set -gx PATH") {
				t.Errorf("fish export line should use 'set -gx PATH': %s", exportLine)
			}
		case "tcsh", "csh":
			if !strings.Contains(exportLine, "setenv PATH") {
				t.Errorf("tcsh/csh export line should use 'setenv PATH': %s", exportLine)
			}
		default:
			if !strings.Contains(exportLine, "export PATH") {
				t.Errorf("%s export line should use 'export PATH': %s", tt.shell, exportLine)
			}
		}
	}
}

func TestIsInPath(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testbin")

	// Test when directory is not in PATH
	os.Setenv("PATH", "/usr/bin:/bin")
	if isInPath(testDir) {
		t.Errorf("isInPath should return false when directory is not in PATH")
	}

	// Test when directory is in PATH
	pathSep := string(os.PathListSeparator)
	os.Setenv("PATH", "/usr/bin"+pathSep+testDir+pathSep+"/bin")
	if !isInPath(testDir) {
		t.Errorf("isInPath should return true when directory is in PATH")
	}

	// Test with trailing slash
	os.Setenv("PATH", "/usr/bin"+pathSep+testDir+string(filepath.Separator)+pathSep+"/bin")
	if !isInPath(testDir) {
		t.Errorf("isInPath should return true even with trailing slash")
	}
}

func TestIsInConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".bashrc")
	binDir := filepath.Join(tmpDir, "bin")

	// Test with non-existent file
	if isInConfigFile(configFile, binDir) {
		t.Errorf("isInConfigFile should return false for non-existent file")
	}

	// Test with file that doesn't contain PATH
	content := "# Some other content\nalias ll='ls -la'\n"
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	if isInConfigFile(configFile, binDir) {
		t.Errorf("isInConfigFile should return false when PATH is not in file")
	}

	// Test with file that contains PATH export
	content = fmt.Sprintf("export PATH=\"%s:$PATH\"\n", binDir)
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	if !isInConfigFile(configFile, binDir) {
		t.Errorf("isInConfigFile should return true when PATH export is in file")
	}

	// Test with $HOME/bin pattern
	content = "export PATH=\"$HOME/bin:$PATH\"\n"
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	if !isInConfigFile(configFile, binDir) {
		t.Errorf("isInConfigFile should return true for $HOME/bin pattern")
	}
}

func TestAddToConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".bashrc")
	exportLine := "\n# Test export\nexport PATH=\"$HOME/bin:$PATH\"\n"

	// Test adding to non-existent file
	if err := addToConfigFile(configFile, exportLine); err != nil {
		t.Errorf("addToConfigFile should create file if it doesn't exist: %v", err)
	}

	// Verify content was written
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if !strings.Contains(string(content), exportLine) {
		t.Errorf("Config file should contain export line")
	}

	// Test appending to existing file
	if err := addToConfigFile(configFile, "# Another line\n"); err != nil {
		t.Errorf("addToConfigFile should append to existing file: %v", err)
	}

	// Verify both lines are present
	content, err = os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if !strings.Contains(string(content), exportLine) {
		t.Errorf("Config file should still contain original export line")
	}

	if !strings.Contains(string(content), "# Another line") {
		t.Errorf("Config file should contain appended line")
	}
}

func TestCreateBinDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")

	// Test creating directory
	if err := createBinDirectory(binDir); err != nil {
		t.Errorf("createBinDirectory failed: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(binDir)
	if err != nil {
		t.Errorf("bin directory should exist: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("bin should be a directory")
	}

	// Test with existing directory (should not error)
	if err := createBinDirectory(binDir); err != nil {
		t.Errorf("createBinDirectory should succeed with existing directory: %v", err)
	}
}

func TestCopyBinary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source binary
	srcPath := filepath.Join(tmpDir, "source")
	srcContent := []byte("test binary content")
	if err := os.WriteFile(srcPath, srcContent, 0755); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test copying binary
	destPath := filepath.Join(tmpDir, "dest")
	if err := copyBinary(srcPath, destPath); err != nil {
		t.Errorf("copyBinary failed: %v", err)
	}

	// Verify destination exists and has same content
	destContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}

	if string(destContent) != string(srcContent) {
		t.Errorf("Destination content doesn't match source")
	}

	// Verify permissions (should be executable)
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("Failed to stat destination: %v", err)
	}

	mode := info.Mode()
	if mode&0111 == 0 {
		t.Errorf("Destination file should be executable, got mode: %v", mode)
	}

	// Test copying when destination exists without force flag
	oldForce := installForce
	installForce = false
	defer func() { installForce = oldForce }()

	if err := copyBinary(srcPath, destPath); err == nil {
		t.Errorf("copyBinary should fail when destination exists and force=false")
	}

	// Test with force flag
	installForce = true
	if err := copyBinary(srcPath, destPath); err != nil {
		t.Errorf("copyBinary should succeed with force=true: %v", err)
	}
}

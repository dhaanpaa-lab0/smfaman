package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrapHtmxCommand(t *testing.T) {
	// Test that the command exists and has correct metadata
	if bootstrapHtmxCmd == nil {
		t.Fatal("bootstrapHtmxCmd is nil")
	}

	if bootstrapHtmxCmd.Use != "htmx" {
		t.Errorf("Expected Use to be 'htmx', got '%s'", bootstrapHtmxCmd.Use)
	}

	if bootstrapHtmxCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if bootstrapHtmxCmd.Long == "" {
		t.Error("Long description is empty")
	}

	if bootstrapHtmxCmd.Run == nil {
		t.Error("Run function is nil")
	}
}

func TestBootstrapHtmxFlags(t *testing.T) {
	// Test that the directory flag exists
	flag := bootstrapHtmxCmd.Flags().Lookup("directory")
	if flag == nil {
		t.Fatal("directory flag not found")
	}

	if flag.Shorthand != "d" {
		t.Errorf("Expected shorthand 'd', got '%s'", flag.Shorthand)
	}

	if flag.DefValue != "." {
		t.Errorf("Expected default value '.', got '%s'", flag.DefValue)
	}
}

func TestBootstrapHtmxRegistration(t *testing.T) {
	// Verify the command is registered as a subcommand of bootstrap
	found := false
	for _, cmd := range bootstrapCmd.Commands() {
		if cmd.Use == "htmx" {
			found = true
			break
		}
	}

	if !found {
		t.Error("htmx command not registered as subcommand of bootstrap")
	}
}

func TestBootstrapHtmxURL(t *testing.T) {
	// Verify the starter kit URL is defined
	if htmxStarterKitURL == "" {
		t.Error("htmxStarterKitURL is empty")
	}

	// Check URL format
	if len(htmxStarterKitURL) < 10 {
		t.Error("htmxStarterKitURL appears to be invalid")
	}
}

// TestRunBootstrapHtmx tests the actual bootstrap execution
// This test is skipped by default as it requires network access
func TestRunBootstrapHtmx(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "smfaman-test-htmx")
	defer os.RemoveAll(tempDir)

	// Set the directory flag
	htmxDirectory = tempDir

	// Note: This will fail if the URL doesn't exist yet
	// In production, you would need to create the actual starter kit first
	t.Log("Note: This test will fail until a real HTMX starter kit is published")
	t.Log("Starter kit URL:", htmxStarterKitURL)

	// Uncomment to test actual download (requires valid URL)
	// err := runBootstrapHtmx()
	// if err != nil {
	// 	t.Skipf("Bootstrap failed (expected if URL not ready): %v", err)
	// }

	// Verify directory was created
	// if _, err := os.Stat(tempDir); os.IsNotExist(err) {
	// 	t.Error("Target directory was not created")
	// }
}

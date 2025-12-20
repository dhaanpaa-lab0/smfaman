package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	htmxDirectory string
)

// bootstrapHtmxCmd represents the bootstrap htmx command
var bootstrapHtmxCmd = &cobra.Command{
	Use:   "htmx",
	Short: "Bootstrap a new HTMX project",
	Long: `Bootstrap a new HTMX project by downloading and extracting a
starter kit with HTMX and a simple backend server.

This starter kit includes:
  - HTMX library: The core htmx.js file for dynamic HTML
  - Sample HTML templates: Example pages demonstrating HTMX features
  - Backend server: A simple development server (Go-based)
  - Static assets: Basic CSS and JavaScript files

The starter kit will be downloaded from GitHub and extracted to the
specified directory (defaults to current directory).

After extraction, you can start the development server by running:
  - On Mac/Linux/WSL: ./start.sh
  - On Windows: start.bat

Example:
  smfaman bootstrap htmx
  smfaman bootstrap htmx --directory my-htmx-app`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBootstrapHtmx(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	bootstrapCmd.AddCommand(bootstrapHtmxCmd)
	bootstrapHtmxCmd.Flags().StringVarP(&htmxDirectory, "directory", "d", ".", "Directory to extract the HTMX starter kit (default: current directory)")
}

const (
	// URL for htmx starter kit - update this to point to an actual release
	// Example: a Go + HTMX starter template with a simple server
	htmxStarterKitURL = "https://github.com/htmx-go/starter-template/releases/latest/download/htmx-starter.zip"
)

// runBootstrapHtmx executes the HTMX bootstrap command
func runBootstrapHtmx() error {
	fmt.Println("🚀 Bootstrapping HTMX project...")
	fmt.Printf("📦 Downloading HTMX starter kit from: %s\n", htmxStarterKitURL)

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(htmxDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download the starter kit
	zipPath := filepath.Join(os.TempDir(), "htmx-starter.zip")
	if err := downloadZipFile(htmxStarterKitURL, zipPath); err != nil {
		return fmt.Errorf("failed to download starter kit: %w", err)
	}
	defer os.Remove(zipPath) // Clean up temp file

	fmt.Println("📂 Extracting files...")

	// Extract the zip file
	if err := extractZip(zipPath, htmxDirectory); err != nil {
		return fmt.Errorf("failed to extract starter kit: %w", err)
	}

	fmt.Println("\n✅ HTMX project bootstrapped successfully!")
	fmt.Printf("\n📁 Project location: %s\n", htmxDirectory)
	fmt.Println("\n🎯 Next steps:")
	fmt.Println("   1. Navigate to the project directory:")

	if htmxDirectory != "." {
		fmt.Printf("      cd %s\n", htmxDirectory)
	}

	fmt.Println("\n   2. Start the development server:")
	if runtime.GOOS == "windows" {
		fmt.Println("      start.bat")
	} else {
		fmt.Println("      ./start.sh")
	}

	fmt.Println("\n   3. Open your browser to the URL shown by the server (typically http://localhost:8080)")
	fmt.Println("\n📚 Documentation: https://htmx.org/docs")
	fmt.Println("🎨 Examples: https://htmx.org/examples")

	return nil
}

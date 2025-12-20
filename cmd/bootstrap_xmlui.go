package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	xmluiDirectory string
)

// bootstrapXmluiCmd represents the bootstrap xmlui command
var bootstrapXmluiCmd = &cobra.Command{
	Use:   "xmlui",
	Short: "Bootstrap a new XMLUI project",
	Long: `Bootstrap a new XMLUI project by downloading and extracting the
official XMLUI starter kit (xmlui-invoice).

This starter kit includes:
  - XMLUI Invoice: A complete business application demonstrating XMLUI features
  - XMLUI Engine: The core framework file needed to run applications
  - Test Server: A simple local server for development and testing

The starter kit will be downloaded from the official XMLUI repository and
extracted to the specified directory (defaults to current directory).

After extraction, you can start the development server by running:
  - On Mac/Linux/WSL: ./start.sh
  - On Windows: start.bat

Example:
  smfaman bootstrap xmlui
  smfaman bootstrap xmlui --directory my-xmlui-app`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBootstrapXmlui(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	bootstrapCmd.AddCommand(bootstrapXmluiCmd)
	bootstrapXmluiCmd.Flags().StringVarP(&xmluiDirectory, "directory", "d", ".", "Directory to extract the XMLUI starter kit (default: current directory)")
}

const (
	xmluiStarterKitURL = "https://github.com/xmlui-org/xmlui-invoice/releases/latest/download/xmlui-invoice.zip"
)

// runBootstrapXmlui executes the XMLUI bootstrap command
func runBootstrapXmlui() error {
	fmt.Println("🚀 Bootstrapping XMLUI project...")
	fmt.Printf("📦 Downloading XMLUI starter kit from: %s\n", xmluiStarterKitURL)

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(xmluiDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download the starter kit
	zipPath := filepath.Join(os.TempDir(), "xmlui-invoice.zip")
	if err := downloadZipFile(xmluiStarterKitURL, zipPath); err != nil {
		return fmt.Errorf("failed to download starter kit: %w", err)
	}
	defer os.Remove(zipPath) // Clean up temp file

	fmt.Println("📂 Extracting files...")

	// Extract the zip file
	if err := extractZip(zipPath, xmluiDirectory); err != nil {
		return fmt.Errorf("failed to extract starter kit: %w", err)
	}

	fmt.Println("\n✅ XMLUI project bootstrapped successfully!")
	fmt.Printf("\n📁 Project location: %s\n", xmluiDirectory)
	fmt.Println("\n🎯 Next steps:")
	fmt.Println("   1. Navigate to the project directory:")

	if xmluiDirectory != "." {
		fmt.Printf("      cd %s\n", xmluiDirectory)
	}

	fmt.Println("\n   2. Start the development server:")
	if runtime.GOOS == "windows" {
		fmt.Println("      start.bat")
	} else {
		fmt.Println("      ./start.sh")
	}

	fmt.Println("\n   3. Open your browser to the URL shown by the server")
	fmt.Println("\n📚 Documentation: https://docs.xmlui.org")
	fmt.Println("🎨 Demo & Gallery: https://demo.xmlui.org")

	return nil
}

// downloadZipFile downloads a file from a URL to a local path
func downloadZipFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Show download progress
	fmt.Printf("⬇️  Downloading... ")

	// Write the body to file
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%.2f MB downloaded\n", float64(written)/(1024*1024))

	return nil
}

// extractZip extracts a zip archive to a destination directory
func extractZip(zipPath, destPath string) error {
	// Open the zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Extract each file
	for _, file := range reader.File {
		// Construct the full path
		path := filepath.Join(destPath, file.Name)

		// Check for ZipSlip vulnerability
		if !filepath.HasPrefix(path, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if file.FileInfo().IsDir() {
			// Create directory
			os.MkdirAll(path, file.Mode())
			continue
		}

		// Create file parent directory
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// Create the file
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		// Open the file from zip
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		// Copy contents
		_, err = io.Copy(outFile, rc)

		// Close files
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

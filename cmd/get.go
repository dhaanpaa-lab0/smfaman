package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

var (
	getForce   bool
	getTimeout int
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <url>",
	Short: "Download a frontend config from a remote HTTP server",
	Long: `Download a frontend configuration file from a remote HTTP server and save it locally.

This command fetches a YAML configuration file from the specified URL, validates it,
and saves it to the local directory. The file is saved using the name specified with
the -f flag (default: smartfrontend.yaml).

The downloaded config is validated to ensure it's a valid frontend configuration
before being saved. If the target file already exists, use --force to overwrite it.

Example:
  smfaman get https://example.com/frontend.yaml
  smfaman get https://example.com/config.yaml -f myproject.yaml
  smfaman get https://example.com/frontend.yaml --force`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configURL := args[0]

		if err := downloadAndSaveConfig(configURL, FrontendConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolVar(&getForce, "force", false, "Overwrite existing config file if it exists")
	getCmd.Flags().IntVar(&getTimeout, "timeout", 30, "HTTP request timeout in seconds")
}

// downloadAndSaveConfig downloads a config file from a URL and saves it locally
func downloadAndSaveConfig(configURL, targetPath string) error {
	// Validate URL
	parsedURL, err := url.Parse(configURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https protocol, got: %s", parsedURL.Scheme)
	}

	// Check if target file already exists
	if !getForce {
		if _, err := os.Stat(targetPath); err == nil {
			return fmt.Errorf("file %s already exists, use --force to overwrite", targetPath)
		}
	}

	fmt.Printf("Downloading config from %s...\n", configURL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(getTimeout) * time.Second,
	}

	// Download the file
	resp, err := client.Get(configURL)
	if err != nil {
		return fmt.Errorf("failed to download config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Check content type (should be YAML or text)
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && 
	   contentType != "application/yaml" && 
	   contentType != "application/x-yaml" &&
	   contentType != "text/yaml" &&
	   contentType != "text/plain" &&
	   contentType != "text/x-yaml" {
		fmt.Printf("Warning: Content-Type is %s (expected YAML)\n", contentType)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Validate that it's a valid frontend config
	var config frontend_config.FrontendConfig
	if err := yaml.Unmarshal(body, &config); err != nil {
		return fmt.Errorf("invalid frontend config format: %w", err)
	}

	// Basic validation
	if config.Destination == "" {
		return fmt.Errorf("config validation failed: destination field is required")
	}

	if config.Libraries == nil {
		return fmt.Errorf("config validation failed: libraries field is required")
	}

	// Save to file
	if err := os.WriteFile(targetPath, body, 0644); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	// Print success message with config details
	fmt.Printf("\n✓ Config downloaded successfully!\n\n")
	fmt.Printf("Saved to:     %s\n", targetPath)
	if config.ProjectName != "" {
		fmt.Printf("Project:      %s\n", config.ProjectName)
	}
	fmt.Printf("Destination:  %s\n", config.Destination)
	if config.CDN != "" {
		fmt.Printf("Default CDN:  %s\n", config.CDN)
	}
	fmt.Printf("Libraries:    %d configured\n", len(config.Libraries))

	if len(config.Libraries) > 0 {
		fmt.Println("\nConfigured libraries:")
		for name, cfg := range config.Libraries {
			cdnInfo := ""
			if cfg.CDN != "" {
				cdnInfo = fmt.Sprintf(" (CDN: %s)", cfg.CDN)
			}
			fmt.Printf("  • %s@%s%s\n", name, cfg.Version, cdnInfo)
		}
	}

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  • Review config: cat %s\n", targetPath)
	fmt.Printf("  • Sync libraries: smfaman sync\n")

	return nil
}

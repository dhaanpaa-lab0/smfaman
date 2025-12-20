package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete <package-name>",
	Aliases: []string{"del", "pkgdel", "d"},
	Short:   "Remove a library from the Smart Frontend Asset Manager Configuration",
	Long: `Remove a library from your frontend configuration file.

This command removes the specified library from the configuration file.
It does NOT delete the actual downloaded files from your filesystem.

The library will be removed from the config file specified by the -f flag
(default: smartfrontend.yaml).

Examples:
  smfaman delete react
  smfaman del bootstrap
  smfaman pkgdel jquery
  smfaman d lodash`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]

		if err := deleteLibraryFromConfig(packageName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

// deleteLibraryFromConfig removes a library from the frontend config
func deleteLibraryFromConfig(packageName string) error {
	// Load existing config
	config, err := loadConfigForDelete(FrontendConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if library exists
	libConfig, exists := config.Libraries[packageName]
	if !exists {
		return fmt.Errorf("library '%s' not found in config", packageName)
	}

	// Remove library from config
	delete(config.Libraries, packageName)

	// Save config
	if err := saveConfigForDelete(FrontendConfig, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print success message
	fmt.Printf("\n✓ Library removed successfully!\n\n")
	fmt.Printf("Package:  %s@%s\n", packageName, libConfig.Version)
	fmt.Printf("\nConfig updated: %s\n", FrontendConfig)
	fmt.Printf("\nNote: Downloaded files were not deleted from your filesystem.\n")
	fmt.Printf("To clean up files, manually delete them from the output directory.\n")

	return nil
}

// loadConfigForDelete loads a frontend config from a file
func loadConfigForDelete(path string) (*frontend_config.FrontendConfig, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file '%s' does not exist. Run 'smfaman init' first", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config frontend_config.FrontendConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure Libraries map is initialized
	if config.Libraries == nil {
		config.Libraries = make(map[string]frontend_config.LibraryConfig)
	}

	return &config, nil
}

// saveConfigForDelete saves a frontend config to a file
func saveConfigForDelete(path string, config *frontend_config.FrontendConfig) error {
	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

// pkgmgrCmd represents the pkgmgr command
var pkgmgrCmd = &cobra.Command{
	Use:   "pkgmgr",
	Short: "Interactive TUI package manager for editing frontend configuration",
	Long: `Launch an interactive Terminal User Interface (TUI) for managing your frontend
configuration file.

The package manager TUI allows you to:
  • View all libraries in your configuration
  • Add new libraries to the configuration
  • Edit existing library configurations (version, CDN, files, output path)
  • Delete libraries from the configuration
  • Edit global settings (project name, destination, default CDN)
  • Save changes back to the configuration file

Navigation:
  • Use arrow keys or tab/shift+tab to navigate
  • Press 'enter' to edit a selected library
  • Press 'a' to add a new library
  • Press 'v' or 'i' on version field to select version interactively
  • Press 'd' to delete the selected library
  • Press 'g' to edit global settings
  • Press 's' to save and quit
  • Press 'q' or 'esc' to quit without saving
  • Press 'ctrl+c' to force quit

Examples:
  smfaman pkgmgr`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPkgmgr(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pkgmgrCmd)
}

func runPkgmgr() error {
	// Load existing config
	config, err := loadConfigForPkgmgr(FrontendConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Run TUI
	p := tea.NewProgram(newPkgmgrModel(config, FrontendConfig))
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if we should save
	if m, ok := finalModel.(pkgmgrModel); ok {
		if m.saved {
			// Save config
			if err := saveConfigForPkgmgr(FrontendConfig, config); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
		}
	}

	return nil
}

// loadConfigForPkgmgr loads a frontend config from a file
func loadConfigForPkgmgr(path string) (*frontend_config.FrontendConfig, error) {
	return loadConfig(path)
}

// saveConfigForPkgmgr saves a frontend config to a file
func saveConfigForPkgmgr(path string, config *frontend_config.FrontendConfig) error {
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

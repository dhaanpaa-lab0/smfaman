/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var forceOverwrite bool

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new smart frontend asset config file",
	Long: `Initialize a new frontend asset configuration file in the current directory.

This command creates a new smartfrontend.yaml file (or the name specified with -f)
with the basic structure needed to manage your frontend dependencies.

The configuration file defines which libraries to download from CDNs, their versions,
and where to store them locally. Once initialized, you can add libraries with the
'add' command and download them with the 'sync' command.

Example:
  smfaman init
  smfaman init -f myproject.yaml
  smfaman init --force  # Overwrite existing config`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if config file already exists
		if _, err := os.Stat(FrontendConfig); err == nil && !forceOverwrite {
			fmt.Printf("Error: %s already exists\n", FrontendConfig)
			fmt.Println("Remove the existing file, use a different name with -f flag, or use --force to overwrite")
			os.Exit(1)
		}

		// Create and run the Bubble Tea program
		p := tea.NewProgram(newInitModel(FrontendConfig))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running init: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Add force flag
	initCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing config file if it exists")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

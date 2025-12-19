/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Download libraries that are defined in the configuration file but not present locally",
	Long: `Synchronize your local frontend assets with the configuration file.

This command reads your frontend configuration file and downloads any libraries
that are defined but not yet present locally. It will:
  - Check which libraries are defined in the configuration
  - Verify which ones are missing or outdated locally
  - Download missing assets from the specified CDN (jsDelivr, UNPKG, or CDNJS)
  - Verify integrity using SRI hashes where available

This is useful after cloning a project or updating the configuration file.

Example:
  smfaman sync
  smfaman sync -f myproject.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync called")
		fmt.Println(FrontendConfig)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

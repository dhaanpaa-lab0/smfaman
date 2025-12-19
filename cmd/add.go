/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new library to the Smart Frontend Asset Manager Configuration",
	Long: `Add a new frontend library to your configuration file.

This command adds a library entry to your frontend configuration, specifying
which CDN to use (jsDelivr, UNPKG, or CDNJS), the library name, version, and
where to store it locally.

The library will be added to the configuration file but not downloaded until
you run 'smfaman sync'.

Example:
  smfaman add bootstrap@5.3.0
  smfaman add react@18.2.0 --cdn unpkg
  smfaman add -f myproject.yaml lodash@4.17.21`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add called")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

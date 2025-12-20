package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	cleanDryRun bool
	cleanForce  bool
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:     "clean",
	Aliases: []string{"rm", "remove"},
	Short:   "Remove destination folders specified in the frontend configuration",
	Long: `Remove all library destination folders specified in the frontend configuration file.

This command will delete the destination directories for all libraries defined
in your configuration file. By default, it will prompt for confirmation before
deleting anything.

The command will:
  • Load the frontend configuration file
  • Resolve destination paths for each library
  • Remove each library's destination folder
  • Report what was deleted

Safety features:
  • Use --dry-run to see what would be deleted without actually deleting
  • Use --force to skip confirmation prompt
  • Only deletes directories that exist
  • Shows detailed output of operations

Examples:
  smfaman clean                    # Remove all library folders (with prompt)
  smfaman clean --dry-run          # Show what would be deleted
  smfaman clean --force            # Remove without confirmation
  smfaman clean -f smartfe.yaml    # Clean using specific config file`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runClean(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be deleted without actually deleting")
	cleanCmd.Flags().BoolVar(&cleanForce, "force", false, "Skip confirmation prompt")
}

func runClean() error {
	// Load config
	config, err := loadConfig(FrontendConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(config.Libraries) == 0 {
		fmt.Println("No libraries configured. Nothing to clean.")
		return nil
	}

	// Get all library destinations
	destinations, err := config.GetLibraryDestinations()
	if err != nil {
		return fmt.Errorf("failed to get library destinations: %w", err)
	}

	// Filter to only directories that exist
	existingDirs := make(map[string]string)
	for libName, destPath := range destinations {
		if info, err := os.Stat(destPath); err == nil && info.IsDir() {
			existingDirs[libName] = destPath
		}
	}

	if len(existingDirs) == 0 {
		fmt.Println("No destination directories found. Nothing to clean.")
		return nil
	}

	// Show what will be deleted
	fmt.Printf("Configuration file: %s\n\n", FrontendConfig)
	fmt.Printf("The following directories will be %s:\n\n", getActionVerb(cleanDryRun))
	for libName, destPath := range existingDirs {
		fmt.Printf("  • %s → %s\n", libName, destPath)
	}
	fmt.Printf("\nTotal: %d director%s\n\n", len(existingDirs), pluralize(len(existingDirs), "y", "ies"))

	// Dry run - just show and exit
	if cleanDryRun {
		fmt.Println("Dry run mode: No files were deleted.")
		return nil
	}

	// Prompt for confirmation unless --force
	if !cleanForce {
		if !promptConfirmation("Do you want to proceed?") {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete directories
	deletedCount := 0
	failedCount := 0
	for libName, destPath := range existingDirs {
		if err := os.RemoveAll(destPath); err != nil {
			fmt.Printf("✗ Failed to remove %s (%s): %v\n", libName, destPath, err)
			failedCount++
		} else {
			fmt.Printf("✓ Removed %s (%s)\n", libName, destPath)
			deletedCount++
		}
	}

	// Summary
	fmt.Printf("\n")
	fmt.Printf("Deleted: %d director%s\n", deletedCount, pluralize(deletedCount, "y", "ies"))
	if failedCount > 0 {
		fmt.Printf("Failed:  %d director%s\n", failedCount, pluralize(failedCount, "y", "ies"))
		return fmt.Errorf("failed to remove %d director%s", failedCount, pluralize(failedCount, "y", "ies"))
	}

	return nil
}

// promptConfirmation prompts the user for yes/no confirmation
func promptConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/N): ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// getActionVerb returns the appropriate verb based on dry-run mode
func getActionVerb(dryRun bool) string {
	if dryRun {
		return "deleted (dry run)"
	}
	return "deleted"
}

// pluralize returns singular or plural form based on count
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

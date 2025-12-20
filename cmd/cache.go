package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage CDN metadata and package file cache",
	Long: `Manage the local cache of CDN metadata and downloaded package files.

The cache has two components:
  1. Metadata cache: Stores API responses from CDNs (24-hour TTL)
  2. Package cache: Stores downloaded library files (no expiration)

Subcommands:
  stats          - Show cache statistics
  clear          - Clear all cached data (metadata and packages)
  clear-packages - Clear only cached package files
  clean          - Remove expired metadata cache entries`,
}

// cacheStatsCmd shows cache statistics
var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	Long:  `Display statistics about both metadata and package file caches.`,
	Run: func(cmd *cobra.Command, args []string) {
		stats, err := frontend_mgr.CacheManager.Stats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting cache stats: %v\n", err)
			os.Exit(1)
		}

		if !stats.Enabled {
			fmt.Println("Cache is disabled")
			return
		}

		fmt.Println("Cache Statistics:")
		fmt.Println("═════════════════════════════════════════")
		fmt.Printf("Cache Directory:    %s\n", stats.CacheDir)
		fmt.Printf("Metadata TTL:       %v\n", stats.TTL)
		fmt.Printf("Package Cache:      %s\n", formatEnabled(stats.PackageCache))
		fmt.Println()

		fmt.Println("Metadata Cache:")
		fmt.Println("─────────────────────────────────────────")
		fmt.Printf("Entries:            %d\n", stats.MetadataEntries)
		fmt.Printf("Expired:            %d\n", stats.ExpiredEntries)
		fmt.Printf("Size:               %s\n", formatBytes(stats.MetadataSize))
		fmt.Println()

		if stats.PackageCache {
			fmt.Println("Package Cache:")
			fmt.Println("─────────────────────────────────────────")
			fmt.Printf("Files:              %d\n", stats.PackageFiles)
			fmt.Printf("Size:               %s\n", formatBytes(stats.PackageSize))
			fmt.Println()
		}

		fmt.Printf("Total Cache Size:   %s\n", formatBytes(stats.TotalSize))

		if stats.ExpiredEntries > 0 {
			fmt.Printf("\nRun 'smfaman cache clean' to remove %d expired metadata entries\n", stats.ExpiredEntries)
		}
		if stats.PackageFiles > 0 {
			fmt.Printf("Run 'smfaman cache clear-packages' to clear %d cached package files\n", stats.PackageFiles)
		}
	},
}

// formatBytes formats byte count to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatEnabled formats boolean as enabled/disabled
func formatEnabled(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

// cacheClearCmd clears all cache
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cached data (metadata and packages)",
	Long: `Remove all cached data including CDN metadata and downloaded package files.

This will:
  • Clear all metadata cache entries
  • Clear all cached package files
  • Force fresh API calls and downloads for subsequent requests`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := frontend_mgr.CacheManager.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing metadata cache: %v\n", err)
			os.Exit(1)
		}

		if err := frontend_mgr.CacheManager.ClearPackages(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing package cache: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ All cache cleared successfully")
	},
}

// cacheClearPackagesCmd clears only package cache
var cacheClearPackagesCmd = &cobra.Command{
	Use:   "clear-packages",
	Short: "Clear cached package files",
	Long: `Remove all cached package files while keeping metadata cache intact.

Downloaded library files will need to be re-downloaded on next sync.
This is useful when you want to free up disk space but keep metadata cached.`,
	Run: func(cmd *cobra.Command, args []string) {
		stats, err := frontend_mgr.CacheManager.Stats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting cache stats: %v\n", err)
			os.Exit(1)
		}

		if stats.PackageFiles == 0 {
			fmt.Println("Package cache is already empty")
			return
		}

		fmt.Printf("Clearing %d cached package files (%s)...\n",
			stats.PackageFiles, formatBytes(stats.PackageSize))

		if err := frontend_mgr.CacheManager.ClearPackages(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing package cache: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Package cache cleared successfully")
	},
}

// cacheCleanCmd removes expired entries
var cacheCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove expired metadata cache entries",
	Long: `Remove only expired metadata cache entries while keeping valid cached data.

Package cache files are never expired and won't be affected by this command.
Use 'clear-packages' to remove cached package files.`,
	Run: func(cmd *cobra.Command, args []string) {
		removed, err := frontend_mgr.CacheManager.ClearExpired()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning cache: %v\n", err)
			os.Exit(1)
		}

		if removed == 0 {
			fmt.Println("No expired metadata entries found")
		} else {
			fmt.Printf("✓ Removed %d expired metadata cache entries\n", removed)
		}
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheClearPackagesCmd)
	cacheCmd.AddCommand(cacheCleanCmd)
}

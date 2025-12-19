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
	Short: "Manage CDN metadata cache",
	Long: `Manage the local cache of CDN metadata.

The cache stores API responses from CDNs to improve performance and reduce
API calls. Cached data has a TTL (time-to-live) of 24 hours by default.

Subcommands:
  stats  - Show cache statistics
  clear  - Clear all cached data
  clean  - Remove expired cache entries`,
}

// cacheStatsCmd shows cache statistics
var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	Long:  `Display statistics about the cache including size, entry count, and expired entries.`,
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
		fmt.Println("─────────────────────────────────────────")
		fmt.Printf("Cache Directory:  %s\n", stats.CacheDir)
		fmt.Printf("TTL:              %v\n", stats.TTL)
		fmt.Printf("Total Entries:    %d\n", stats.TotalEntries)
		fmt.Printf("Expired Entries:  %d\n", stats.ExpiredEntries)
		fmt.Printf("Total Size:       %.2f KB\n", float64(stats.TotalSize)/1024)

		if stats.ExpiredEntries > 0 {
			fmt.Printf("\nRun 'smfaman cache clean' to remove %d expired entries\n", stats.ExpiredEntries)
		}
	},
}

// cacheClearCmd clears all cache
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cached data",
	Long:  `Remove all cached CDN metadata. This will force fresh API calls for all subsequent requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := frontend_mgr.CacheManager.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing cache: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Cache cleared successfully")
	},
}

// cacheCleanCmd removes expired entries
var cacheCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove expired cache entries",
	Long:  `Remove only expired cache entries while keeping valid cached data.`,
	Run: func(cmd *cobra.Command, args []string) {
		removed, err := frontend_mgr.CacheManager.ClearExpired()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning cache: %v\n", err)
			os.Exit(1)
		}

		if removed == 0 {
			fmt.Println("No expired entries found")
		} else {
			fmt.Printf("✓ Removed %d expired cache entries\n", removed)
		}
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheCleanCmd)
}

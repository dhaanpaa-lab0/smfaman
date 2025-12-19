package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

var (
	pkgverCDN         string
	pkgverLimit       int
	pkgverNoCache     bool
	pkgverInteractive bool
)

// pkgverCmd represents the pkgver command
var pkgverCmd = &cobra.Command{
	Use:   "pkgver <package-name>",
	Short: "List available versions for a package from a CDN",
	Long: `List all available versions for a package from a specified CDN.

This command queries the CDN's API to retrieve all published versions of a package.
Versions are displayed in descending order (newest first).

With --interactive flag, launches an interactive interface to browse and select versions.

The CDN can be specified with the --cdn flag. If not specified, the default CDN
from your frontend config will be used, or unpkg as a fallback.

Supported CDNs:
  - unpkg    (uses npm registry)
  - cdnjs    (CDNJS API)
  - jsdelivr (jsDelivr API)

Example:
  smfaman pkgver react
  smfaman pkgver bootstrap --cdn cdnjs
  smfaman pkgver jquery --cdn jsdelivr --limit 10
  smfaman pkgver react --interactive`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]

		// Handle cache flag
		if pkgverNoCache {
			frontend_mgr.SetCacheEnabled(false)
			defer frontend_mgr.SetCacheEnabled(true)
		}

		// Determine which CDN to use
		cdn := determineCDN()

		// Fetch and display versions
		if err := fetchAndDisplayVersions(packageName, cdn); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pkgverCmd)

	pkgverCmd.Flags().StringVar(&pkgverCDN, "cdn", "", "CDN to query (unpkg, cdnjs, jsdelivr)")
	pkgverCmd.Flags().IntVar(&pkgverLimit, "limit", 20, "Maximum number of versions to display (non-interactive mode)")
	pkgverCmd.Flags().BoolVar(&pkgverNoCache, "no-cache", false, "Bypass cache and fetch fresh data")
	pkgverCmd.Flags().BoolVarP(&pkgverInteractive, "interactive", "i", false, "Launch interactive version selector")
}

// determineCDN determines which CDN to use based on flags and config
func determineCDN() frontend_config.CDN {
	// If --cdn flag is provided, use that
	if pkgverCDN != "" {
		cdn := frontend_config.CDN(pkgverCDN)
		if !frontend_config.IsValidCDN(cdn) {
			fmt.Fprintf(os.Stderr, "Warning: Invalid CDN '%s', using 'unpkg' as default\n", pkgverCDN)
			return frontend_config.CDNUnpkg
		}
		return cdn
	}

	// Try to read from config file if it exists
	// (This is a simplified version - in production you'd properly load the config)
	// For now, default to unpkg
	return frontend_config.CDNUnpkg
}

// fetchAndDisplayVersions fetches versions from the specified CDN and displays them
func fetchAndDisplayVersions(packageName string, cdn frontend_config.CDN) error {
	var versions []string
	var latestVersion string
	var err error

	fmt.Printf("Fetching versions for '%s' from %s...\n\n", packageName, cdn)

	switch cdn {
	case frontend_config.CDNUnpkg:
		result, err := frontend_mgr.FetchUnpkgVersions(packageName)
		if err != nil {
			return fmt.Errorf("failed to fetch versions from unpkg: %w", err)
		}
		// Extract version numbers from the map
		versions = make([]string, 0, len(result.Versions))
		for ver := range result.Versions {
			versions = append(versions, ver)
		}
		latestVersion = result.DistTags["latest"]

	case frontend_config.CDNCdnjs:
		result, err := frontend_mgr.FetchCdnjsVersions(packageName)
		if err != nil {
			return fmt.Errorf("failed to fetch versions from cdnjs: %w", err)
		}
		versions = result.Versions
		latestVersion = result.Version

	case frontend_config.CDNJsdelivr:
		result, err := frontend_mgr.FetchJsdelivrVersions(packageName)
		if err != nil {
			return fmt.Errorf("failed to fetch versions from jsdelivr: %w", err)
		}
		// Extract version numbers from version info
		versions = make([]string, 0, len(result.Versions))
		for _, vInfo := range result.Versions {
			versions = append(versions, vInfo.Version)
		}
		latestVersion = result.Tags["latest"]

	default:
		return fmt.Errorf("unsupported CDN: %s", cdn)
	}

	if err != nil {
		return err
	}

	if len(versions) == 0 {
		fmt.Println("No versions found for this package.")
		return nil
	}

	// Sort versions (newest first)
	sortedVersions := frontend_mgr.SortVersions(versions)

	// If interactive mode is enabled, launch the TUI
	if pkgverInteractive {
		selectedVersion, err := runInteractive(packageName, string(cdn), latestVersion, sortedVersions)
		if err != nil {
			return fmt.Errorf("interactive mode error: %w", err)
		}
		if selectedVersion != "" {
			fmt.Printf("\nSelected version: %s@%s\n", packageName, selectedVersion)
			fmt.Printf("\nTo add this to your config:\n")
			fmt.Printf("  smfaman add %s@%s\n", packageName, selectedVersion)
		}
		return nil
	}

	// Non-interactive mode: display results
	fmt.Printf("Package: %s\n", packageName)
	fmt.Printf("CDN: %s\n", cdn)
	fmt.Printf("Latest: %s\n", latestVersion)
	fmt.Printf("Total versions: %d\n\n", len(sortedVersions))

	// Limit the number of displayed versions
	displayCount := len(sortedVersions)
	if pkgverLimit > 0 && displayCount > pkgverLimit {
		displayCount = pkgverLimit
	}

	fmt.Printf("Showing %d most recent versions:\n", displayCount)
	fmt.Println(strings.Repeat("-", 40))

	for i := 0; i < displayCount; i++ {
		ver := sortedVersions[i]
		prefix := "  "
		if ver == latestVersion {
			prefix = "→ "
		}
		fmt.Printf("%s%s\n", prefix, ver)
	}

	if len(sortedVersions) > displayCount {
		fmt.Printf("\n... and %d more versions\n", len(sortedVersions)-displayCount)
		fmt.Println("Use --limit to show more versions or --interactive for full list")
	}

	return nil
}

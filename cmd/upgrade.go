package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

var (
	upgradeDryRun     bool
	upgradeInteractive bool
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:     "upgrade [package-name[@version]]",
	Aliases: []string{"u"},
	Short:   "Upgrade library version(s) in the Smart Frontend Asset Manager Configuration",
	Long: `Upgrade one or more libraries to newer versions.

This command can be used in three ways:

1. Upgrade a specific library to a specific version:
   smfaman upgrade react@18.3.0

2. Upgrade a specific library to the latest version:
   smfaman upgrade react

3. Upgrade all libraries to their latest versions:
   smfaman upgrade

The command will fetch the latest available versions from the configured CDN
for each library and update the configuration file accordingly.

Use --dry-run to preview changes without modifying the config file.
Use --interactive to select versions interactively.

Examples:
  smfaman upgrade react@18.3.0
  smfaman upgrade react
  smfaman upgrade --dry-run
  smfaman u bootstrap --interactive
  smfaman u`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if len(args) == 0 {
			// Upgrade all libraries
			err = upgradeAllLibraries()
		} else {
			// Upgrade specific library
			err = upgradeSpecificLibrary(args[0])
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().BoolVar(&upgradeDryRun, "dry-run", false, "Show what would be upgraded without making changes")
	upgradeCmd.Flags().BoolVarP(&upgradeInteractive, "interactive", "i", false, "Interactively select version")
}

// upgradeSpecificLibrary upgrades a specific library to a specified or latest version
func upgradeSpecificLibrary(packageSpec string) error {
	// Parse package name and version
	packageName, specifiedVersion := parsePackageSpec(packageSpec)

	// Load existing config
	config, err := loadConfigForUpgrade(FrontendConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if library exists in config
	libConfig, exists := config.Libraries[packageName]
	if !exists {
		return fmt.Errorf("library '%s' not found in config. Use 'smfaman add' to add it first", packageName)
	}

	currentVersion := libConfig.Version

	// Determine CDN to use
	cdn := config.GetLibraryCDN(libConfig)

	var newVersion string

	if upgradeInteractive {
		// Interactive mode
		versions, latestVersion, err := fetchVersionsForUpgrade(packageName, cdn)
		if err != nil {
			return err
		}

		newVersion, err = runInteractive(packageName, string(cdn), latestVersion, versions)
		if err != nil {
			return fmt.Errorf("interactive mode error: %w", err)
		}

		if newVersion == "" {
			fmt.Println("Cancelled.")
			return nil
		}
	} else if specifiedVersion != "" {
		// Validate specified version
		if err := validateVersionForUpgrade(packageName, specifiedVersion, cdn); err != nil {
			return err
		}
		newVersion = specifiedVersion
	} else {
		// Get latest version
		_, latestVersion, err := fetchVersionsForUpgrade(packageName, cdn)
		if err != nil {
			return err
		}
		newVersion = latestVersion
	}

	// Check if already up to date
	if currentVersion == newVersion {
		fmt.Printf("✓ Library '%s' is already at version %s\n", packageName, currentVersion)
		return nil
	}

	// Show upgrade info
	fmt.Printf("\nUpgrading '%s': %s → %s\n", packageName, currentVersion, newVersion)

	if upgradeDryRun {
		fmt.Println("\n[DRY RUN] No changes made to config file.")
		return nil
	}

	// Update version
	libConfig.Version = newVersion
	config.Libraries[packageName] = libConfig

	// Save config
	if err := saveConfigForUpgrade(FrontendConfig, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print success message
	fmt.Printf("\n✓ Library upgraded successfully!\n\n")
	fmt.Printf("Package:  %s\n", packageName)
	fmt.Printf("Old:      %s\n", currentVersion)
	fmt.Printf("New:      %s\n", newVersion)
	fmt.Printf("CDN:      %s\n", cdn)
	fmt.Printf("\nConfig updated: %s\n", FrontendConfig)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  • Sync libraries: smfaman sync\n")

	return nil
}

// upgradeAllLibraries upgrades all libraries to their latest versions
func upgradeAllLibraries() error {
	// Load existing config
	config, err := loadConfigForUpgrade(FrontendConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(config.Libraries) == 0 {
		fmt.Println("No libraries found in config.")
		return nil
	}

	fmt.Printf("Checking for updates for %d library(ies)...\n\n", len(config.Libraries))

	type upgradeInfo struct {
		name           string
		currentVersion string
		newVersion     string
		cdn            frontend_config.CDN
	}

	var upgrades []upgradeInfo
	var upToDate []string
	var errors []string

	// Check each library for updates
	for libName, libConfig := range config.Libraries {
		currentVersion := libConfig.Version
		cdn := config.GetLibraryCDN(libConfig)

		// Fetch latest version
		_, latestVersion, err := fetchVersionsForUpgrade(libName, cdn)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", libName, err))
			continue
		}

		if currentVersion == latestVersion {
			upToDate = append(upToDate, fmt.Sprintf("%s@%s", libName, currentVersion))
		} else {
			upgrades = append(upgrades, upgradeInfo{
				name:           libName,
				currentVersion: currentVersion,
				newVersion:     latestVersion,
				cdn:            cdn,
			})
		}
	}

	// Display summary
	if len(upgrades) == 0 {
		fmt.Println("✓ All libraries are up to date!")
		if len(upToDate) > 0 {
			fmt.Println("\nCurrent versions:")
			for _, lib := range upToDate {
				fmt.Printf("  • %s\n", lib)
			}
		}
		return nil
	}

	fmt.Printf("Found %d upgrade(s) available:\n\n", len(upgrades))
	for _, u := range upgrades {
		fmt.Printf("  • %s: %s → %s (from %s)\n", u.name, u.currentVersion, u.newVersion, u.cdn)
	}

	if len(upToDate) > 0 {
		fmt.Printf("\nAlready up to date (%d):\n", len(upToDate))
		for _, lib := range upToDate {
			fmt.Printf("  • %s\n", lib)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(errors))
		for _, errMsg := range errors {
			fmt.Printf("  • %s\n", errMsg)
		}
	}

	if upgradeDryRun {
		fmt.Println("\n[DRY RUN] No changes made to config file.")
		return nil
	}

	// Apply upgrades
	fmt.Println("\nApplying upgrades...")
	for _, u := range upgrades {
		libConfig := config.Libraries[u.name]
		libConfig.Version = u.newVersion
		config.Libraries[u.name] = libConfig
	}

	// Save config
	if err := saveConfigForUpgrade(FrontendConfig, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n✓ Successfully upgraded %d library(ies)!\n", len(upgrades))
	fmt.Printf("\nConfig updated: %s\n", FrontendConfig)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  • Sync libraries: smfaman sync\n")

	return nil
}

// fetchVersionsForUpgrade fetches versions from the appropriate CDN
func fetchVersionsForUpgrade(packageName string, cdn frontend_config.CDN) (versions []string, latest string, err error) {
	switch cdn {
	case frontend_config.CDNUnpkg:
		result, err := frontend_mgr.FetchUnpkgVersions(packageName)
		if err != nil {
			return nil, "", fmt.Errorf("failed to fetch versions from unpkg: %w", err)
		}
		versions = make([]string, 0, len(result.Versions))
		for ver := range result.Versions {
			versions = append(versions, ver)
		}
		latest = result.DistTags["latest"]

	case frontend_config.CDNCdnjs:
		result, err := frontend_mgr.FetchCdnjsVersions(packageName)
		if err != nil {
			return nil, "", fmt.Errorf("failed to fetch versions from cdnjs: %w", err)
		}
		versions = result.Versions
		latest = result.Version

	case frontend_config.CDNJsdelivr:
		result, err := frontend_mgr.FetchJsdelivrVersions(packageName)
		if err != nil {
			return nil, "", fmt.Errorf("failed to fetch versions from jsdelivr: %w", err)
		}
		versions = make([]string, 0, len(result.Versions))
		for _, vInfo := range result.Versions {
			versions = append(versions, vInfo.Version)
		}
		latest = result.Tags["latest"]

	default:
		return nil, "", fmt.Errorf("unsupported CDN: %s", cdn)
	}

	if len(versions) == 0 {
		return nil, "", fmt.Errorf("no versions found for package '%s'", packageName)
	}

	// Sort versions
	sortedVersions := frontend_mgr.SortVersions(versions)
	return sortedVersions, latest, nil
}

// validateVersionForUpgrade checks if a version exists for a package on a CDN
func validateVersionForUpgrade(packageName, version string, cdn frontend_config.CDN) error {
	versions, _, err := fetchVersionsForUpgrade(packageName, cdn)
	if err != nil {
		return err
	}

	// Check if version exists
	for _, v := range versions {
		if v == version {
			return nil
		}
	}

	return fmt.Errorf("version '%s' not found for package '%s' on %s", version, packageName, cdn)
}

// loadConfigForUpgrade loads a frontend config from a file
func loadConfigForUpgrade(path string) (*frontend_config.FrontendConfig, error) {
	return loadConfig(path)
}

// saveConfigForUpgrade saves a frontend config to a file
func saveConfigForUpgrade(path string, config *frontend_config.FrontendConfig) error {
	return saveConfig(path, config)
}

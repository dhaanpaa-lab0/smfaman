package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

var (
	addCDN         string
	addInteractive bool
	addForce       bool
	addFiles       []string
	addOutputPath  string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add <package-name>[@version]",
	Aliases: []string{"pkgadd", "a"},
	Short:   "Add new library to the Smart Frontend Asset Manager Configuration",
	Long: `Add a new library to your frontend configuration file.

You can specify the version directly using package@version syntax, or use
the --interactive flag to browse and select from available versions.

The package name is required in all cases. If no version is specified and
interactive mode is not enabled, the latest version will be used.

The library will be added to the config file specified by the -f flag
(default: smartfrontend.yaml). If the library already exists, use --force
to overwrite its configuration.

You can optionally specify:
  - CDN to use with --cdn flag
  - Specific files to download with --files flag
  - Custom output path with --output flag

Examples:
  smfaman add react@18.2.0
  smfaman add react --interactive
  smfaman add bootstrap --cdn cdnjs
  smfaman add jquery@3.7.1 --files "dist/jquery.min.js"
  smfaman add lodash --output "./custom/lodash"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageSpec := args[0]

		if err := addLibraryToConfig(packageSpec); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVar(&addCDN, "cdn", "", "CDN to use for this library (unpkg, cdnjs, jsdelivr)")
	addCmd.Flags().BoolVarP(&addInteractive, "interactive", "i", false, "Interactively select version")
	addCmd.Flags().BoolVar(&addForce, "force", false, "Overwrite library if it already exists in config")
	addCmd.Flags().StringArrayVar(&addFiles, "files", nil, "Specific files to download (can be specified multiple times)")
	addCmd.Flags().StringVar(&addOutputPath, "output", "", "Custom output path for this library")
}

// addLibraryToConfig adds a library to the frontend config
func addLibraryToConfig(packageSpec string) error {
	// Parse package name and version
	packageName, specifiedVersion := parsePackageSpec(packageSpec)

	// Load existing config
	config, err := loadConfig(FrontendConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if library already exists
	if _, exists := config.Libraries[packageName]; exists && !addForce {
		return fmt.Errorf("library '%s' already exists in config, use --force to overwrite", packageName)
	}

	// Determine CDN to use
	cdn := determineCDNForAdd(config)

	var selectedVersion string

	// If interactive mode, launch version selector
	if addInteractive {
		versions, latestVersion, err := fetchVersionsForCDN(packageName, cdn)
		if err != nil {
			return err
		}

		selectedVersion, err = runInteractive(packageName, string(cdn), latestVersion, versions)
		if err != nil {
			return fmt.Errorf("interactive mode error: %w", err)
		}

		if selectedVersion == "" {
			fmt.Println("Cancelled.")
			return nil
		}
	} else if specifiedVersion != "" {
		// Validate specified version
		selectedVersion = specifiedVersion
		if err := validateVersion(packageName, selectedVersion, cdn); err != nil {
			return err
		}
	} else {
		// No version specified and not interactive - use latest
		_, latestVersion, err := fetchVersionsForCDN(packageName, cdn)
		if err != nil {
			return err
		}
		selectedVersion = latestVersion
		fmt.Printf("No version specified, using latest: %s\n", latestVersion)
	}

	// Create library config
	libConfig := frontend_config.LibraryConfig{
		Version: selectedVersion,
	}

	// Add optional fields if specified
	if addCDN != "" {
		libConfig.CDN = frontend_config.CDN(addCDN)
	}

	if len(addFiles) > 0 {
		libConfig.Files = addFiles
	}

	if addOutputPath != "" {
		libConfig.OutputPath = addOutputPath
	}

	// Add to config
	config.Libraries[packageName] = libConfig

	// Save config
	if err := saveConfig(FrontendConfig, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print success message
	fmt.Printf("\n✓ Library added successfully!\n\n")
	fmt.Printf("Package:  %s@%s\n", packageName, selectedVersion)
	fmt.Printf("CDN:      %s\n", cdn)
	if libConfig.OutputPath != "" {
		fmt.Printf("Output:   %s\n", libConfig.OutputPath)
	}
	if len(libConfig.Files) > 0 {
		fmt.Printf("Files:    %v\n", libConfig.Files)
	}
	fmt.Printf("\nConfig updated: %s\n", FrontendConfig)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  • Sync libraries: smfaman sync\n")

	return nil
}

// parsePackageSpec splits package@version into name and version
// Handles both regular packages (react@18.2.0) and scoped packages (@babel/core@7.22.0)
func parsePackageSpec(spec string) (name, version string) {
	// If it starts with @, it's a scoped package
	if strings.HasPrefix(spec, "@") {
		// Find the second @ (version separator)
		idx := strings.Index(spec[1:], "@")
		if idx == -1 {
			// No version specified for scoped package
			return spec, ""
		}
		// idx is relative to spec[1:], so actual index is idx+1
		return spec[:idx+1], spec[idx+2:]
	}

	// Regular package
	parts := strings.SplitN(spec, "@", 2)
	name = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	return
}

// determineCDNForAdd determines which CDN to use for adding a library
func determineCDNForAdd(config *frontend_config.FrontendConfig) frontend_config.CDN {
	// Priority: --cdn flag > config default > unpkg
	if addCDN != "" {
		cdn := frontend_config.CDN(addCDN)
		if !frontend_config.IsValidCDN(cdn) {
			fmt.Fprintf(os.Stderr, "Warning: Invalid CDN '%s', using config default or 'unpkg'\n", addCDN)
		} else {
			return cdn
		}
	}

	if config.CDN != "" && frontend_config.IsValidCDN(config.CDN) {
		return config.CDN
	}

	return frontend_config.CDNUnpkg
}

// fetchVersionsForCDN fetches versions from the appropriate CDN
func fetchVersionsForCDN(packageName string, cdn frontend_config.CDN) (versions []string, latest string, err error) {
	fmt.Printf("Fetching versions for '%s' from %s...\n", packageName, cdn)

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

// validateVersion checks if a version exists for a package on a CDN
func validateVersion(packageName, version string, cdn frontend_config.CDN) error {
	versions, _, err := fetchVersionsForCDN(packageName, cdn)
	if err != nil {
		return err
	}

	// Check if version exists
	for _, v := range versions {
		if v == version {
			fmt.Printf("✓ Version %s found for %s\n", version, packageName)
			return nil
		}
	}

	return fmt.Errorf("version '%s' not found for package '%s' on %s", version, packageName, cdn)
}

// loadConfig loads a frontend config from a file
func loadConfig(path string) (*frontend_config.FrontendConfig, error) {
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

// saveConfig saves a frontend config to a file
func saveConfig(path string, config *frontend_config.FrontendConfig) error {
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

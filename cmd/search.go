package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

var (
	searchInteractive bool
	searchLimit       int
	searchCDN         string
	searchJSON        bool
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Aliases: []string{"srch", "find", "s"},
	Short:   "Search for frontend packages across CDNs",
	Long: `Search for frontend packages across multiple CDNs (CDNJS, UNPKG, jsDelivr).

Examples:
  # Search for packages (CLI mode - outputs table)
  smfaman search react

  # Search with interactive TUI
  smfaman search react --interactive
  smfaman search react -i

  # Interactive mode with query prompt
  smfaman search --interactive
  smfaman search -i

  # Search specific CDN
  smfaman search jquery --cdn cdnjs
  smfaman search vue --cdn npm

  # Limit number of results
  smfaman search bootstrap --limit 10

  # Output as JSON (for automation)
  smfaman search lodash --json

Supported CDN values:
  all      - Search all CDNs (default)
  cdnjs    - Search only CDNJS
  npm      - Search npm registry (for UNPKG and jsDelivr)`,
	Args: cobra.MaximumNArgs(1),
	Run:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVarP(&searchInteractive, "interactive", "i", false, "Interactive mode with TUI")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 20, "Maximum number of results to return")
	searchCmd.Flags().StringVarP(&searchCDN, "cdn", "c", "all", "Which CDN to search (all, cdnjs, npm)")
	searchCmd.Flags().BoolVarP(&searchJSON, "json", "j", false, "Output results as JSON")
}

func runSearch(cmd *cobra.Command, args []string) {
	var query string
	if len(args) > 0 {
		query = args[0]
	}

	if searchInteractive {
		// Run interactive TUI
		runSearchTUI(query)
		return
	}

	// CLI mode requires a query
	if query == "" {
		fmt.Println("Error: query argument is required in CLI mode")
		fmt.Println("Use --interactive or -i flag to search interactively without providing a query")
		return
	}

	// Run CLI mode
	results, err := performSearch(query, searchCDN, searchLimit)
	if err != nil {
		fmt.Printf("Error searching for packages: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Printf("No packages found matching '%s'\n", query)
		return
	}

	if searchJSON {
		// Output as JSON for automation
		outputJSON(results)
	} else {
		// Output as formatted table
		outputTable(results)
	}
}

// performSearch executes the search based on CDN selection
func performSearch(query, cdn string, limit int) ([]frontend_mgr.SearchResult, error) {
	switch strings.ToLower(cdn) {
	case "cdnjs":
		return frontend_mgr.SearchCdnjs(query, limit)
	case "npm":
		return frontend_mgr.SearchNpm(query, limit)
	case "all":
		return frontend_mgr.SearchAllCDNs(query, limit)
	default:
		return nil, fmt.Errorf("unsupported CDN: %s (supported: all, cdnjs, npm)", cdn)
	}
}

// outputJSON outputs results as JSON
func outputJSON(results []frontend_mgr.SearchResult) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// outputTable outputs results as a formatted table
func outputTable(results []frontend_mgr.SearchResult) {
	// Calculate column widths
	maxName := len("PACKAGE")
	maxVersion := len("VERSION")
	maxCDN := len("CDN")
	maxDesc := len("DESCRIPTION")

	for _, r := range results {
		if len(r.Name) > maxName {
			maxName = len(r.Name)
		}
		if len(r.Version) > maxVersion {
			maxVersion = len(r.Version)
		}
		if len(r.CDN) > maxCDN {
			maxCDN = len(r.CDN)
		}
		descLen := len(r.Description)
		if descLen > 80 {
			descLen = 80 // Truncate long descriptions
		}
		if descLen > maxDesc {
			maxDesc = descLen
		}
	}

	// Limit column widths to reasonable maximums
	if maxName > 40 {
		maxName = 40
	}
	if maxVersion > 15 {
		maxVersion = 15
	}
	if maxCDN > 20 {
		maxCDN = 20
	}
	if maxDesc > 80 {
		maxDesc = 80
	}

	// Print header
	headerFormat := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds\n", maxName, maxVersion, maxCDN, maxDesc)
	fmt.Printf(headerFormat, "PACKAGE", "VERSION", "CDN", "DESCRIPTION")

	// Print separator
	separator := strings.Repeat("─", maxName) + "  " +
		strings.Repeat("─", maxVersion) + "  " +
		strings.Repeat("─", maxCDN) + "  " +
		strings.Repeat("─", maxDesc)
	fmt.Println(separator)

	// Print rows
	rowFormat := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds\n", maxName, maxVersion, maxCDN, maxDesc)
	for _, r := range results {
		name := truncate(r.Name, maxName)
		version := truncate(r.Version, maxVersion)
		cdn := truncate(r.CDN, maxCDN)
		desc := truncate(r.Description, maxDesc)
		fmt.Printf(rowFormat, name, version, cdn, desc)
	}

	fmt.Printf("\nFound %d package(s)\n", len(results))
}

// truncate truncates a string to the specified length, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

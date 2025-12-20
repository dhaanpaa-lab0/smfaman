package cmd

import (
	"github.com/spf13/cobra"
)

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap new projects from various frameworks",
	Long: `Bootstrap new projects by downloading and setting up starter kits
from various frontend frameworks and libraries.

This command provides subcommands for different frameworks, each of which
will download the appropriate starter kit, extract it, and set up a new
project in your current directory.

Available frameworks:
  xmlui - Bootstrap a new XMLUI project (declarative XML-based UI framework)
  htmx  - Bootstrap a new HTMX project (hypermedia-driven web applications)

Example:
  smfaman bootstrap xmlui
  smfaman bootstrap xmlui --directory my-xmlui-app
  smfaman bootstrap htmx
  smfaman bootstrap htmx --directory my-htmx-app`,
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	installForce bool
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install smfaman binary to user's bin directory and update PATH",
	Long: `Install the smfaman binary to the user's bin directory and ensure it's in PATH.

This command will:
  • Create ~/bin directory if it doesn't exist (%USERPROFILE%\bin on Windows)
  • Copy the smfaman binary to ~/bin
  • Check if ~/bin is in your PATH
  • Add ~/bin to PATH if needed (persistent across sessions)
  • Modify the appropriate shell configuration file based on your shell

Supported shells:
  • bash (Linux/Mac)
  • zsh (Mac/Linux)
  • fish (Linux/Mac)
  • PowerShell (Windows)
  • Windows Command Prompt

The PATH modification is persistent and will survive system restarts.
After installation, you may need to restart your terminal or run the
suggested command to reload your shell configuration.

Examples:
  smfaman install              # Install to ~/bin
  smfaman install --force      # Overwrite if already installed`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVar(&installForce, "force", false, "Overwrite existing installation")
}

func runInstall() error {
	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Determine bin directory based on OS
	binDir := filepath.Join(homeDir, "bin")

	fmt.Printf("Installing smfaman...\n\n")
	fmt.Printf("Source:      %s\n", exePath)
	fmt.Printf("Destination: %s\n", binDir)
	fmt.Printf("OS:          %s\n", runtime.GOOS)
	fmt.Printf("Arch:        %s\n\n", runtime.GOARCH)

	// Create bin directory if it doesn't exist
	if err := createBinDirectory(binDir); err != nil {
		return err
	}

	// Copy binary to bin directory
	destPath := filepath.Join(binDir, getBinaryName())
	if err := copyBinary(exePath, destPath); err != nil {
		return err
	}

	// Check and update PATH
	if err := ensureInPath(binDir); err != nil {
		return err
	}

	fmt.Printf("\n✓ Installation complete!\n\n")

	// Show reload instructions
	showReloadInstructions()

	return nil
}

func createBinDirectory(binDir string) error {
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		fmt.Printf("Creating directory: %s\n", binDir)
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return fmt.Errorf("failed to create bin directory: %w", err)
		}
		fmt.Printf("✓ Directory created\n\n")
	} else {
		fmt.Printf("✓ Directory already exists\n\n")
	}
	return nil
}

func copyBinary(src, dest string) error {
	// Check if destination already exists
	if _, err := os.Stat(dest); err == nil {
		if !installForce {
			return fmt.Errorf("binary already exists at %s (use --force to overwrite)", dest)
		}
		fmt.Printf("Overwriting existing binary...\n")
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source binary: %w", err)
	}

	// Write to destination with executable permissions
	if err := os.WriteFile(dest, data, 0755); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}

	fmt.Printf("✓ Binary copied to %s\n", dest)
	return nil
}

func ensureInPath(binDir string) error {
	// Check if already in PATH
	if isInPath(binDir) {
		fmt.Printf("✓ %s is already in PATH\n", binDir)
		return nil
	}

	fmt.Printf("\n%s is not in PATH. Adding it now...\n", binDir)

	// Add to PATH based on OS
	if runtime.GOOS == "windows" {
		return addToPathWindows(binDir)
	}
	return addToPathUnix(binDir)
}

func isInPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)

	// Normalize the directory path
	dir = filepath.Clean(dir)

	for _, p := range strings.Split(pathEnv, pathSep) {
		if filepath.Clean(p) == dir {
			return true
		}
	}
	return false
}

func addToPathWindows(binDir string) error {
	// Use PowerShell to modify user PATH environment variable permanently
	// This is more reliable than using setx which has length limitations

	psScript := fmt.Sprintf(`
$path = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($path -notlike '*%s*') {
    [Environment]::SetEnvironmentVariable('Path', $path + ';%s', 'User')
    Write-Host 'PATH updated successfully'
} else {
    Write-Host 'Already in PATH'
}
`, binDir, binDir)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update PATH (try running as administrator): %w\nOutput: %s", err, output)
	}

	fmt.Printf("✓ Added to Windows user PATH\n")
	fmt.Printf("  Note: Current terminal session still has old PATH\n")

	return nil
}

func addToPathUnix(binDir string) error {
	// Detect shell
	shell := detectShell()

	// Determine config file to modify
	configFile, exportLine := getShellConfig(shell, binDir)

	if configFile == "" {
		fmt.Printf("⚠ Could not determine shell config file for shell: %s\n", shell)
		fmt.Printf("  Please manually add this line to your shell's config file:\n")
		fmt.Printf("  export PATH=\"$HOME/bin:$PATH\"\n")
		return nil
	}

	// Check if already in config file
	if isInConfigFile(configFile, binDir) {
		fmt.Printf("✓ PATH already configured in %s\n", configFile)
		return nil
	}

	// Add to config file
	if err := addToConfigFile(configFile, exportLine); err != nil {
		return err
	}

	fmt.Printf("✓ Added PATH to %s\n", configFile)

	return nil
}

func detectShell() string {
	// Try to get shell from environment
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}

	// Extract shell name from path (e.g., /bin/bash -> bash)
	return filepath.Base(shell)
}

func getShellConfig(shell, binDir string) (configFile, exportLine string) {
	homeDir, _ := os.UserHomeDir()

	exportLine = fmt.Sprintf("\n# Added by smfaman\nexport PATH=\"%s:$PATH\"\n", binDir)

	switch shell {
	case "bash":
		// On macOS, use .bash_profile; on Linux, use .bashrc
		if runtime.GOOS == "darwin" {
			configFile = filepath.Join(homeDir, ".bash_profile")
			// Also check .profile as fallback
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				profilePath := filepath.Join(homeDir, ".profile")
				if _, err := os.Stat(profilePath); err == nil {
					configFile = profilePath
				}
			}
		} else {
			configFile = filepath.Join(homeDir, ".bashrc")
		}

	case "zsh":
		configFile = filepath.Join(homeDir, ".zshrc")

	case "fish":
		configFile = filepath.Join(homeDir, ".config", "fish", "config.fish")
		exportLine = fmt.Sprintf("\n# Added by smfaman\nset -gx PATH %s $PATH\n", binDir)

	case "ksh":
		configFile = filepath.Join(homeDir, ".kshrc")

	case "tcsh", "csh":
		configFile = filepath.Join(homeDir, ".cshrc")
		exportLine = fmt.Sprintf("\n# Added by smfaman\nsetenv PATH %s:$PATH\n", binDir)

	default:
		// Try .profile as fallback
		configFile = filepath.Join(homeDir, ".profile")
	}

	return configFile, exportLine
}

func isInConfigFile(configFile, binDir string) bool {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return false
	}

	content := string(data)
	// Check for various PATH export patterns
	patterns := []string{
		binDir,
		"$HOME/bin",
		"${HOME}/bin",
		"~/bin",
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) && strings.Contains(content, "PATH") {
			return true
		}
	}

	return false
}

func addToConfigFile(configFile, exportLine string) error {
	// Create directory if it doesn't exist (for fish config)
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open file in append mode, create if doesn't exist
	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(exportLine); err != nil {
		return fmt.Errorf("failed to write to config file: %w", err)
	}

	return nil
}

func getBinaryName() string {
	name := "smfaman"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func showReloadInstructions() {
	shell := detectShell()
	homeDir, _ := os.UserHomeDir()

	fmt.Println("To start using smfaman, either:")
	fmt.Println("  1. Restart your terminal, or")

	if runtime.GOOS == "windows" {
		fmt.Println("  2. Open a new PowerShell/Command Prompt window")
	} else {
		configFile, _ := getShellConfig(shell, filepath.Join(homeDir, "bin"))
		switch shell {
		case "bash":
			fmt.Printf("  2. Run: source %s\n", configFile)
		case "zsh":
			fmt.Println("  2. Run: source ~/.zshrc")
		case "fish":
			fmt.Println("  2. Run: source ~/.config/fish/config.fish")
		default:
			fmt.Printf("  2. Reload your shell configuration\n")
		}
	}

	fmt.Println("\nThen verify installation with:")
	fmt.Println("  smfaman --version")
}

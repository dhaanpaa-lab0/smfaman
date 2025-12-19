package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

var (
	syncForce  bool
	syncDryRun bool
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Download libraries that are defined in the configuration file but not present locally",
	Long: `Synchronize your local frontend assets with the configuration file.

This command reads your frontend configuration file and downloads any libraries
that are defined but not yet present locally. It will:
  - Check which libraries are defined in the configuration
  - Verify which ones are missing or outdated locally
  - Download missing assets from the specified CDN (jsDelivr, UNPKG, or CDNJS)
  - Show progress bars for each download

This is useful after cloning a project or updating the configuration file.

Flags:
  --force: Re-download all files even if they exist locally
  --dry-run: Show what would be downloaded without actually downloading

Example:
  smfaman sync
  smfaman sync -f myproject.yaml
  smfaman sync --force
  smfaman sync --dry-run`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSync(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "Re-download all files even if they exist")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Show what would be downloaded without downloading")
}

// DownloadTask represents a file to download
type DownloadTask struct {
	LibraryName string
	Version     string
	CDN         frontend_config.CDN
	FilePath    string // Path on CDN
	DestPath    string // Local destination path
	URL         string
	Size        int64
}

// runSync executes the sync command
func runSync() error {
	// Load config
	config, err := loadConfig(FrontendConfig)
	if err != nil {
		return err
	}

	if len(config.Libraries) == 0 {
		fmt.Println("No libraries defined in configuration.")
		return nil
	}

	// Build download tasks
	tasks, err := buildDownloadTasks(config)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("✓ All libraries are up to date!")
		return nil
	}

	// Show summary
	fmt.Printf("\nLibraries to sync: %d\n", len(config.Libraries))
	fmt.Printf("Files to download: %d\n\n", len(tasks))

	if syncDryRun {
		fmt.Println("Dry run - would download:")
		for _, task := range tasks {
			fmt.Printf("  • %s@%s: %s → %s\n", task.LibraryName, task.Version, task.FilePath, task.DestPath)
		}
		return nil
	}

	// Run interactive download with progress
	return runInteractiveDownload(tasks)
}

// buildDownloadTasks creates a list of files to download
func buildDownloadTasks(config *frontend_config.FrontendConfig) ([]DownloadTask, error) {
	var tasks []DownloadTask

	for libName, libConfig := range config.Libraries {
		// Determine CDN
		cdn := config.GetLibraryCDN(libConfig)
		if cdn == "" {
			cdn = frontend_config.CDNUnpkg
		}

		// Get destination path
		destPath, err := config.GetLibraryDestination(libName, libConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get destination for %s: %w", libName, err)
		}

		// Fetch file list from CDN (uses caching)
		files, err := fetchFileList(libName, libConfig.Version, cdn)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch files for %s: %w", libName, err)
		}

		// Filter files if specific files are configured
		if len(libConfig.Files) > 0 {
			files = filterFiles(files, libConfig.Files)
		}

		// Create download tasks
		for _, file := range files {
			localPath := filepath.Join(destPath, file.Path)

			// Skip if file exists and not forcing
			if !syncForce {
				if _, err := os.Stat(localPath); err == nil {
					continue
				}
			}

			task := DownloadTask{
				LibraryName: libName,
				Version:     libConfig.Version,
				CDN:         cdn,
				FilePath:    file.Path,
				DestPath:    localPath,
				URL:         file.URL,
				Size:        file.Size,
			}
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// CDNFile represents a file available on a CDN
type CDNFile struct {
	Path string
	URL  string
	Size int64
}

// fetchFileList fetches the list of files for a library from the CDN
func fetchFileList(libName, version string, cdn frontend_config.CDN) ([]CDNFile, error) {
	var files []CDNFile

	switch cdn {
	case frontend_config.CDNUnpkg:
		meta, err := frontend_mgr.FetchUnpkgMeta(libName, version)
		if err != nil {
			return nil, err
		}
		for _, file := range meta.Files {
			if file.Type == "file" {
				files = append(files, CDNFile{
					Path: strings.TrimPrefix(file.Path, "/"),
					URL:  fmt.Sprintf("https://unpkg.com/%s@%s%s", libName, version, file.Path),
					Size: int64(file.Size),
				})
			}
		}

	case frontend_config.CDNCdnjs:
		resp, err := frontend_mgr.FetchCdnjsVersion(libName, version)
		if err != nil {
			return nil, err
		}
		for _, file := range resp.Files {
			files = append(files, CDNFile{
				Path: file,
				URL:  fmt.Sprintf("https://cdnjs.cloudflare.com/ajax/libs/%s/%s/%s", libName, version, file),
				Size: 0, // CDNJS doesn't provide size in metadata
			})
		}

	case frontend_config.CDNJsdelivr:
		resp, err := frontend_mgr.FetchJsdelivrPackage(libName, version)
		if err != nil {
			return nil, err
		}
		// Recursively collect files from jsDelivr tree
		files = collectJsdelivrFiles(libName, version, resp.Files, "")

	default:
		return nil, fmt.Errorf("unsupported CDN: %s", cdn)
	}

	return files, nil
}

// collectJsdelivrFiles recursively collects files from jsDelivr file tree
func collectJsdelivrFiles(libName, version string, jsFiles []frontend_mgr.JsdelivrFile, basePath string) []CDNFile {
	var files []CDNFile

	for _, f := range jsFiles {
		path := filepath.Join(basePath, f.Name)

		if f.Type == "file" {
			files = append(files, CDNFile{
				Path: path,
				URL:  fmt.Sprintf("https://cdn.jsdelivr.net/npm/%s@%s/%s", libName, version, path),
				Size: int64(f.Size),
			})
		} else if f.Type == "directory" && len(f.Files) > 0 {
			// Recursively collect files from subdirectories
			files = append(files, collectJsdelivrFiles(libName, version, f.Files, path)...)
		}
	}

	return files
}

// filterFiles filters file list based on configured files
func filterFiles(files []CDNFile, patterns []string) []CDNFile {
	var filtered []CDNFile

	for _, file := range files {
		for _, pattern := range patterns {
			// Simple pattern matching (exact or prefix match)
			if file.Path == pattern || strings.HasPrefix(file.Path, pattern) {
				filtered = append(filtered, file)
				break
			}
		}
	}

	return filtered
}

// downloadFile downloads a file from URL to destination
func downloadFile(url, destPath string) error {
	// Create destination directory
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Create file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// runInteractiveDownload runs the download with progress UI
func runInteractiveDownload(tasks []DownloadTask) error {
	m := newSyncModel(tasks)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running interactive download: %w", err)
	}

	if sm, ok := finalModel.(syncModel); ok {
		if sm.err != nil {
			return sm.err
		}

		fmt.Printf("\n✓ Sync complete!\n")
		fmt.Printf("Downloaded %d files\n", len(tasks))
	}

	return nil
}

// Messages for the sync model
type downloadStartMsg struct{ task DownloadTask }
type downloadProgressMsg struct{ percent float64 }
type downloadCompleteMsg struct{ task DownloadTask }
type downloadErrorMsg struct{ err error }
type allCompleteMsg struct{}
type tickMsg time.Time

// syncModel is the Bubble Tea model for sync progress
type syncModel struct {
	tasks        []DownloadTask
	currentIndex int
	currentTask  *DownloadTask
	progress     float64
	completed    int
	err          error
	downloading  bool
	startTime    time.Time
}

func newSyncModel(tasks []DownloadTask) syncModel {
	return syncModel{
		tasks:        tasks,
		currentIndex: 0,
		completed:    0,
	}
}

func (m syncModel) Init() tea.Cmd {
	if len(m.tasks) > 0 {
		return m.startDownload()
	}
	return nil
}

func (m syncModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case downloadStartMsg:
		m.currentTask = &msg.task
		m.progress = 0
		m.downloading = true
		m.startTime = time.Now()
		return m, tea.Batch(
			m.downloadNext(),
			tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)

	case tickMsg:
		if m.downloading && m.currentTask != nil {
			// Simulate progress based on elapsed time
			elapsed := time.Since(m.startTime)
			// Progress increases over 200ms
			m.progress = min(elapsed.Seconds()/0.2, 0.99)
			return m, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		}
		return m, nil

	case downloadProgressMsg:
		m.progress = msg.percent
		return m, nil

	case downloadCompleteMsg:
		m.downloading = false
		m.progress = 1.0
		m.completed++
		m.currentIndex++

		if m.currentIndex >= len(m.tasks) {
			return m, func() tea.Msg { return allCompleteMsg{} }
		}

		return m, m.startDownload()

	case downloadErrorMsg:
		m.err = msg.err
		return m, tea.Quit

	case allCompleteMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m syncModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.currentTask == nil {
		return "Preparing to download...\n"
	}

	var s strings.Builder

	// Overall progress
	s.WriteString(fmt.Sprintf("Syncing libraries... [%d/%d files]\n\n", m.completed, len(m.tasks)))

	// Current file
	s.WriteString(fmt.Sprintf("Library: %s@%s\n", m.currentTask.LibraryName, m.currentTask.Version))
	s.WriteString(fmt.Sprintf("File:    %s\n", m.currentTask.FilePath))

	// Progress bar
	barWidth := 40
	filled := int(m.progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	s.WriteString(fmt.Sprintf("\n[%s] %.1f%%\n", bar, m.progress*100))

	return s.String()
}

func (m syncModel) downloadNext() tea.Cmd {
	return func() tea.Msg {
		if m.currentIndex >= len(m.tasks) {
			return allCompleteMsg{}
		}

		task := m.tasks[m.currentIndex]

		// Download the file (this happens in background)
		err := downloadFile(task.URL, task.DestPath)
		if err != nil {
			return downloadErrorMsg{err: fmt.Errorf("failed to download %s: %w", task.FilePath, err)}
		}

		// Add a small delay to ensure progress is visible
		time.Sleep(100 * time.Millisecond)

		return downloadCompleteMsg{task: task}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (m syncModel) startDownload() tea.Cmd {
	return func() tea.Msg {
		if m.currentIndex >= len(m.tasks) {
			return allCompleteMsg{}
		}
		task := m.tasks[m.currentIndex]
		return downloadStartMsg{task: task}
	}
}

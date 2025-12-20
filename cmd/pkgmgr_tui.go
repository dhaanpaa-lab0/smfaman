package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

// View modes
const (
	viewLibraryList = iota
	viewEditLibrary
	viewAddLibrary
	viewEditGlobal
	viewVersionSelection
)

// Edit fields for library
const (
	editFieldVersion = iota
	editFieldCDN
	editFieldFiles
	editFieldOutputPath
	editFieldCount
)

// Global edit fields
const (
	globalFieldProjectName = iota
	globalFieldDestination
	globalFieldCDN
	globalFieldCount
)

var (
	pkgmgrTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginLeft(2)

	pkgmgrItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	pkgmgrSelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170"))

	pkgmgrHelpStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			PaddingBottom(1).
			Foreground(lipgloss.Color("240"))

	pkgmgrHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				MarginBottom(1)

	pkgmgrLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))

	pkgmgrValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86"))
)

// Message types for async operations
type versionsFetchedMsg struct {
	versions []string
	latest   string
	err      error
}

type versionSelectedMsg struct {
	version string
}

type libraryItem struct {
	name    string
	version string
	cdn     frontend_config.CDN
}

func (i libraryItem) FilterValue() string { return i.name }

type libraryItemDelegate struct{}

func (d libraryItemDelegate) Height() int                             { return 1 }
func (d libraryItemDelegate) Spacing() int                            { return 0 }
func (d libraryItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d libraryItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(libraryItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s@%s", i.name, i.version)
	if i.cdn != "" {
		str = fmt.Sprintf("%s (%s)", str, i.cdn)
	}

	fn := pkgmgrItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return pkgmgrSelectedItemStyle.Render("→ " + strings.Join(s, " "))
		}
	} else {
		fn = func(s ...string) string {
			return pkgmgrItemStyle.Render("  " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type pkgmgrModel struct {
	config          *frontend_config.FrontendConfig
	configPath      string
	list            list.Model
	view            int
	editInputs      []textinput.Model
	focusIndex      int
	cdnChoice       int
	cdnOptions      []string
	editingLib      string
	err             error
	saved           bool
	successMsg      string
	quitting        bool
	versionSelector *pkgverModel
	fetchingVersions bool
	versionError    string
}

func newPkgmgrModel(config *frontend_config.FrontendConfig, configPath string) pkgmgrModel {
	items := make([]list.Item, 0, len(config.Libraries))
	for name, libConfig := range config.Libraries {
		items = append(items, libraryItem{
			name:    name,
			version: libConfig.Version,
			cdn:     libConfig.CDN,
		})
	}

	const defaultWidth = 80
	const defaultHeight = 20

	l := list.New(items, libraryItemDelegate{}, defaultWidth, defaultHeight)
	l.Title = "Frontend Package Manager"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = pkgmgrTitleStyle
	l.Styles.HelpStyle = pkgmgrHelpStyle

	// Set custom keybindings
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "edit"),
			),
			key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add"),
			),
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete"),
			),
			key.NewBinding(
				key.WithKeys("g"),
				key.WithHelp("g", "global settings"),
			),
			key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "save & quit"),
			),
		}
	}

	m := pkgmgrModel{
		config:     config,
		configPath: configPath,
		list:       l,
		view:       viewLibraryList,
		cdnOptions: []string{"", "unpkg", "cdnjs", "jsdelivr"},
	}

	return m
}

func (m pkgmgrModel) Init() tea.Cmd {
	return nil
}

// fetchVersionsCmd fetches versions for a package asynchronously
func fetchVersionsCmd(packageName string, cdn frontend_config.CDN) tea.Cmd {
	return func() tea.Msg {
		var versions []string
		var latest string
		var err error

		switch cdn {
		case frontend_config.CDNUnpkg:
			result, fetchErr := frontend_mgr.FetchUnpkgVersions(packageName)
			if fetchErr != nil {
				err = fetchErr
			} else {
				versions = make([]string, 0, len(result.Versions))
				for ver := range result.Versions {
					versions = append(versions, ver)
				}
				latest = result.DistTags["latest"]
			}

		case frontend_config.CDNCdnjs:
			result, fetchErr := frontend_mgr.FetchCdnjsVersions(packageName)
			if fetchErr != nil {
				err = fetchErr
			} else {
				versions = result.Versions
				latest = result.Version
			}

		case frontend_config.CDNJsdelivr:
			result, fetchErr := frontend_mgr.FetchJsdelivrVersions(packageName)
			if fetchErr != nil {
				err = fetchErr
			} else {
				versions = make([]string, 0, len(result.Versions))
				for _, vInfo := range result.Versions {
					versions = append(versions, vInfo.Version)
				}
				latest = result.Tags["latest"]
			}

		default:
			err = fmt.Errorf("unsupported CDN: %s", cdn)
		}

		if err == nil && len(versions) > 0 {
			versions = frontend_mgr.SortVersions(versions)
		}

		return versionsFetchedMsg{
			versions: versions,
			latest:   latest,
			err:      err,
		}
	}
}

func (m pkgmgrModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case versionsFetchedMsg:
		m.fetchingVersions = false
		if msg.err != nil {
			m.versionError = fmt.Sprintf("Error fetching versions: %v", msg.err)
			m.view = viewAddLibrary
			return m, nil
		}
		// Get package name from first input
		packageName := m.editInputs[0].Value()
		cdn := m.cdnOptions[m.cdnChoice]
		if cdn == "" {
			cdn = string(m.config.CDN)
		}
		if cdn == "" {
			cdn = "unpkg"
		}
		selector := newPkgverModel(packageName, cdn, msg.latest, msg.versions)
		m.versionSelector = &selector
		m.view = viewVersionSelection
		return m, nil

	case versionSelectedMsg:
		// Set the selected version in the version input field
		if msg.version != "" {
			m.editInputs[1].SetValue(msg.version)
		}
		m.view = viewAddLibrary
		m.versionSelector = nil
		m.focusIndex = 1 // Focus back on version field
		return m, textinput.Blink

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		if m.versionSelector != nil {
			m.versionSelector.list.SetWidth(msg.Width)
			m.versionSelector.list.SetHeight(msg.Height - 4)
		}
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		switch m.view {
		case viewLibraryList:
			return m.updateLibraryList(msg)
		case viewEditLibrary:
			return m.updateEditLibrary(msg)
		case viewAddLibrary:
			return m.updateAddLibrary(msg)
		case viewEditGlobal:
			return m.updateEditGlobal(msg)
		case viewVersionSelection:
			return m.updateVersionSelection(msg)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pkgmgrModel) updateLibraryList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit

	case "s":
		// Save and quit
		m.saved = true
		m.quitting = true
		return m, tea.Quit

	case "enter":
		// Edit selected library
		if item, ok := m.list.SelectedItem().(libraryItem); ok {
			m.editingLib = item.name
			m.view = viewEditLibrary
			m.focusIndex = 0
			m.initEditLibraryInputs(item)
			return m, textinput.Blink
		}

	case "a":
		// Add new library
		m.view = viewAddLibrary
		m.focusIndex = 0
		m.initAddLibraryInputs()
		return m, textinput.Blink

	case "d":
		// Delete selected library
		if item, ok := m.list.SelectedItem().(libraryItem); ok {
			delete(m.config.Libraries, item.name)
			m.refreshList()
		}

	case "g":
		// Edit global settings
		m.view = viewEditGlobal
		m.focusIndex = 0
		m.initEditGlobalInputs()
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *pkgmgrModel) initEditLibraryInputs(item libraryItem) {
	m.editInputs = make([]textinput.Model, editFieldCount)

	libConfig := m.config.Libraries[item.name]

	// Version
	t := textinput.New()
	t.Placeholder = "Version"
	t.SetValue(libConfig.Version)
	t.Focus()
	t.CharLimit = 50
	t.Width = 50
	t.Prompt = "> "
	t.PromptStyle = focusedStyle
	t.TextStyle = focusedStyle
	m.editInputs[editFieldVersion] = t

	// CDN (managed separately with arrow keys)
	t = textinput.New()
	t.Placeholder = "Use arrow keys to select"
	t.Blur()
	t.CharLimit = 20
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[editFieldCDN] = t

	// Set CDN choice
	m.cdnChoice = 0 // default (empty)
	for i, opt := range m.cdnOptions {
		if frontend_config.CDN(opt) == libConfig.CDN {
			m.cdnChoice = i
			break
		}
	}

	// Files (comma-separated)
	t = textinput.New()
	t.Placeholder = "Files (comma-separated, empty for all)"
	t.SetValue(strings.Join(libConfig.Files, ", "))
	t.Blur()
	t.CharLimit = 200
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[editFieldFiles] = t

	// Output Path
	t = textinput.New()
	t.Placeholder = "Output path (empty for default)"
	t.SetValue(libConfig.OutputPath)
	t.Blur()
	t.CharLimit = 200
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[editFieldOutputPath] = t
}

func (m *pkgmgrModel) initAddLibraryInputs() {
	m.editInputs = make([]textinput.Model, editFieldCount+1) // +1 for name field

	// Name
	t := textinput.New()
	t.Placeholder = "Package name (e.g., react, @babel/core)"
	t.Focus()
	t.CharLimit = 100
	t.Width = 50
	t.Prompt = "> "
	t.PromptStyle = focusedStyle
	t.TextStyle = focusedStyle
	m.editInputs[0] = t

	// Version
	t = textinput.New()
	t.Placeholder = "Version"
	t.Blur()
	t.CharLimit = 50
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[1] = t

	// CDN
	t = textinput.New()
	t.Placeholder = "Use arrow keys to select"
	t.Blur()
	t.CharLimit = 20
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[2] = t
	m.cdnChoice = 0

	// Files
	t = textinput.New()
	t.Placeholder = "Files (comma-separated, empty for all)"
	t.Blur()
	t.CharLimit = 200
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[3] = t

	// Output Path
	t = textinput.New()
	t.Placeholder = "Output path (empty for default)"
	t.Blur()
	t.CharLimit = 200
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[4] = t
}

func (m *pkgmgrModel) initEditGlobalInputs() {
	m.editInputs = make([]textinput.Model, globalFieldCount)

	// Project Name
	t := textinput.New()
	t.Placeholder = "Project name"
	t.SetValue(m.config.ProjectName)
	t.Focus()
	t.CharLimit = 100
	t.Width = 50
	t.Prompt = "> "
	t.PromptStyle = focusedStyle
	t.TextStyle = focusedStyle
	m.editInputs[globalFieldProjectName] = t

	// Destination
	t = textinput.New()
	t.Placeholder = "./frontend/{library_name}"
	t.SetValue(m.config.Destination)
	t.Blur()
	t.CharLimit = 200
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[globalFieldDestination] = t

	// CDN
	t = textinput.New()
	t.Placeholder = "Use arrow keys to select"
	t.Blur()
	t.CharLimit = 20
	t.Width = 50
	t.Prompt = "> "
	m.editInputs[globalFieldCDN] = t

	// Set CDN choice
	m.cdnChoice = 0
	for i, opt := range m.cdnOptions {
		if frontend_config.CDN(opt) == m.config.CDN {
			m.cdnChoice = i
			break
		}
	}
}

func (m pkgmgrModel) updateEditLibrary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewLibraryList
		return m, nil

	case "tab", "shift+tab", "enter", "up", "down":
		s := msg.String()

		// Handle save
		if s == "enter" && m.focusIndex == editFieldCount {
			m.saveLibraryEdit()
			m.view = viewLibraryList
			m.refreshList()
			return m, nil
		}

		// Handle CDN selection
		if m.focusIndex == editFieldCDN {
			if s == "up" {
				m.cdnChoice--
				if m.cdnChoice < 0 {
					m.cdnChoice = len(m.cdnOptions) - 1
				}
				return m, nil
			} else if s == "down" {
				m.cdnChoice++
				if m.cdnChoice >= len(m.cdnOptions) {
					m.cdnChoice = 0
				}
				return m, nil
			}
		}

		// Navigate fields
		if s == "up" || s == "shift+tab" {
			m.focusIndex--
		} else {
			m.focusIndex++
		}

		if m.focusIndex > editFieldCount {
			m.focusIndex = 0
		} else if m.focusIndex < 0 {
			m.focusIndex = editFieldCount
		}

		cmds := make([]tea.Cmd, len(m.editInputs))
		for i := 0; i < len(m.editInputs); i++ {
			if i == m.focusIndex {
				cmds[i] = m.editInputs[i].Focus()
				m.editInputs[i].PromptStyle = focusedStyle
				m.editInputs[i].TextStyle = focusedStyle
			} else {
				m.editInputs[i].Blur()
				m.editInputs[i].PromptStyle = blurredStyle
				m.editInputs[i].TextStyle = noStyle
			}
		}

		return m, tea.Batch(cmds...)
	}

	cmd := m.updateEditInputs(msg)
	return m, cmd
}

func (m pkgmgrModel) updateAddLibrary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewLibraryList
		m.versionError = ""
		return m, nil

	case "v", "i":
		// Trigger interactive version selection when on version field
		if m.focusIndex == 1 {
			packageName := m.editInputs[0].Value()
			if packageName == "" {
				m.versionError = "Please enter a package name first"
				return m, nil
			}

			// Determine CDN
			cdn := m.cdnOptions[m.cdnChoice]
			if cdn == "" {
				cdn = string(m.config.CDN)
			}
			if cdn == "" {
				cdn = "unpkg"
			}

			m.fetchingVersions = true
			m.versionError = ""
			return m, fetchVersionsCmd(packageName, frontend_config.CDN(cdn))
		}

	case "tab", "shift+tab", "enter", "up", "down":
		s := msg.String()

		// Handle save
		if s == "enter" && m.focusIndex == editFieldCount+1 {
			if m.saveNewLibrary() {
				m.view = viewLibraryList
				m.refreshList()
			}
			return m, nil
		}

		// Handle CDN selection (field index 2)
		if m.focusIndex == 2 {
			if s == "up" {
				m.cdnChoice--
				if m.cdnChoice < 0 {
					m.cdnChoice = len(m.cdnOptions) - 1
				}
				return m, nil
			} else if s == "down" {
				m.cdnChoice++
				if m.cdnChoice >= len(m.cdnOptions) {
					m.cdnChoice = 0
				}
				return m, nil
			}
		}

		// Navigate fields
		if s == "up" || s == "shift+tab" {
			m.focusIndex--
		} else {
			m.focusIndex++
		}

		if m.focusIndex > editFieldCount+1 {
			m.focusIndex = 0
		} else if m.focusIndex < 0 {
			m.focusIndex = editFieldCount + 1
		}

		cmds := make([]tea.Cmd, len(m.editInputs))
		for i := 0; i < len(m.editInputs); i++ {
			if i == m.focusIndex {
				cmds[i] = m.editInputs[i].Focus()
				m.editInputs[i].PromptStyle = focusedStyle
				m.editInputs[i].TextStyle = focusedStyle
			} else {
				m.editInputs[i].Blur()
				m.editInputs[i].PromptStyle = blurredStyle
				m.editInputs[i].TextStyle = noStyle
			}
		}

		return m, tea.Batch(cmds...)
	}

	cmd := m.updateEditInputs(msg)
	return m, cmd
}

func (m pkgmgrModel) updateVersionSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Cancel version selection
		m.view = viewAddLibrary
		m.versionSelector = nil
		return m, nil

	case "enter":
		// Select version
		if m.versionSelector != nil {
			if item, ok := m.versionSelector.list.SelectedItem().(versionItem); ok {
				return m, func() tea.Msg {
					return versionSelectedMsg{version: item.version}
				}
			}
		}
		return m, nil
	}

	// Update the version selector
	if m.versionSelector != nil {
		var cmd tea.Cmd
		m.versionSelector.list, cmd = m.versionSelector.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m pkgmgrModel) updateEditGlobal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewLibraryList
		return m, nil

	case "tab", "shift+tab", "enter", "up", "down":
		s := msg.String()

		// Handle save
		if s == "enter" && m.focusIndex == globalFieldCount {
			m.saveGlobalEdit()
			m.view = viewLibraryList
			return m, nil
		}

		// Handle CDN selection
		if m.focusIndex == globalFieldCDN {
			if s == "up" {
				m.cdnChoice--
				if m.cdnChoice < 0 {
					m.cdnChoice = len(m.cdnOptions) - 1
				}
				return m, nil
			} else if s == "down" {
				m.cdnChoice++
				if m.cdnChoice >= len(m.cdnOptions) {
					m.cdnChoice = 0
				}
				return m, nil
			}
		}

		// Navigate fields
		if s == "up" || s == "shift+tab" {
			m.focusIndex--
		} else {
			m.focusIndex++
		}

		if m.focusIndex > globalFieldCount {
			m.focusIndex = 0
		} else if m.focusIndex < 0 {
			m.focusIndex = globalFieldCount
		}

		cmds := make([]tea.Cmd, len(m.editInputs))
		for i := 0; i < len(m.editInputs); i++ {
			if i == m.focusIndex {
				cmds[i] = m.editInputs[i].Focus()
				m.editInputs[i].PromptStyle = focusedStyle
				m.editInputs[i].TextStyle = focusedStyle
			} else {
				m.editInputs[i].Blur()
				m.editInputs[i].PromptStyle = blurredStyle
				m.editInputs[i].TextStyle = noStyle
			}
		}

		return m, tea.Batch(cmds...)
	}

	cmd := m.updateEditInputs(msg)
	return m, cmd
}

func (m *pkgmgrModel) updateEditInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.editInputs))

	for i := range m.editInputs {
		if i == m.focusIndex && (m.view == viewEditLibrary && i != editFieldCDN ||
			m.view == viewAddLibrary && i != 2 ||
			m.view == viewEditGlobal && i != globalFieldCDN) {
			m.editInputs[i], cmds[i] = m.editInputs[i].Update(msg)
		}
	}

	return tea.Batch(cmds...)
}

func (m *pkgmgrModel) saveLibraryEdit() {
	libConfig := m.config.Libraries[m.editingLib]

	libConfig.Version = m.editInputs[editFieldVersion].Value()

	cdnStr := m.cdnOptions[m.cdnChoice]
	libConfig.CDN = frontend_config.CDN(cdnStr)

	filesStr := m.editInputs[editFieldFiles].Value()
	if filesStr != "" {
		files := strings.Split(filesStr, ",")
		for i, f := range files {
			files[i] = strings.TrimSpace(f)
		}
		libConfig.Files = files
	} else {
		libConfig.Files = nil
	}

	libConfig.OutputPath = m.editInputs[editFieldOutputPath].Value()

	m.config.Libraries[m.editingLib] = libConfig
}

func (m *pkgmgrModel) saveNewLibrary() bool {
	name := m.editInputs[0].Value()
	if name == "" {
		return false
	}

	version := m.editInputs[1].Value()
	if version == "" {
		return false
	}

	libConfig := frontend_config.LibraryConfig{
		Version: version,
	}

	cdnStr := m.cdnOptions[m.cdnChoice]
	if cdnStr != "" {
		libConfig.CDN = frontend_config.CDN(cdnStr)
	}

	filesStr := m.editInputs[3].Value()
	if filesStr != "" {
		files := strings.Split(filesStr, ",")
		for i, f := range files {
			files[i] = strings.TrimSpace(f)
		}
		libConfig.Files = files
	}

	libConfig.OutputPath = m.editInputs[4].Value()

	m.config.Libraries[name] = libConfig
	return true
}

func (m *pkgmgrModel) saveGlobalEdit() {
	m.config.ProjectName = m.editInputs[globalFieldProjectName].Value()
	m.config.Destination = m.editInputs[globalFieldDestination].Value()

	cdnStr := m.cdnOptions[m.cdnChoice]
	m.config.CDN = frontend_config.CDN(cdnStr)
}

func (m *pkgmgrModel) refreshList() {
	items := make([]list.Item, 0, len(m.config.Libraries))
	for name, libConfig := range m.config.Libraries {
		items = append(items, libraryItem{
			name:    name,
			version: libConfig.Version,
			cdn:     libConfig.CDN,
		})
	}
	m.list.SetItems(items)
}

func (m pkgmgrModel) View() string {
	if m.quitting {
		if m.saved {
			return successStyle.Render("✓ Config saved successfully!\n")
		}
		return ""
	}

	switch m.view {
	case viewLibraryList:
		return m.viewLibraryListRender()
	case viewEditLibrary:
		return m.viewEditLibraryRender()
	case viewAddLibrary:
		return m.viewAddLibraryRender()
	case viewEditGlobal:
		return m.viewEditGlobalRender()
	case viewVersionSelection:
		return m.viewVersionSelectionRender()
	}

	return ""
}

func (m pkgmgrModel) viewLibraryListRender() string {
	return "\n" + m.list.View()
}

func (m pkgmgrModel) viewEditLibraryRender() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(pkgmgrHeaderStyle.Render(fmt.Sprintf("Edit Library: %s", m.editingLib)) + "\n\n")

	// Version
	if m.focusIndex == editFieldVersion {
		b.WriteString(focusedStyle.Render("Version:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Version:") + "\n")
	}
	b.WriteString(m.editInputs[editFieldVersion].View() + "\n\n")

	// CDN
	if m.focusIndex == editFieldCDN {
		b.WriteString(focusedStyle.Render("CDN (empty for default):") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("CDN (empty for default):") + "\n")
	}
	for i, option := range m.cdnOptions {
		cursor := " "
		displayOpt := option
		if displayOpt == "" {
			displayOpt = "(use default)"
		}
		if i == m.cdnChoice {
			cursor = "●"
			if m.focusIndex == editFieldCDN {
				b.WriteString(focusedStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s\n", cursor, displayOpt))
			}
		} else {
			if m.focusIndex == editFieldCDN {
				b.WriteString(blurredStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			} else {
				b.WriteString(helpStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			}
		}
	}
	b.WriteString("\n")

	// Files
	if m.focusIndex == editFieldFiles {
		b.WriteString(focusedStyle.Render("Files:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Files:") + "\n")
	}
	b.WriteString(m.editInputs[editFieldFiles].View() + "\n\n")

	// Output Path
	if m.focusIndex == editFieldOutputPath {
		b.WriteString(focusedStyle.Render("Output Path:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Output Path:") + "\n")
	}
	b.WriteString(m.editInputs[editFieldOutputPath].View() + "\n\n")

	// Save button
	button := blurredButton
	if m.focusIndex == editFieldCount {
		button = focusedButton
	}
	b.WriteString(button + "\n\n")

	b.WriteString(helpStyle.Render("tab/shift+tab: navigate • up/down: select CDN • enter: save • esc: cancel"))

	return b.String()
}

func (m pkgmgrModel) viewAddLibraryRender() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(pkgmgrHeaderStyle.Render("Add New Library") + "\n\n")

	// Name
	if m.focusIndex == 0 {
		b.WriteString(focusedStyle.Render("Package Name:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Package Name:") + "\n")
	}
	b.WriteString(m.editInputs[0].View() + "\n\n")

	// Version
	if m.focusIndex == 1 {
		b.WriteString(focusedStyle.Render("Version:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Version:") + "\n")
	}
	b.WriteString(m.editInputs[1].View() + "\n")
	if m.focusIndex == 1 {
		b.WriteString(helpStyle.Render("  Press 'v' or 'i' for interactive version selection") + "\n")
	}
	if m.fetchingVersions {
		b.WriteString(helpStyle.Render("  Fetching versions...") + "\n")
	}
	if m.versionError != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("  "+m.versionError) + "\n")
	}
	b.WriteString("\n")

	// CDN
	if m.focusIndex == 2 {
		b.WriteString(focusedStyle.Render("CDN (empty for default):") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("CDN (empty for default):") + "\n")
	}
	for i, option := range m.cdnOptions {
		cursor := " "
		displayOpt := option
		if displayOpt == "" {
			displayOpt = "(use default)"
		}
		if i == m.cdnChoice {
			cursor = "●"
			if m.focusIndex == 2 {
				b.WriteString(focusedStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s\n", cursor, displayOpt))
			}
		} else {
			if m.focusIndex == 2 {
				b.WriteString(blurredStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			} else {
				b.WriteString(helpStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			}
		}
	}
	b.WriteString("\n")

	// Files
	if m.focusIndex == 3 {
		b.WriteString(focusedStyle.Render("Files:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Files:") + "\n")
	}
	b.WriteString(m.editInputs[3].View() + "\n\n")

	// Output Path
	if m.focusIndex == 4 {
		b.WriteString(focusedStyle.Render("Output Path:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Output Path:") + "\n")
	}
	b.WriteString(m.editInputs[4].View() + "\n\n")

	// Add button
	button := blurredButton
	if m.focusIndex == editFieldCount+1 {
		button = focusedButton
	}
	b.WriteString(button + "\n\n")

	b.WriteString(helpStyle.Render("tab/shift+tab: navigate • up/down: select CDN • enter: add • esc: cancel"))

	return b.String()
}

func (m pkgmgrModel) viewEditGlobalRender() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(pkgmgrHeaderStyle.Render("Global Settings") + "\n\n")

	// Project Name
	if m.focusIndex == globalFieldProjectName {
		b.WriteString(focusedStyle.Render("Project Name:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Project Name:") + "\n")
	}
	b.WriteString(m.editInputs[globalFieldProjectName].View() + "\n\n")

	// Destination
	if m.focusIndex == globalFieldDestination {
		b.WriteString(focusedStyle.Render("Destination Path:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Destination Path:") + "\n")
	}
	b.WriteString(m.editInputs[globalFieldDestination].View() + "\n")
	b.WriteString(helpStyle.Render("  Use {library_name} as placeholder") + "\n\n")

	// CDN
	if m.focusIndex == globalFieldCDN {
		b.WriteString(focusedStyle.Render("Default CDN:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Default CDN:") + "\n")
	}
	for i, option := range m.cdnOptions {
		cursor := " "
		displayOpt := option
		if displayOpt == "" {
			displayOpt = "(none)"
		}
		if i == m.cdnChoice {
			cursor = "●"
			if m.focusIndex == globalFieldCDN {
				b.WriteString(focusedStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s\n", cursor, displayOpt))
			}
		} else {
			if m.focusIndex == globalFieldCDN {
				b.WriteString(blurredStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			} else {
				b.WriteString(helpStyle.Render(fmt.Sprintf("  %s %s\n", cursor, displayOpt)))
			}
		}
	}
	b.WriteString("\n")

	// Save button
	button := blurredButton
	if m.focusIndex == globalFieldCount {
		button = focusedButton
	}
	b.WriteString(button + "\n\n")

	b.WriteString(helpStyle.Render("tab/shift+tab: navigate • up/down: select CDN • enter: save • esc: cancel"))

	return b.String()
}

func (m pkgmgrModel) viewVersionSelectionRender() string {
	if m.versionSelector != nil {
		return "\n" + m.versionSelector.list.View()
	}
	return "\nLoading versions..."
}

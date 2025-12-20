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
	"nexus-sds.com/smfaman/pkgs/frontend_mgr"
)

// Styles
var (
	searchTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginLeft(2).
				MarginBottom(1)

	searchTableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				PaddingLeft(2)

	searchItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	searchSelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color("170")).
				Bold(true)

	searchPaginationStyle = list.DefaultStyles().
				PaginationStyle.
				PaddingLeft(4)

	searchHelpStyle = list.DefaultStyles().
			HelpStyle.
			PaddingLeft(4).
			PaddingBottom(1)

	searchQuitTextStyle = lipgloss.NewStyle().
				Margin(1, 0, 2, 4)

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				Width(15)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(1, 2).
			MarginTop(1).
			MarginLeft(2)

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				MarginLeft(2).
				MarginBottom(1)
)

// View states
type viewState int

const (
	viewQueryInput viewState = iota
	viewSearchResults
	viewPackageDetail
	viewLoading
)

// Messages
type searchCompletedMsg struct {
	results []frontend_mgr.SearchResult
	err     error
}

// Search result item for the list
type searchResultItem struct {
	result frontend_mgr.SearchResult
	index  int
}

func (i searchResultItem) FilterValue() string {
	keywords := strings.Join(i.result.Keywords, " ")
	return i.result.Name + " " + i.result.Description + " " + keywords
}

// Custom delegate for table-style rendering
type searchResultDelegate struct {
	maxNameWidth    int
	maxVersionWidth int
	maxCDNWidth     int
}

func (d searchResultDelegate) Height() int                             { return 1 }
func (d searchResultDelegate) Spacing() int                            { return 0 }
func (d searchResultDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d searchResultDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(searchResultItem)
	if !ok {
		return
	}

	// Format columns with fixed widths
	name := padRight(truncate(i.result.Name, d.maxNameWidth), d.maxNameWidth)
	version := padRight(truncate(i.result.Version, d.maxVersionWidth), d.maxVersionWidth)
	cdn := padRight(truncate(i.result.CDN, d.maxCDNWidth), d.maxCDNWidth)
	desc := truncate(i.result.Description, 50)

	line := fmt.Sprintf("%s  %s  %s  %s", name, version, cdn, desc)

	if index == m.Index() {
		fmt.Fprint(w, searchSelectedItemStyle.Render("→ "+line))
	} else {
		fmt.Fprint(w, searchItemStyle.Render("  "+line))
	}
}

// Main TUI model
type searchTUIModel struct {
	state         viewState
	queryInput    textinput.Model
	list          list.Model
	delegate      searchResultDelegate
	results       []frontend_mgr.SearchResult
	selectedPkg   *frontend_mgr.SearchResult
	query         string
	err           error
	quitting      bool
	width         int
	height        int
}

func newSearchTUIModel(initialQuery string) searchTUIModel {
	ti := textinput.New()
	ti.Placeholder = "Enter package name (e.g., react, vue, bootstrap)..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 60

	// If we have an initial query, start with that
	if initialQuery != "" {
		ti.SetValue(initialQuery)
	}

	return searchTUIModel{
		state:      viewQueryInput,
		queryInput: ti,
	}
}

func (m searchTUIModel) Init() tea.Cmd {
	// If we already have a query, start searching immediately
	if m.queryInput.Value() != "" {
		return tea.Sequence(
			textinput.Blink,
			m.performSearch,
		)
	}
	return textinput.Blink
}

func (m searchTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Only update list size if we're in results view and list is initialized
		if m.state == viewSearchResults && len(m.results) > 0 {
			m.list.SetWidth(msg.Width)
			m.list.SetHeight(msg.Height - 6)
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case viewQueryInput:
			return m.updateQueryInput(msg)
		case viewSearchResults:
			return m.updateSearchResults(msg)
		case viewPackageDetail:
			return m.updatePackageDetail(msg)
		}

	case searchCompletedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.quitting = true
			return m, tea.Quit
		}

		m.results = msg.results
		m.state = viewSearchResults

		// Create list with results
		items := make([]list.Item, len(msg.results))
		for i, r := range msg.results {
			items[i] = searchResultItem{
				result: r,
				index:  i,
			}
		}

		// Calculate column widths
		maxName, maxVersion, maxCDN := 15, 10, 15
		for _, r := range msg.results {
			if len(r.Name) > maxName && maxName < 40 {
				maxName = len(r.Name)
				if maxName > 40 {
					maxName = 40
				}
			}
			if len(r.Version) > maxVersion && maxVersion < 15 {
				maxVersion = len(r.Version)
				if maxVersion > 15 {
					maxVersion = 15
				}
			}
			if len(r.CDN) > maxCDN && maxCDN < 20 {
				maxCDN = len(r.CDN)
				if maxCDN > 20 {
					maxCDN = 20
				}
			}
		}

		m.delegate = searchResultDelegate{
			maxNameWidth:    maxName,
			maxVersionWidth: maxVersion,
			maxCDNWidth:     maxCDN,
		}

		width := m.width
		if width == 0 {
			width = 120
		}
		height := m.height - 6
		if height < 10 {
			height = 20
		}

		l := list.New(items, m.delegate, width, height)
		l.Title = fmt.Sprintf("Search results for '%s' (%d packages found)", m.query, len(msg.results))
		l.SetShowStatusBar(true)
		l.SetFilteringEnabled(true)
		l.Styles.Title = searchTitleStyle
		l.Styles.PaginationStyle = searchPaginationStyle
		l.Styles.HelpStyle = searchHelpStyle

		l.AdditionalShortHelpKeys = func() []key.Binding {
			return []key.Binding{
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "view details"),
				),
				key.NewBinding(
					key.WithKeys("n"),
					key.WithHelp("n", "new search"),
				),
			}
		}

		m.list = l
		return m, nil
	}

	var cmd tea.Cmd
	switch m.state {
	case viewQueryInput:
		m.queryInput, cmd = m.queryInput.Update(msg)
	case viewSearchResults:
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m searchTUIModel) updateQueryInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit

	case "enter":
		query := strings.TrimSpace(m.queryInput.Value())
		if query == "" {
			return m, nil
		}
		m.query = query
		m.state = viewLoading
		return m, m.performSearch
	}

	var cmd tea.Cmd
	m.queryInput, cmd = m.queryInput.Update(msg)
	return m, cmd
}

func (m searchTUIModel) updateSearchResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "q", "esc":
		// Go back to query input
		m.state = viewQueryInput
		m.queryInput.SetValue("")
		m.queryInput.Focus()
		return m, textinput.Blink

	case "n":
		// New search
		m.state = viewQueryInput
		m.queryInput.SetValue("")
		m.queryInput.Focus()
		return m, textinput.Blink

	case "enter":
		// View package details
		i, ok := m.list.SelectedItem().(searchResultItem)
		if ok {
			m.selectedPkg = &i.result
			m.state = viewPackageDetail
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m searchTUIModel) updatePackageDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "q", "esc", "enter":
		// Go back to search results
		m.state = viewSearchResults
		m.selectedPkg = nil
		return m, nil
	}
	return m, nil
}

func (m searchTUIModel) View() string {
	if m.quitting {
		if m.err != nil {
			return searchQuitTextStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
		}
		return searchQuitTextStyle.Render("Cancelled.\n")
	}

	switch m.state {
	case viewQueryInput:
		return m.viewQueryInput()
	case viewSearchResults:
		return m.viewSearchResults()
	case viewPackageDetail:
		return m.viewPackageDetail()
	case viewLoading:
		return m.viewLoading()
	}

	return ""
}

func (m searchTUIModel) viewQueryInput() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(inputPromptStyle.Render("🔍 Search for Frontend Packages"))
	b.WriteString("\n\n")
	b.WriteString(searchItemStyle.Render("  " + m.queryInput.View()))
	b.WriteString("\n\n")
	b.WriteString(searchHelpStyle.Render("  Press Enter to search • Esc to cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m searchTUIModel) viewSearchResults() string {
	var b strings.Builder

	// Add table header
	nameHeader := padRight("PACKAGE", m.delegate.maxNameWidth)
	versionHeader := padRight("VERSION", m.delegate.maxVersionWidth)
	cdnHeader := padRight("CDN", m.delegate.maxCDNWidth)
	descHeader := "DESCRIPTION"

	header := fmt.Sprintf("  %s  %s  %s  %s", nameHeader, versionHeader, cdnHeader, descHeader)

	b.WriteString("\n")
	b.WriteString(m.list.View())

	// Insert header after title
	view := b.String()
	lines := strings.Split(view, "\n")
	if len(lines) > 2 {
		// Insert header after title line
		newLines := []string{lines[0], lines[1]}
		newLines = append(newLines, searchTableHeaderStyle.Render(header))
		newLines = append(newLines, searchTableHeaderStyle.Render("  "+strings.Repeat("─", len(header)-2)))
		newLines = append(newLines, lines[2:]...)
		view = strings.Join(newLines, "\n")
	}

	return view
}

func (m searchTUIModel) viewPackageDetail() string {
	if m.selectedPkg == nil {
		return ""
	}

	pkg := m.selectedPkg
	var b strings.Builder

	b.WriteString("\n\n")
	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("  📦 %s", pkg.Name)))
	b.WriteString("\n")

	// Build detail box content
	var details strings.Builder
	details.WriteString(detailLabelStyle.Render("Version:") + "  " + detailValueStyle.Render(pkg.Version) + "\n")
	details.WriteString(detailLabelStyle.Render("CDN:") + "  " + detailValueStyle.Render(pkg.CDN) + "\n")

	if pkg.Description != "" {
		details.WriteString("\n")
		details.WriteString(detailLabelStyle.Render("Description:") + "\n")
		details.WriteString(detailValueStyle.Render(wordWrap(pkg.Description, 70)) + "\n")
	}

	if pkg.Homepage != "" {
		details.WriteString("\n")
		details.WriteString(detailLabelStyle.Render("Homepage:") + "\n")
		details.WriteString(detailValueStyle.Render(pkg.Homepage) + "\n")
	}

	if len(pkg.Keywords) > 0 {
		details.WriteString("\n")
		details.WriteString(detailLabelStyle.Render("Keywords:") + "\n")
		details.WriteString(detailValueStyle.Render(strings.Join(pkg.Keywords, ", ")) + "\n")
	}

	details.WriteString("\n")
	details.WriteString(detailLabelStyle.Render("Add to config:") + "\n")
	details.WriteString(detailValueStyle.Render(fmt.Sprintf("smfaman add %s@%s", pkg.Name, pkg.Version)) + "\n")

	b.WriteString(detailBoxStyle.Render(details.String()))
	b.WriteString("\n\n")
	b.WriteString(searchHelpStyle.Render("  Press Enter/Esc to go back • Ctrl+C to quit"))
	b.WriteString("\n")

	return b.String()
}

func (m searchTUIModel) viewLoading() string {
	return searchQuitTextStyle.Render(fmt.Sprintf("🔍 Searching for '%s'...\n", m.query))
}

func (m searchTUIModel) performSearch() tea.Msg {
	results, err := performSearch(m.query, searchCDN, searchLimit)
	return searchCompletedMsg{
		results: results,
		err:     err,
	}
}

// runSearchTUI starts the interactive search interface
func runSearchTUI(initialQuery string) {
	m := newSearchTUIModel(initialQuery)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running interactive mode: %v\n", err)
	}
}

// Helper functions
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func wordWrap(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var wrapped strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)
		if i == 0 {
			wrapped.WriteString(word)
			lineLen = wordLen
		} else if lineLen+wordLen+1 <= width {
			wrapped.WriteString(" " + word)
			lineLen += wordLen + 1
		} else {
			wrapped.WriteString("\n" + word)
			lineLen = wordLen
		}
	}

	return wrapped.String()
}

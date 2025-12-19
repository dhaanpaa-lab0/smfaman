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
)

var (
	pkgverTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginLeft(2)

	pkgverItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	pkgverSelectedItemStyle = lipgloss.NewStyle().
					PaddingLeft(2).
					Foreground(lipgloss.Color("170"))

	pkgverLatestItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("42")).
				Bold(true)

	pkgverPaginationStyle = list.DefaultStyles().
				PaginationStyle.
				PaddingLeft(4)

	pkgverHelpStyle = list.DefaultStyles().
			HelpStyle.
			PaddingLeft(4).
			PaddingBottom(1)

	pkgverQuitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type versionItem struct {
	version   string
	isLatest  bool
	index     int
	totalVers int
}

func (i versionItem) FilterValue() string { return i.version }

type versionItemDelegate struct{}

func (d versionItemDelegate) Height() int                             { return 1 }
func (d versionItemDelegate) Spacing() int                            { return 0 }
func (d versionItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d versionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(versionItem)
	if !ok {
		return
	}

	str := i.version

	// Show index number
	prefix := fmt.Sprintf("%3d. ", i.index+1)

	// Add latest marker
	if i.isLatest {
		prefix = "→  "
		str = fmt.Sprintf("%s (latest)", str)
	}

	fn := pkgverItemStyle.Render
	if index == m.Index() {
		if i.isLatest {
			fn = func(s ...string) string {
				return pkgverLatestItemStyle.Render(prefix + strings.Join(s, " "))
			}
		} else {
			fn = func(s ...string) string {
				return pkgverSelectedItemStyle.Render(prefix + strings.Join(s, " "))
			}
		}
	} else {
		fn = func(s ...string) string {
			return pkgverItemStyle.Render(prefix + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type pkgverModel struct {
	list          list.Model
	packageName   string
	cdn           string
	latestVersion string
	totalVersions int
	choice        string
	quitting      bool
	filter        textinput.Model
	filtering     bool
}

func newPkgverModel(packageName, cdn, latestVersion string, versions []string) pkgverModel {
	items := make([]list.Item, len(versions))
	for i, v := range versions {
		items[i] = versionItem{
			version:   v,
			isLatest:  v == latestVersion,
			index:     i,
			totalVers: len(versions),
		}
	}

	const defaultWidth = 80
	const defaultHeight = 20

	l := list.New(items, versionItemDelegate{}, defaultWidth, defaultHeight)
	l.Title = fmt.Sprintf("Select a version for %s (from %s)", packageName, cdn)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = pkgverTitleStyle
	l.Styles.PaginationStyle = pkgverPaginationStyle
	l.Styles.HelpStyle = pkgverHelpStyle

	// Set custom keybindings
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
		}
	}

	return pkgverModel{
		list:          l,
		packageName:   packageName,
		cdn:           cdn,
		latestVersion: latestVersion,
		totalVersions: len(versions),
	}
}

func (m pkgverModel) Init() tea.Cmd {
	return nil
}

func (m pkgverModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(versionItem)
			if ok {
				m.choice = i.version
			}
			return m, tea.Quit

		case "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pkgverModel) View() string {
	if m.choice != "" {
		return pkgverQuitTextStyle.Render(fmt.Sprintf("Selected: %s@%s\n", m.packageName, m.choice))
	}
	if m.quitting {
		return pkgverQuitTextStyle.Render("Cancelled.\n")
	}
	return "\n" + m.list.View()
}

// runInteractive starts the interactive version selector
func runInteractive(packageName, cdn, latestVersion string, versions []string) (string, error) {
	m := newPkgverModel(packageName, cdn, latestVersion, versions)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running interactive mode: %w", err)
	}

	if m, ok := finalModel.(pkgverModel); ok {
		if m.choice != "" {
			return m.choice, nil
		}
	}

	return "", nil
}

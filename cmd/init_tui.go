package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
	"nexus-sds.com/smfaman/pkgs/frontend_config"
)

// Form fields
const (
	fieldProjectName = iota
	fieldDestination
	fieldCDN
	fieldCount
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))
)

type initModel struct {
	inputs      []textinput.Model
	focusIndex  int
	cdnChoice   int
	cdnOptions  []string
	configFile  string
	err         error
	submitted   bool
	successMsg  string
}

type submitCompleteMsg struct {
	err        error
	successMsg string
}

func newInitModel(configFile string) initModel {
	m := initModel{
		inputs:     make([]textinput.Model, fieldCount),
		cdnOptions: []string{"unpkg", "cdnjs", "jsdelivr"},
		configFile: configFile,
	}

	var t textinput.Model

	// Project Name
	t = textinput.New()
	t.Placeholder = "my-project"
	t.Focus()
	t.CharLimit = 100
	t.Width = 50
	t.Prompt = "> "
	t.PromptStyle = focusedStyle
	t.TextStyle = focusedStyle
	m.inputs[fieldProjectName] = t

	// Destination
	t = textinput.New()
	t.Placeholder = "./frontend/{library_name}"
	t.CharLimit = 200
	t.Width = 50
	t.Prompt = "> "
	t.PromptStyle = blurredStyle
	t.TextStyle = noStyle
	m.inputs[fieldDestination] = t

	// CDN (text representation, managed separately)
	t = textinput.New()
	t.Placeholder = "Use arrow keys to select"
	t.CharLimit = 20
	t.Width = 50
	t.Prompt = "> "
	t.PromptStyle = blurredStyle
	t.TextStyle = noStyle
	m.inputs[fieldCDN] = t
	m.inputs[fieldCDN].Blur()

	return m
}

func (m initModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case submitCompleteMsg:
		m.submitted = true
		m.err = msg.err
		m.successMsg = msg.successMsg
		if msg.err == nil {
			// Quit after showing success message
			return m, tea.Quit
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Handle submit
			if s == "enter" && m.focusIndex == fieldCount {
				return m, m.submitForm()
			}

			// Handle CDN selection
			if m.focusIndex == fieldCDN {
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

			// Cycle through inputs
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > fieldCount {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = fieldCount
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = blurredStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *initModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text input the focused field, and skip CDN field (it's managed with arrow keys)
	for i := range m.inputs {
		if i == m.focusIndex && i != fieldCDN {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
	}

	return tea.Batch(cmds...)
}

func (m initModel) View() string {
	var b strings.Builder

	// Show success or error message if submitted
	if m.submitted {
		if m.err != nil {
			b.WriteString(errorStyle.Render("Error: "+m.err.Error()) + "\n")
			return b.String()
		}
		b.WriteString(successStyle.Render(m.successMsg) + "\n")
		return b.String()
	}

	// Title
	b.WriteString(titleStyle.Render("Initialize Smart Frontend Config") + "\n\n")

	// Project Name
	if m.focusIndex == fieldProjectName {
		b.WriteString(focusedStyle.Render("Project Name:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Project Name:") + "\n")
	}
	b.WriteString(m.inputs[fieldProjectName].View() + "\n\n")

	// Destination
	if m.focusIndex == fieldDestination {
		b.WriteString(focusedStyle.Render("Destination Path:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Destination Path:") + "\n")
	}
	b.WriteString(m.inputs[fieldDestination].View() + "\n")
	b.WriteString(helpStyle.Render("  Use {library_name} as placeholder") + "\n\n")

	// CDN Selection
	if m.focusIndex == fieldCDN {
		b.WriteString(focusedStyle.Render("Default CDN:") + "\n")
	} else {
		b.WriteString(blurredStyle.Render("Default CDN:") + "\n")
	}

	for i, option := range m.cdnOptions {
		cursor := " "
		if i == m.cdnChoice {
			cursor = "●"
			if m.focusIndex == fieldCDN {
				b.WriteString(focusedStyle.Render(fmt.Sprintf("  %s %s\n", cursor, option)))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s\n", cursor, option))
			}
		} else {
			if m.focusIndex == fieldCDN {
				b.WriteString(blurredStyle.Render(fmt.Sprintf("  %s %s\n", cursor, option)))
			} else {
				b.WriteString(helpStyle.Render(fmt.Sprintf("  %s %s\n", cursor, option)))
			}
		}
	}
	b.WriteString("\n")

	// Submit button
	button := &blurredButton
	if m.focusIndex == fieldCount {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n%s\n\n", *button)

	// Help
	b.WriteString(helpStyle.Render("tab/shift+tab: navigate • up/down: select CDN • enter: submit • ctrl+c: quit"))

	return b.String()
}

func (m *initModel) submitForm() tea.Cmd {
	configFile := m.configFile
	projectName := m.inputs[fieldProjectName].Value()
	if projectName == "" {
		projectName = m.inputs[fieldProjectName].Placeholder
	}

	destination := m.inputs[fieldDestination].Value()
	if destination == "" {
		destination = m.inputs[fieldDestination].Placeholder
	}

	cdn := frontend_config.CDN(m.cdnOptions[m.cdnChoice])

	return func() tea.Msg {
		// Create config
		config := frontend_config.FrontendConfig{
			ProjectName: projectName,
			Destination: destination,
			CDN:         cdn,
			Libraries:   make(map[string]frontend_config.LibraryConfig),
		}

		// Marshal to YAML
		data, err := yaml.Marshal(&config)
		if err != nil {
			return submitCompleteMsg{
				err: fmt.Errorf("failed to marshal config: %w", err),
			}
		}

		// Write to file
		err = os.WriteFile(configFile, data, 0644)
		if err != nil {
			return submitCompleteMsg{
				err: fmt.Errorf("failed to write config file: %w", err),
			}
		}

		successMsg := fmt.Sprintf("✓ Created %s successfully!\n\nProject: %s\nDestination: %s\nCDN: %s\n\nNext steps:\n  • Add libraries: smfaman add <library>@<version>\n  • Sync libraries: smfaman sync",
			configFile, projectName, destination, cdn)

		return submitCompleteMsg{
			successMsg: successMsg,
		}
	}
}

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62")).Align(lipgloss.Left)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	normalStyle   = lipgloss.NewStyle()
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	diffStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// TUIState represents the current state of the TUI
type TUIState int

const (
	stateSelectGroup TUIState = iota
	stateSelectFirstFile
	stateSelectSecondFile
	stateViewDiff
)

// model represents the TUI model
type model struct {
	groups      [][]string
	currentGroup int
	state       TUIState
	cursor      int
	firstFile   string
	secondFile  string
	diffOutput  string
	diffExec    *DiffExecutor
	width       int
	height      int
}

// initialModel creates a new model with initial state
func initialModel(groups [][]string, diffExec *DiffExecutor) model {
	return model{
		groups:      groups,
		currentGroup: 0,
		state:       stateSelectGroup,
		cursor:      0,
		diffExec:    diffExec,
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// If selecting second file and cursor lands on first file, skip it
				if m.state == stateSelectSecondFile && m.firstFile != "" {
					group := m.getCurrentGroup()
					if m.cursor < len(group) && group[m.cursor] == m.firstFile {
						if m.cursor > 0 {
							m.cursor--
						} else {
							// If at the beginning, jump to the end
							m.cursor = len(group) - 1
							// If the last file is also the first file, go back one more
							if m.cursor > 0 && group[m.cursor] == m.firstFile {
								m.cursor--
							}
						}
					}
				}
			}
			return m, nil

		case "down", "j":
			var max int
			switch m.state {
			case stateSelectGroup:
				max = len(m.groups) - 1
			case stateSelectFirstFile, stateSelectSecondFile:
				max = len(m.getCurrentGroup()) - 1
			case stateViewDiff:
				return m, nil
			}
			if m.cursor < max {
				m.cursor++
				// If selecting second file and cursor lands on first file, skip it
				if m.state == stateSelectSecondFile && m.firstFile != "" {
					group := m.getCurrentGroup()
					if m.cursor < len(group) && group[m.cursor] == m.firstFile {
						if m.cursor < max {
							m.cursor++
						} else {
							// If at the end, jump to the beginning
							m.cursor = 0
							// If the first file is also the selected first file, skip it
							if group[m.cursor] == m.firstFile && m.cursor < max {
								m.cursor++
							}
						}
					}
				}
			}
			return m, nil

		case "enter", " ":
			return m.handleEnter()

		case "esc":
			return m.handleEscape()

		case "n":
			if m.state == stateSelectGroup {
				if m.currentGroup < len(m.groups)-1 {
					m.currentGroup++
					m.cursor = 0
				}
				return m, nil
			}
			return m, nil
		}
	}

	return m, nil
}

// handleEnter handles the enter key press
func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateSelectGroup:
		if len(m.groups) == 0 {
			return m, tea.Quit
		}
		m.state = stateSelectFirstFile
		m.cursor = 0
		return m, nil

	case stateSelectFirstFile:
		group := m.getCurrentGroup()
		if m.cursor < len(group) {
			m.firstFile = group[m.cursor]
			m.state = stateSelectSecondFile
			// Set cursor to first available file (skip the selected first file)
			m.cursor = 0
			if m.cursor < len(group) && group[m.cursor] == m.firstFile {
				if len(group) > 1 {
					m.cursor = 1
				}
			}
		}
		return m, nil

	case stateSelectSecondFile:
		group := m.getCurrentGroup()
		if m.cursor < len(group) {
			selectedFile := group[m.cursor]
			if selectedFile == m.firstFile {
				// Can't select the same file
				return m, nil
			}
			m.secondFile = selectedFile
			// Generate diff
			diff, err := m.diffExec.DiffSideBySide(m.firstFile, m.secondFile)
			if err != nil {
				m.diffOutput = fmt.Sprintf("Error generating diff: %v", err)
			} else {
				m.diffOutput = diff
			}
			m.state = stateViewDiff
		}
		return m, nil

	case stateViewDiff:
		// After viewing diff, go back to selecting first file
		m.state = stateSelectFirstFile
		m.firstFile = ""
		m.secondFile = ""
		m.diffOutput = ""
		m.cursor = 0
		return m, nil
	}

	return m, nil
}

// handleEscape handles the escape key press
func (m model) handleEscape() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateSelectFirstFile:
		// Go back to group selection
		m.state = stateSelectGroup
		m.cursor = 0
		return m, nil

	case stateSelectSecondFile:
		// Go back to first file selection
		m.state = stateSelectFirstFile
		m.firstFile = ""
		m.cursor = 0
		return m, nil

	case stateViewDiff:
		// Go back to second file selection
		m.state = stateSelectSecondFile
		m.secondFile = ""
		m.diffOutput = ""
		m.cursor = 0
		return m, nil
	}

	return m, nil
}

// getCurrentGroup returns the current group of files
func (m model) getCurrentGroup() []string {
	if m.currentGroup >= len(m.groups) {
		return nil
	}
	return m.groups[m.currentGroup]
}

// View renders the UI
func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	switch m.state {
	case stateSelectGroup:
		s.WriteString(m.renderGroupSelection())

	case stateSelectFirstFile:
		s.WriteString(m.renderFileSelection("Select first file:"))

	case stateSelectSecondFile:
		s.WriteString(m.renderFileSelection("Select second file:"))

	case stateViewDiff:
		s.WriteString(m.renderDiff())
	}

	s.WriteString("\n\n")
	s.WriteString(m.renderHelp())

	return s.String()
}

// renderGroupSelection renders the group selection view
func (m model) renderGroupSelection() string {
	var s strings.Builder

	if len(m.groups) == 0 {
		s.WriteString("No groups of similar files found.\n")
		return s.String()
	}

	s.WriteString(titleStyle.Render(fmt.Sprintf("Found %d group(s) of similar files", len(m.groups))))
	s.WriteString("\n\n")

	for i, group := range m.groups {
		style := normalStyle
		if i == m.cursor {
			style = selectedStyle
		}

		// Use fixed-width prefix area (3 chars) for consistent alignment
		prefix := "   "
		if i == m.cursor {
			prefix = ">  "
		}

		// Show group number and file count - apply style only to the text, not the prefix
		groupText := fmt.Sprintf("Group %d: %d files", i+1, len(group))
		s.WriteString(prefix)
		s.WriteString(style.Render(groupText))
		s.WriteString("\n")
		
		// Show the filenames in this group
		var filenames []string
		for _, file := range group {
			filenames = append(filenames, filepath.Base(file))
		}
		fileList := strings.Join(filenames, ", ")
		// Use consistent indentation for file list (4 spaces to align with group text)
		indent := "    "
		s.WriteString(indent)
		s.WriteString(helpStyle.Render(fileList))
		s.WriteString("\n\n")
	}

	return s.String()
}

// renderFileSelection renders the file selection view
func (m model) renderFileSelection(prompt string) string {
	var s strings.Builder

	group := m.getCurrentGroup()
	if len(group) == 0 {
		return "No files in group."
	}

	s.WriteString(titleStyle.Render(fmt.Sprintf("Group %d: %d files\n\n", m.currentGroup+1, len(group))))
	s.WriteString(titleStyle.Render(prompt))
	s.WriteString("\n\n")

	for i, file := range group {
		style := normalStyle
		if i == m.cursor {
			style = selectedStyle
		}

		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		filename := filepath.Base(file)
		// Skip the first file if we're selecting the second file
		if m.state == stateSelectSecondFile && file == m.firstFile {
			// Show it but make it clear it's already selected
			s.WriteString(helpStyle.Render(fmt.Sprintf("%s%s (already selected as first file)", prefix, filename)))
		} else {
			s.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, filename)))
		}
		s.WriteString("\n")
	}

	if m.state == stateSelectSecondFile && m.firstFile != "" {
		s.WriteString("\n")
		s.WriteString(helpStyle.Render(fmt.Sprintf("First file: %s", filepath.Base(m.firstFile))))
	}

	return s.String()
}

// renderDiff renders the diff view
func (m model) renderDiff() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Comparing files:\n\n"))
	s.WriteString(fmt.Sprintf("File 1: %s\n", filepath.Base(m.firstFile)))
	s.WriteString(fmt.Sprintf("File 2: %s\n\n", filepath.Base(m.secondFile)))
	s.WriteString(strings.Repeat("─", m.width))
	s.WriteString("\n\n")

	// Split diff output into lines and display
	diffLines := strings.Split(m.diffOutput, "\n")
	maxLines := m.height - 15 // Leave room for header and help
	if maxLines < 1 {
		maxLines = 10
	}

	if len(diffLines) > maxLines {
		s.WriteString(diffStyle.Render(strings.Join(diffLines[:maxLines], "\n")))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render(fmt.Sprintf("... (%d more lines, scroll to see more)", len(diffLines)-maxLines)))
	} else {
		s.WriteString(diffStyle.Render(m.diffOutput))
	}

	return s.String()
}

// renderHelp renders the help text
func (m model) renderHelp() string {
	var help string
	switch m.state {
	case stateSelectGroup:
		help = "↑/↓: navigate  Enter: select group  n: next group  q: quit"
	case stateSelectFirstFile:
		help = "↑/↓: navigate  Enter: select file  Esc: back  q: quit"
	case stateSelectSecondFile:
		help = "↑/↓: navigate  Enter: select file  Esc: back  q: quit"
	case stateViewDiff:
		help = "Enter: select another pair  Esc: back  q: quit"
	}
	return helpStyle.Render(help)
}

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/worldsayshi/cir/internal/types"
)

// StatusBarModel represents the status bar that shows context files and application state
type StatusBarModel struct {
	workingFiles []types.WorkingFile
	status       string
}

// NewStatusBarModel creates a new status bar model
func NewStatusBarModel() StatusBarModel {
	return StatusBarModel{
		workingFiles: []types.WorkingFile{},
		status:       "Ready",
	}
}

// View renders the status bar
func (m StatusBarModel) View(width int) string {
	// Format working files
	files := make([]string, 0, len(m.workingFiles))
	for _, file := range m.workingFiles {
		files = append(files, file.Path)
	}

	filesText := ""
	if len(files) > 0 {
		filesText = "Files: " + strings.Join(files, ", ")
	} else {
		filesText = "No context files"
	}

	// Truncate if too long
	maxFileTextLen := width - 20
	if len(filesText) > maxFileTextLen && maxFileTextLen > 3 {
		filesText = filesText[:maxFileTextLen-3] + "..."
	}

	// Format status
	statusText := fmt.Sprintf("Status: %s", m.status)

	// Combine and style
	return lipgloss.JoinVertical(
		lipgloss.Left,
		contextBarStyle.Render(filesText),
		statusStyle.Render(statusText),
	)
}

// UpdateWorkingFiles updates the list of working files displayed in the status bar
func (m *StatusBarModel) UpdateWorkingFiles(files []types.WorkingFile) {
	m.workingFiles = files
}

// UpdateStatus sets the status message
func (m *StatusBarModel) UpdateStatus(status string) {
	m.status = status
}

// Styles
var (
	contextBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EEEEEE")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1).
			Width(100)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EEEEEE")).
			Background(lipgloss.Color("#444444")).
			Padding(0, 1).
			Width(100)
)

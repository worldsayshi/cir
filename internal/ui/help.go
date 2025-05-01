package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// helpModel displays keyboard shortcuts and help information
type helpModel struct{}

// newHelpModel creates a new help model
func newHelpModel() helpModel {
	return helpModel{}
}

// View renders the help screen
func (m helpModel) View() string {
	content := `
  CIR KEYBOARD SHORTCUTS

  Normal Mode:
  i : Enter insert mode
  ? : Show/hide this help
  ctrl+e : Open session file
  ctrl+y : Edit context files
  j/k, ↑/↓ : Scroll chat history up/down

  Insert Mode:
  esc : Return to normal mode
  ctrl+s : Submit message and return to normal mode

  Global:
  ctrl+c, q : Quit

  Press any key to close help
`
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Render(content)
}

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

  ? : Show/hide this help
  tab : Toggle between chat history and input
  ctrl+s : Submit message
  ctrl+e : Open session file
  ctrl+y : Edit context files
  ctrl+c, q : Quit

  Chat History Navigation:
  ↑/↓ : Scroll up/down when focused

  Press any key to close help
`
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Render(content)
}

package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/worldsayshi/cir/internal/types"
)

// Styles for different message types
var (
	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7CFC00")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF69B4")).
			Bold(true)

	systemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E90FF")).
			Bold(true)

	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E0E0"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	selectedHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)

	selectedContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700"))
)

// wrapText wraps the text to the specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text // No wrapping needed
	}

	var wrapped strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			wrapped.WriteString("\n")
		}

		if len(line) == 0 {
			continue
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		lineLen := 0
		for j, word := range words {
			wordLen := len([]rune(word))

			if j == 0 {
				wrapped.WriteString(word)
				lineLen = wordLen
			} else if lineLen+wordLen+1 > width {
				wrapped.WriteString("\n" + word)
				lineLen = wordLen
			} else {
				wrapped.WriteString(" " + word)
				lineLen += wordLen + 1
			}
		}
	}

	return wrapped.String()
}

// formatChatHistory converts message list to formatted string with highlighting for selected message
func formatChatHistory(messages []types.Message, width int, selectedIndex int) string {
	if len(messages) == 0 {
		return "No messages yet. Type a message and press Ctrl+S to send."
	}

	// Allow some space for borders and padding
	contentWidth := width - 4
	if contentWidth < 20 {
		contentWidth = 20 // Minimum reasonable width
	}

	var formatted strings.Builder

	for i, msg := range messages {
		// Add separator between messages, but not before the first one
		if i > 0 {
			separator := separatorStyle.Render(strings.Repeat("─", min(50, contentWidth)))
			formatted.WriteString("\n" + separator + "\n\n")
		}

		// Determine if this message is selected
		isSelected := i == selectedIndex

		// Create style based on selection state
		var headerStyle, contentTextStyle lipgloss.Style
		var prefix string

		if isSelected {
			// Highlight the message when selected
			headerStyle = selectedHeaderStyle
			contentTextStyle = selectedContentStyle
			prefix = "▶ " // Add indicator for selected message
		} else {
			// Normal styling when not selected
			prefix = "  " // Space for alignment

			switch msg.Role {
			case "user":
				headerStyle = userStyle
			case "assistant":
				headerStyle = assistantStyle
			case "system":
				headerStyle = systemStyle
			default:
				headerStyle = lipgloss.NewStyle()
			}
			contentTextStyle = contentStyle
		}

		// Format message header based on role
		switch msg.Role {
		case "user":
			formatted.WriteString(prefix + headerStyle.Render("User") + "\n")
			// If we have a question field, use that instead of content
			if msg.Question != "" {
				formatted.WriteString(prefix + contentTextStyle.Render(wrapText(msg.Question, contentWidth-2)) + "\n")
			} else {
				formatted.WriteString(prefix + contentTextStyle.Render(wrapText(msg.Content, contentWidth-2)) + "\n")
			}
		case "assistant":
			formatted.WriteString(prefix + headerStyle.Render("Assistant") + "\n")
			formatted.WriteString(prefix + contentTextStyle.Render(wrapText(msg.Content, contentWidth-2)) + "\n")
		case "system":
			formatted.WriteString(prefix + headerStyle.Render("System") + "\n")
			formatted.WriteString(prefix + contentTextStyle.Render(wrapText(msg.Content, contentWidth-2)) + "\n")
		default:
			formatted.WriteString(prefix + headerStyle.Render(msg.Role) + "\n")
			formatted.WriteString(prefix + contentTextStyle.Render(wrapText(msg.Content, contentWidth-2)) + "\n")
		}
	}

	return formatted.String()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

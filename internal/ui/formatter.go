package ui

import (
	"fmt"
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

// formatChatHistory converts message list to formatted string
func formatChatHistory(messages []types.Message, width int) string {
	if len(messages) == 0 {
		return "No messages yet. Type a message and press Ctrl+S to send."
	}

	// Allow some space for borders and padding
	contentWidth := width - 4
	if contentWidth < 20 {
		contentWidth = 20 // Minimum reasonable width
	}

	var formatted strings.Builder
	separator := separatorStyle.Render(strings.Repeat("─", min(50, contentWidth)))

	for _, msg := range messages {
		// Add separator between messages
		if formatted.Len() > 0 {
			formatted.WriteString("\n" + separator + "\n\n")
		}

		// Format message header based on role
		switch msg.Role {
		case "user":
			formatted.WriteString(userStyle.Render("User") + "\n")
			// If we have a question field, use that instead of content
			if msg.Question != "" {
				formatted.WriteString(contentStyle.Render(wrapText(msg.Question, contentWidth)) + "\n")
			} else {
				formatted.WriteString(contentStyle.Render(wrapText(msg.Content, contentWidth)) + "\n")
			}
		case "assistant":
			formatted.WriteString(assistantStyle.Render("Assistant") + "\n")
			formatted.WriteString(contentStyle.Render(wrapText(msg.Content, contentWidth)) + "\n")
		case "system":
			formatted.WriteString(systemStyle.Render("System") + "\n")
			formatted.WriteString(contentStyle.Render(wrapText(msg.Content, contentWidth)) + "\n")
		default:
			formatted.WriteString(fmt.Sprintf("%s:\n%s\n", msg.Role, wrapText(msg.Content, contentWidth)))
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

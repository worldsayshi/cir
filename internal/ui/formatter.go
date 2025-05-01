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

// formatChatHistory converts message list to formatted string
func formatChatHistory(messages []types.Message) string {
	if len(messages) == 0 {
		return "No messages yet. Type a message and press Ctrl+S to send."
	}

	var formatted strings.Builder
	separator := separatorStyle.Render(strings.Repeat("─", 50))

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
				formatted.WriteString(contentStyle.Render(msg.Question) + "\n")
			} else {
				formatted.WriteString(contentStyle.Render(msg.Content) + "\n")
			}
		case "assistant":
			formatted.WriteString(assistantStyle.Render("Assistant") + "\n")
			formatted.WriteString(contentStyle.Render(msg.Content) + "\n")
		case "system":
			formatted.WriteString(systemStyle.Render("System") + "\n")
			formatted.WriteString(contentStyle.Render(msg.Content) + "\n")
		default:
			formatted.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
	}

	return formatted.String()
}

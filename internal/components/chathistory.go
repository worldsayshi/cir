package components

import (
	"strings"

	"github.com/rivo/tview"
	"github.com/worldsayshi/cir/internal/types"
)

type ChatHistory struct {
	*tview.TextView
}

func NewChatHistory(messages []types.Message) (chatHistory *ChatHistory) {
	chatHistory = &ChatHistory{TextView: tview.NewTextView()}
	chatHistory.
		SetBorder(true).
		SetTitle("History")
	chatHistory.Render(messages)

	return chatHistory
}

func (chatHistory *ChatHistory) Render(messages []types.Message) {
	msgsString := []string{}
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			msgsString = append(msgsString, msg.Question)
		case "assistant":
			msgsString = append(msgsString, msg.Content)
		}
	}
	chatHistory.SetText(strings.Join(msgsString, "\n\n---\n"))
	chatHistory.ScrollToEnd()
}

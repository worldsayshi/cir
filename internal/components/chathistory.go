package components

import (
	"strings"

	"github.com/rivo/tview"
	"github.com/worldsayshi/cir/internal/types"
)

func InitChatHistory(workingSession *types.WorkingSession) *tview.TextView {
	chatHistory := tview.NewTextView()
	chatHistory.
		SetBorder(true).
		SetTitle("History")
	RenderChatHistory(chatHistory, workingSession.Messages)
	return chatHistory
}

func RenderChatHistory(chatHistory *tview.TextView, messages []types.Message) {
	msgsString := []string{}
	for _, msg := range messages {
		if msg.Role == "user" {
			msgsString = append(msgsString, msg.Question)
		} else {
			msgsString = append(msgsString, msg.Content)
		}
	}
	chatHistory.SetText(strings.Join(msgsString, "\n\n---\n"))
	chatHistory.ScrollToEnd()
}

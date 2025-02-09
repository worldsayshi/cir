package components

import (
	"strings"

	"github.com/rivo/tview"
	"github.com/worldsayshi/cir/internal/types"
)

type ChatHistory struct {
	*tview.TextView
}

func NewChatHistory(workingSession *types.WorkingSession) (chatHistory *ChatHistory) {
	chatHistory = &ChatHistory{TextView: tview.NewTextView()}
	chatHistory.
		SetBorder(true).
		SetTitle("History")
	chatHistory.Render(workingSession.Messages)

	return chatHistory
}

// func InitChatHistory(workingSession *types.WorkingSession) *tview.TextView {
// 	chatHistory := tview.NewTextView()
// 	chatHistory.
// 		SetBorder(true).
// 		SetTitle("History")
// 	RenderChatHistory(chatHistory, workingSession.Messages)
// 	return chatHistory
// }

func (chatHistory *ChatHistory) Render(messages []types.Message) {
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

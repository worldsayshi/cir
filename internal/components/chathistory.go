package components

import (
	"log"
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
	chatHistory.SetDynamicColors(true).
		SetRegions(true)
	chatHistory.Render(messages)

	return chatHistory
}

func (chatHistory *ChatHistory) Render(messages []types.Message) {
	width := 20
	log.Println("width", width)
	msgsString := []string{}
	for _, msg := range messages {
		// if len(msgsString) > 0 && (msg.Role == "user" || msg.Role == "assistant") {
		// 	msgsString = append(msgsString, "\n[green]"+strings.Repeat("-", width)+"[white]\n")
		// }
		switch msg.Role {
		case "user":
			msgsString = append(msgsString, "\n[green] "+
				msg.Role+" "+
				strings.Repeat("-", width)+"[white]\n")
			msgsString = append(msgsString, msg.Question)
		case "assistant":
			msgsString = append(msgsString, "\n[purple] "+
				strings.Repeat("-", width)+
				" "+msg.Role+
				"[white]\n")
			msgsString = append(msgsString, "[fuchsia]"+msg.Content+"[white]")
			// default:
			// 	msgsString = append(msgsString, msg.Content)
		}
	}
	chatHistory.SetText(strings.Join(msgsString, "\n"))
	chatHistory.ScrollToEnd()
}

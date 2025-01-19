package main

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	newPrimitive := func(text string) tview.Primitive {
		p := tview.NewTextView()
		p.
			SetBorder(true).
			SetTitle(text)
		return p
	}

	messages := []string{}

	chatHistory := newPrimitive("History").(*tview.TextView)
	contextBar := newPrimitive("Context")
	textInputArea := tview.NewTextArea().
		SetPlaceholder("Write here")
	textInputArea.
		SetBorder(true).
		SetTitle("Input")

	// textInputArea.SetDoneFunc(func(key tcell.Key) {
	// 	if key == tcell.KeyEnter {
	// 		text := textInputArea.GetText()
	// 		if text != "" {
	// 			messages = append(messages, text)
	// 			chatHistory.SetText(strings.Join(messages, "\n"))
	// 			textInputArea.SetText("", true)
	// 		}
	// 	}
	// })

	textInputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			text := textInputArea.GetText()
			if text != "" {
				messages = append(messages, text)
				chatHistory.SetText(strings.Join(messages, "\n\n---\n"))
				textInputArea.SetText("", true)
			}
			return nil
		}
		return event
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatHistory, 0, 5, false).
		AddItem(contextBar, 0, 1, false).
		AddItem(textInputArea, 0, 2, true)
	if err := tview.NewApplication().
		SetRoot(flex, true).
		SetFocus(textInputArea).Run(); err != nil {
		panic(err)
	}
}

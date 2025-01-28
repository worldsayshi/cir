package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type InputArea struct {
	*tview.TextArea
}

func NewInputArea() *InputArea {
	textInputArea := tview.NewTextArea().
		SetPlaceholder("Write here")
	textInputArea.
		SetBorder(true).
		SetTitle("Input")
	return &InputArea{TextArea: textInputArea}
}

func (input *InputArea) SetInputText(inputText string) {
	input.SetText(inputText, true)
}

// Ctrl+S to submit
func (input *InputArea) SetSubmitFunc(submitFunc func(text string)) {
	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			text := input.GetText()
			submitFunc(text)
			return nil
		}
		return event
	})
}

package main

import (
	"github.com/rivo/tview"
)

func main() {
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			// SetTextAlign(tview.AlignCenter).
			SetBorder(true).
			SetTitle(text)
	}

	chatHistory := newPrimitive("History")
	contextBar := newPrimitive("Context")
	textInputArea := tview.NewTextArea().
		SetPlaceholder("Write here")
	textInputArea.
		SetBorder(true).
		SetTitle("Input")

	// grid := tview.NewGrid().
	// 	SetRows(3, 0, 3).
	// 	SetBorders(true)
	//
	// grid.AddItem(chatHistory, 10, 0, 0, 0, 0, 100, false).
	// 	AddItem(contextBar, 1, 0, 0, 0, 0, 100, false).
	// 	AddItem(textInputArea, 2, 2, 2, 2, 0, 100, true)

	// box := tview.NewBox().SetBorder(true).SetTitle("Hello, world!")
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

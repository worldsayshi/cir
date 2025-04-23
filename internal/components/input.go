package components

import (
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type InputArea struct {
	*tview.TextArea
	isInternalUpdate atomic.Bool // Flag to prevent change events during programmatic updates
}

func NewInputArea() *InputArea {
	textInputArea := tview.NewTextArea().
		SetPlaceholder("Write here")
	textInputArea.
		SetBorder(true).
		SetTitle("Input")
	return &InputArea{TextArea: textInputArea}
}

// SetChangedFunc overrides the standard tview changed function to prevent
// infinite loops during programmatic updates
func (input *InputArea) SetChangedFunc(handler func()) {
	originalChangedHandler := handler
	input.TextArea.SetChangedFunc(func() {
		// Only trigger external changes if this is not an internal update
		if !input.isInternalUpdate.Load() {
			originalChangedHandler()
		}
	})
}

// SetText overrides the standard tview SetText to mark updates as internal
func (input *InputArea) SetText(text string, emitChange bool) {
	// Mark that we're doing an internal update to prevent change events
	wasInternal := input.isInternalUpdate.Swap(true)

	// Call the underlying SetText method
	input.TextArea.SetText(text, emitChange)

	// Restore the previous state
	input.isInternalUpdate.Store(wasInternal)
}

// Ctrl+S to submit
func (input *InputArea) SetSubmitFunc(submitFunc func(text string)) {
	originalCapture := input.GetInputCapture()

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle submission with Ctrl+S
		if event.Key() == tcell.KeyCtrlS {
			text := input.GetText()
			submitFunc(text)
			return nil
		}

		// Pass all other keys to the original handler, if any
		if originalCapture != nil {
			return originalCapture(event)
		}

		// Otherwise pass through all other keys
		return event
	})
}

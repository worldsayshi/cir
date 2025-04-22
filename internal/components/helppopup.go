package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/worldsayshi/cir/internal/types"
)

type HelpPopup struct {
	keyMappings []types.KeyMapping
	HasPage     func(name string) bool
	AddPage     func(name string, item tview.Primitive, resize, visible bool)
	RemovePage  func(name string)
}

func NewHelpPopup(
	keyMappings []types.KeyMapping,
	HasPage func(name string) bool,
	AddPage func(name string, item tview.Primitive, resize, visible bool),
	RemovePage func(name string),
) *HelpPopup {
	return &HelpPopup{
		HasPage:     HasPage,
		AddPage:     AddPage,
		RemovePage:  RemovePage,
		keyMappings: keyMappings,
	}
}

// showHelpPopup displays a popup with all available key mappings
func (popup *HelpPopup) ShowHelpPopup() {
	if popup.HasPage("help") {
		popup.RemovePage("help")
		return
	}
	// Create a new modal
	modal := tview.NewModal()
	// modal.SetBorder(true)
	modal.SetTitle(" Keyboard Shortcuts ")

	// Build the content for the modal
	var content strings.Builder
	content.WriteString("Available keyboard shortcuts:\n\n")

	// Add the special case for '?' key separately
	content.WriteString("  ? : Show this help window\n")

	// Add all other key mappings
	for _, mapping := range popup.keyMappings {
		if mapping.Key == tcell.KeyRune {
			continue // Skip the help mapping as we've already added it
		}

		keyName := GetKeyName(mapping.Key)
		content.WriteString(fmt.Sprintf("  %s : %s\n", keyName, mapping.Description))
	}

	modal.SetText(content.String())

	// modal.AddButtons([]string{"Close"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		popup.RemovePage("help")
	})

	// Show the modal as a page
	popup.AddPage("help", modal, true, true)
}

// getKeyName returns a human-readable name for a keyboard key
func GetKeyName(key tcell.Key) string {
	switch key {
	case tcell.KeyTab:
		return "Tab"
	case tcell.KeyBacktab:
		return "Shift+Tab"
	case tcell.KeyCtrlE:
		return "Ctrl+E"
	case tcell.KeyCtrlY:
		return "Ctrl+Y"
	default:
		return fmt.Sprintf("Key(%d)", key)
	}
}

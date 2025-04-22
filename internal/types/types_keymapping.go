package types

import "github.com/gdamore/tcell/v2"

// KeyMapping represents a keyboard shortcut and its associated action
type KeyMapping struct {
	Key         tcell.Key
	Description string
	Action      func()
}

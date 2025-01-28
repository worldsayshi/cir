package components

import (
	"strings"

	"github.com/rivo/tview"
	"github.com/worldsayshi/cir/internal/types"
)

type ContextBar struct {
	*tview.TextView
}

func NewContextBar(workingFiles *[]types.WorkingFile) (contextBar *ContextBar) {
	contextBar = &ContextBar{TextView: tview.NewTextView()}
	contextBar.
		SetBorder(true).
		SetTitle("Context")
	contextBar.Render(*workingFiles)
	return contextBar
}

func (contextBar ContextBar) Render(wf []types.WorkingFile) {
	s := []string{}
	for _, f := range wf {
		s = append(s, f.Path)
	}
	contextBar.SetText(strings.Join(s, " | "))
}

package components

import (
	"strings"

	"github.com/rivo/tview"
	"github.com/worldsayshi/cir/internal/types"
)

func initContextBar(workingSession *types.WorkingSession) *tview.TextView {
	contextBar := tview.NewTextView()
	contextBar.
		SetBorder(true).
		SetTitle("Context")
	RenderContextBar(contextBar, workingSession.WorkingFiles)
	return contextBar
}

func RenderContextBar(contextBar *tview.TextView, wf []WorkingFile) {
	s := []string{}
	for _, f := range wf {
		s = append(s, f.Path)
	}
	contextBar.SetText(strings.Join(s, " | "))
}

package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/worldsayshi/cir/internal/components"
	"github.com/worldsayshi/cir/internal/storage"
	"github.com/worldsayshi/cir/internal/types"
)

// AppState holds all application state
type AppState struct {
	workingSession *types.WorkingSession
	sessionFile    string
	isProcessing   bool
}

// CirApplication manages UI components and application lifecycle
type CirApplication struct {
	*tview.Application
	state         *AppState
	chatHistory   *components.ChatHistory
	inputArea     *components.InputArea
	contextBar    *components.ContextBar
	rootContainer *tview.Flex
}

func NewCirApplication(sessionFile string) *CirApplication {
	// Initialize state
	workingSession, err := storage.LoadWorkingSession(sessionFile)
	if err != nil {
		log.Println("Error loading session from file:", sessionFile)
		panic(fmt.Sprintf("Error loading session from file: %v\n%v", sessionFile, err))
	}

	state := &AppState{
		workingSession: workingSession,
		sessionFile:    sessionFile,
		isProcessing:   false,
	}

	// Initialize UI components
	chatHistory := components.NewChatHistory(nil) // Will be populated during render
	contextBar := components.NewContextBar(nil)   // Will be populated during render
	inputArea := components.NewInputArea()

	cirApp := &CirApplication{
		Application:   tview.NewApplication(),
		state:         state,
		chatHistory:   chatHistory,
		inputArea:     inputArea,
		contextBar:    contextBar,
		rootContainer: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	// Set up UI event handlers
	inputArea.SetChangedFunc(func() {
		cirApp.updateState(func(state *AppState) bool {
			state.workingSession.InputText = inputArea.GetText()
			return false
		})
	})

	inputArea.SetSubmitFunc(cirApp.handleChatSubmit)

	// Setup keyboard handlers
	focusableElements := []tview.Primitive{chatHistory, inputArea}
	cirApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			cirApp.cycleFocus(focusableElements, false)
			return nil
		case tcell.KeyBacktab:
			cirApp.cycleFocus(focusableElements, true)
			return nil
		case tcell.KeyCtrlE:
			cirApp.openSessionFile()
			return nil
		case tcell.KeyCtrlY:
			cirApp.editContextFiles()
			return nil
		}
		return event
	})

	// Initial render of the UI
	cirApp.render()

	return cirApp
}

// updateState applies a state change function and triggers a UI update
func (cirApp *CirApplication) updateState(updateFunc func(*AppState) bool) {
	rerender := updateFunc(cirApp.state)
	if rerender {
		cirApp.render()
	}

	// Save state changes to disk
	if err := storage.SaveWorkingSession(cirApp.state.sessionFile, cirApp.state.workingSession); err != nil {
		log.Println("Error saving session:", err)
	}
}

// render updates all UI components based on current state
func (cirApp *CirApplication) render() {

	cirApp.chatHistory.Render(cirApp.state.workingSession.Messages)
	cirApp.contextBar.Render(cirApp.state.workingSession.WorkingFiles)
	cirApp.inputArea.SetText(cirApp.state.workingSession.InputText, false)

	cirApp.inputArea.SetDisabled(cirApp.state.isProcessing)

	// Set up the layout if it hasn't been done yet
	if cirApp.rootContainer.GetItemCount() == 0 {
		cirApp.rootContainer.
			AddItem(cirApp.chatHistory, 0, 5, false).
			AddItem(cirApp.contextBar, 0, 1, false).
			AddItem(cirApp.inputArea, 0, 2, true)

		cirApp.SetRoot(cirApp.rootContainer, true)
		cirApp.SetFocus(cirApp.inputArea)
	}
}

// From: https://github.com/rivo/tview/issues/100#issuecomment-763131391
func (cirApp *CirApplication) cycleFocus(elements []tview.Primitive, reverse bool) {
	for i, el := range elements {
		if !el.HasFocus() {
			continue
		}

		if reverse {
			i = i - 1
			if i < 0 {
				i = len(elements) - 1
			}
		} else {
			i = i + 1
			i = i % len(elements)
		}

		cirApp.SetFocus(elements[i])
		return
	}
}

func (cirApp *CirApplication) openSessionFile() {
	sessionFindingCommand := `(dir=$(pwd); while [ "$dir" != "/" ]; do find "$dir" -maxdepth 1 \( -name "*.yaml" -o -name "*.yml" \) -exec grep -l "^kind: WorkingSession" {} \; 2>/dev/null; if [ -d "$dir/.cir" ]; then find "$dir/.cir" -maxdepth 1 \( -name "*.yaml" -o -name "*.yml" \) -exec grep -l "^kind: WorkingSession" {} \; 2>/dev/null; fi; dir=$(dirname "$dir"); done)`
	tmuxSessionFindingCommand := sessionFindingCommand + ` | fzf-tmux -h -m`
	out, err := exec.Command(
		"bash", "-c", tmuxSessionFindingCommand,
	).CombinedOutput()
	if err != nil {
		log.Println(err)
		return
	}

	filePath := strings.TrimSpace(string(out))
	if filePath == "" {
		return
	}

	// Load the selected session file
	newWorkingSession, err := storage.LoadWorkingSession(filePath)
	if err != nil {
		log.Printf("Error loading session from file: %v\n%v", filePath, err)
		return
	}

	// Update state with new session
	cirApp.updateState(func(state *AppState) bool {
		state.workingSession = newWorkingSession
		state.sessionFile = filePath
		return true
	})
}

func (cirApp *CirApplication) editContextFiles() {
	cmd := "find . -type f -not -path '*/.*' | fzf-tmux -h -m"
	out, err := exec.Command(
		"bash", "-c", cmd,
	).CombinedOutput()
	if err != nil {
		log.Println(err)
		return
	}

	contextFiles := strings.Split(string(out), "\n")
	// filter out empty strings
	selectedWorkingFiles := []types.WorkingFile{}
	for _, f := range contextFiles {
		if f != "" {
			selectedWorkingFiles = append(selectedWorkingFiles, types.WorkingFile{Path: f})
		}
	}

	cirApp.updateState(func(state *AppState) bool {
		state.workingSession.WorkingFiles = selectedWorkingFiles
		return true
	})
}

func (cirApp *CirApplication) Run() error {
	if err := cirApp.Application.Run(); err != nil {
		return err
	}

	// Final save before exiting
	if err := storage.SaveWorkingSession(cirApp.state.sessionFile, cirApp.state.workingSession); err != nil {
		log.Println("Error saving session:", err)
	}

	return nil
}

// Update the checksums of the files that were submitted
// func (cirApp *CirApplication) updateWorkingFileChecksums(filesToSubmit []types.WorkingFile) {
// 	cirApp.updateState(func(state *AppState) {
// 		for i, wf := range state.workingSession.WorkingFiles {
// 			for _, wfSubmit := range filesToSubmit {
// 				if wf.Path == wfSubmit.Path {
// 					state.workingSession.WorkingFiles[i] = wfSubmit
// 				}
// 			}
// 		}
// 	})
// }

func createSystemMessage() types.Message {
	systemMessage := types.Message{
		AiServiceMessage: types.AiServiceMessage{
			Role: "system",
			Content: `The assistant is Cir, a conversational AI and a coding assistant.
Any code or other file content rendered by the assistant should be rendered with markdown
backticks and should always specify the file path as the comment in the first line of the code block.
Example:
` + "```python" + `
# main.py
print("Hello, World!")
` + "```",
		},
	}
	return systemMessage
}

func (cirApp *CirApplication) handleChatSubmit(text string) {
	if text == "" {
		return
	}

	cirApp.updateState(func(state *AppState) bool {
		// Initialize with system message if needed
		if len(state.workingSession.Messages) == 0 {
			state.workingSession.Messages = append(state.workingSession.Messages,
				createSystemMessage())
		}

		filesToSubmit := getFilesToSubmitWithChecksums(state.workingSession.WorkingFiles)
		userMessage := prepareUserMessage(filesToSubmit, text)

		// Add user message
		state.workingSession.Messages = append(
			state.workingSession.Messages, userMessage,
		)

		// Update file checksums
		for i, wf := range state.workingSession.WorkingFiles {
			for _, wfSubmit := range filesToSubmit {
				if wf.Path == wfSubmit.Path {
					state.workingSession.WorkingFiles[i] = wfSubmit
				}
			}
		}

		// Clear input and set processing state
		state.workingSession.InputText = ""
		state.isProcessing = true

		// Add empty message for streaming response
		state.workingSession.Messages = append(
			state.workingSession.Messages,
			types.Message{
				AiServiceMessage: types.AiServiceMessage{Role: "assistant", Content: ""},
			},
		)
		return true
	})

	// Get service messages for API call
	serviceMessages := getServiceMessages(cirApp.state.workingSession.Messages)

	// Start streaming
	resultChan, errChan := streamOpenAI(serviceMessages)

	// Create a goroutine to handle streaming updates
	go cirApp.handleStreamResponse(resultChan, errChan)
}

func (cirApp *CirApplication) handleStreamResponse(resultChan chan string, errChan chan error) {
	accumulated := ""

	for {
		select {
		case chunk, ok := <-resultChan:
			if !ok {
				// Stream completed
				cirApp.updateState(func(state *AppState) bool {
					state.isProcessing = false
					return true
				})
				return
			}

			accumulated += chunk

			cirApp.updateState(func(state *AppState) bool {
				lastIdx := len(state.workingSession.Messages) - 1
				state.workingSession.Messages[lastIdx].AiServiceMessage.Content = accumulated
				return true
			})

		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				cirApp.updateState(func(state *AppState) bool {
					lastIdx := len(state.workingSession.Messages) - 1
					state.workingSession.Messages[lastIdx].Content = fmt.Sprintf("Error: %v", err)
					state.isProcessing = false
					return true
				})
				return
			}
		}
	}
}

// Add WorkingFiles to the content iff checksum is nill or changed
func getFilesToSubmitWithChecksums(wfs []types.WorkingFile) []types.WorkingFile {
	filesToSubmit := []types.WorkingFile{}
	for _, wf := range wfs {
		fileContents, err := os.ReadFile(wf.Path)
		if err != nil {
			log.Println("Error reading context file:", wf.Path, err)
			continue
		}
		checksum := fmt.Sprintf("%x", md5.Sum(fileContents))
		if wf.LastSubmittedChecksum == nil {
			wf.LastSubmittedChecksum = &checksum
			wf.FileContent = fileContents
			filesToSubmit = append(filesToSubmit, wf)
			continue
		}
		if checksum != *wf.LastSubmittedChecksum {
			wf.LastSubmittedChecksum = &checksum
			wf.FileContent = fileContents
			filesToSubmit = append(filesToSubmit, wf)
		}
	}
	return filesToSubmit
}

func getServiceMessages(messages []types.Message) []types.AiServiceMessage {
	lastIdx := len(messages) - 1

	serviceMessages := []types.AiServiceMessage{}
	for _, msg := range messages[:lastIdx] {
		serviceMessages = append(serviceMessages, msg.AiServiceMessage)
	}
	return serviceMessages
}

var promptTemplate string = `{{- range .workingFiles -}}
<context file="{{.Path}}">
{{ printf "%s" .FileContent }}
</context>
{{- end }}
<question>
{{.question}}
</question>`

// Prepare the user message using the template
func prepareUserMessage(filesToSubmit []types.WorkingFile, question string) types.Message {
	var buf bytes.Buffer
	templ := template.Must(template.New("promptTemplate").Parse(promptTemplate))
	templ.Execute(&buf, map[string]interface{}{
		"workingFiles": filesToSubmit,
		"question":     question,
	})
	content := buf.String()
	userMessage := types.Message{
		AiServiceMessage:     types.AiServiceMessage{Role: "user", Content: content},
		Question:             question,
		IncludedWorkingFiles: filesToSubmit,
	}
	return userMessage
}

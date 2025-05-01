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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/worldsayshi/cir/internal/storage"
	"github.com/worldsayshi/cir/internal/types"
	"github.com/worldsayshi/cir/internal/ui"
)

// CirApplication manages the application state and UI
type CirApplication struct {
	state    *ui.AppState
	program  *tea.Program
	teaModel ui.Model
}

func NewCirApplication(sessionFile string) *CirApplication {
	// Initialize state
	workingSession, err := storage.LoadWorkingSession(sessionFile)
	if err != nil {
		log.Println("Error loading session from file:", sessionFile)
		panic(fmt.Sprintf("Error loading session from file: %v\n%v", sessionFile, err))
	}

	state := &ui.AppState{
		WorkingSession: workingSession,
		SessionFile:    sessionFile,
		IsProcessing:   false,
	}

	teaModel := ui.NewModel(state)

	cirApp := &CirApplication{
		state:    state,
		teaModel: teaModel,
	}

	return cirApp
}

func (cirApp *CirApplication) Init() {
	// Set up callback functions for the UI
	cirApp.teaModel.SetAppCallbacks(ui.AppCallbacks{
		SubmitMessage:    cirApp.SubmitMessage,
		OpenSessionFile:  cirApp.OpenSessionFile,
		EditContextFiles: cirApp.EditContextFiles,
	})

	// Update the UI with the loaded session data
	cirApp.teaModel.UpdateChatHistory()
	cirApp.teaModel.UpdateContextBar()

	// Set the input text from the session
	if cirApp.state.WorkingSession.InputText != "" {
		cirApp.teaModel.SetInputText(cirApp.state.WorkingSession.InputText)
	}
}

func (cirApp *CirApplication) Run() error {
	// Initialize Bubble Tea
	p := tea.NewProgram(cirApp.teaModel, tea.WithAltScreen())
	cirApp.program = p

	// Run the Bubble Tea program
	model, err := p.Run()
	if err != nil {
		return err
	}

	// Extract the final model state
	finalModel := model.(ui.Model)

	// Update input text in session with what's in the input area
	cirApp.state.WorkingSession.InputText = finalModel.GetInputText()

	// Final save before exiting
	if err := storage.SaveWorkingSession(cirApp.state.SessionFile, cirApp.state.WorkingSession); err != nil {
		log.Println("Error saving session:", err)
	}

	return nil
}

// Process server-side logic for submitting messages
func (cirApp *CirApplication) SubmitMessage(text string) tea.Cmd {
	return func() tea.Msg {
		if text == "" {
			return nil
		}

		// Initialize with system message if needed
		if len(cirApp.state.WorkingSession.Messages) == 0 {
			cirApp.state.WorkingSession.Messages = append(cirApp.state.WorkingSession.Messages,
				createSystemMessage())
		}

		filesToSubmit := getFilesToSubmitWithChecksums(cirApp.state.WorkingSession.WorkingFiles)
		userMessage := prepareUserMessage(filesToSubmit, text)

		// Add user message
		cirApp.state.WorkingSession.Messages = append(
			cirApp.state.WorkingSession.Messages, userMessage,
		)

		// Update file checksums
		for i, wf := range cirApp.state.WorkingSession.WorkingFiles {
			for _, wfSubmit := range filesToSubmit {
				if wf.Path == wfSubmit.Path {
					cirApp.state.WorkingSession.WorkingFiles[i] = wfSubmit
				}
			}
		}

		// Clear input and set processing state
		cirApp.state.WorkingSession.InputText = ""
		cirApp.state.IsProcessing = true

		// Add empty message for streaming response
		cirApp.state.WorkingSession.Messages = append(
			cirApp.state.WorkingSession.Messages,
			types.Message{
				AiServiceMessage: types.AiServiceMessage{Role: "assistant", Content: ""},
			},
		)

		// Save state after changes
		storage.SaveWorkingSession(cirApp.state.SessionFile, cirApp.state.WorkingSession)

		// Notify UI of changes
		cirApp.program.Send(ui.SessionUpdatedMsg{})

		// Get service messages for API call
		serviceMessages := getServiceMessages(cirApp.state.WorkingSession.Messages)

		// Start streaming in a goroutine
		resultChan, errChan := streamOpenAI(serviceMessages)
		go cirApp.handleStreamResponse(resultChan, errChan)

		return ui.SubmitMessageMsg{Text: text}
	}
}

func (cirApp *CirApplication) handleStreamResponse(resultChan chan string, errChan chan error) {
	accumulated := ""

	for {
		select {
		case chunk, ok := <-resultChan:
			if !ok {
				// Stream completed
				cirApp.state.IsProcessing = false

				// Save the final response
				storage.SaveWorkingSession(cirApp.state.SessionFile, cirApp.state.WorkingSession)

				cirApp.program.Send(ui.StreamResponseDoneMsg{})
				return
			}

			accumulated += chunk

			// Update the message content with accumulated text
			lastIdx := len(cirApp.state.WorkingSession.Messages) - 1
			cirApp.state.WorkingSession.Messages[lastIdx].AiServiceMessage.Content = accumulated

			// Send the chunk to the UI for display
			cirApp.program.Send(ui.StreamResponseChunkMsg{Chunk: chunk})

		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				// Update the message content with error
				lastIdx := len(cirApp.state.WorkingSession.Messages) - 1
				cirApp.state.WorkingSession.Messages[lastIdx].Content = fmt.Sprintf("Error: %v", err)

				// Set processing to false
				cirApp.state.IsProcessing = false

				// Save the error state
				storage.SaveWorkingSession(cirApp.state.SessionFile, cirApp.state.WorkingSession)

				// Notify UI of error
				cirApp.program.Send(ui.StreamResponseErrorMsg{Err: err})
				return
			}
		}
	}
}

// OpenSessionFile opens a session selector to load a different session
func (cirApp *CirApplication) OpenSessionFile() tea.Cmd {
	return func() tea.Msg {
		// Exit alt screen mode temporarily to allow fzf to work
		if cirApp.program != nil {
			cirApp.program.ExitAltScreen()
			defer cirApp.program.EnterAltScreen()
		}

		sessionFindingCommand := `(dir=$(pwd); while [ "$dir" != "/" ]; do find "$dir" -maxdepth 1 \( -name "*.yaml" -o -name "*.yml" \) -exec grep -l "^kind: WorkingSession" {} \; 2>/dev/null; if [ -d "$dir/.cir" ]; then find "$dir/.cir" -maxdepth 1 \( -name "*.yaml" -o -name "*.yml" \) -exec grep -l "^kind: WorkingSession" {} \; 2>/dev/null; fi; dir=$(dirname "$dir"); done)`
		tmuxSessionFindingCommand := sessionFindingCommand + ` | fzf-tmux -h -m`
		out, err := exec.Command(
			"bash", "-c", tmuxSessionFindingCommand,
		).CombinedOutput()
		if err != nil {
			log.Println(err)
			return nil
		}

		filePath := strings.TrimSpace(string(out))
		if filePath == "" {
			return nil
		}

		// Load the selected session file
		newWorkingSession, err := storage.LoadWorkingSession(filePath)
		if err != nil {
			log.Printf("Error loading session from file: %v\n%v", filePath, err)
			return nil
		}

		// Update state with new session
		cirApp.state.WorkingSession = newWorkingSession
		cirApp.state.SessionFile = filePath

		// Save new session
		storage.SaveWorkingSession(cirApp.state.SessionFile, cirApp.state.WorkingSession)

		return ui.SessionUpdatedMsg{}
	}
}

// EditContextFiles allows selecting context files
func (cirApp *CirApplication) EditContextFiles() tea.Cmd {
	return func() tea.Msg {
		// Exit alt screen mode temporarily to allow fzf to work
		if cirApp.program != nil {
			cirApp.program.ExitAltScreen()
			defer cirApp.program.EnterAltScreen()
		}

		cmd := "find . -type f -not -path '*/.*' | fzf-tmux -h -m | cat"
		out, err := exec.Command(
			"bash", "-c", cmd,
		).CombinedOutput()
		if err != nil {
			log.Println("Error executing command:", cmd)
			log.Println(err)
			return nil
		}

		contextFiles := strings.Split(string(out), "\n")
		// filter out empty strings
		selectedWorkingFiles := []types.WorkingFile{}
		for _, f := range contextFiles {
			if f != "" {
				selectedWorkingFiles = append(selectedWorkingFiles, types.WorkingFile{Path: f})
			}
		}

		cirApp.state.WorkingSession.WorkingFiles = selectedWorkingFiles
		storage.SaveWorkingSession(cirApp.state.SessionFile, cirApp.state.WorkingSession)

		return ui.SessionUpdatedMsg{}
	}
}

// Add WorkingFiles to the content iff checksum is nil or changed
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

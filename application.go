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
	"github.com/worldsayshi/cir/internal/types"
)

type CirApplication struct {
	*tview.Application
	chatHistory    *components.ChatHistory
	inputArea      *components.InputArea
	contextBar     *components.ContextBar
	workingSession *types.WorkingSession
	sessionFile    string
}

func NewCirApplication(sessionFile string) *CirApplication {
	workingSession, err := loadWorkingSession(sessionFile)
	if err != nil {
		log.Println("Error loading session from file:", sessionFile)
		panic(fmt.Sprintf("Error loading session from file: %v\n%v", sessionFile, err))
	}

	// Chat history
	chatHistory := components.NewChatHistory(workingSession.Messages)

	// Context bar
	contextBar := components.NewContextBar(workingSession.WorkingFiles)

	// Text input area
	inputArea := components.NewInputArea()

	cirApp := &CirApplication{
		Application:    tview.NewApplication(),
		chatHistory:    chatHistory,
		inputArea:      inputArea,
		contextBar:     contextBar,
		workingSession: workingSession,
		sessionFile:    sessionFile,
	}

	// Redraw chat history when it changes
	chatHistory.SetChangedFunc(func() {
		cirApp.Draw()
	})

	inputArea.SetInputText(workingSession.InputText)

	// Update input text in working session
	inputArea.SetChangedFunc(func() {
		cirApp.workingSession.InputText = inputArea.GetText()
	})

	inputArea.SetSubmitFunc(cirApp.handleChatSubmit)
	cirApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	focusableElements := []tview.Primitive{chatHistory, inputArea}
	cirApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// Tab and Shift+Tab to cycle focus
		case tcell.KeyTab:
			cirApp.cycleFocus(focusableElements, false)
			return nil
		case tcell.KeyBacktab:
			cirApp.cycleFocus(focusableElements, true)
			return nil
		case tcell.KeyCtrlE:
			cirApp.openSessionFile()
			return nil
		// Ctrl+Y to edit context files
		case tcell.KeyCtrlY:
			cirApp.editContextFiles()
			return nil
		}
		return event
	})

	return cirApp
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
	sessionFindingCommand := `find $(pwd) $(while [ "$(pwd)" != "/" ]; do cd ..; echo $(pwd); done) -path "*/.cir" -o -path "$(pwd)" -o -path "*/$(basename $(pwd))" | xargs -I{} find {} -type f \( -name "*.yaml" -o -name "*.yml" \) -exec grep -l "^kind: WorkingSession" {} \;`
	tmuxSessionFindingCommand := sessionFindingCommand + ` | fzf-tmux -h -m` // ` | xargs -I{} tmux split-window -h -p 50 -c {}`
	//tmuxSessionFindingCommand := `tmux list-panes -F "#{pane_current_path}" | xargs -I{} find {} -type f \( -name "*.yaml" -o -name "*.yml" \) -exec grep -l "^kind: WorkingSession" {} \;`
	out, err := exec.Command(
		"bash", "-c", tmuxSessionFindingCommand,
	).CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(out))
}

func (cirApp *CirApplication) editContextFiles() {
	cmd := "find . -type f -not -path '*/.*' | fzf-tmux -h -m"
	out, err := exec.Command(
		"bash", "-c", cmd,
	).CombinedOutput()
	if err != nil {
		log.Println(err)
	}

	contextFiles := strings.Split(string(out), "\n")
	// filter out empty strings
	selectedWorkingFiles := []types.WorkingFile{}
	for _, f := range contextFiles {
		if f != "" {
			selectedWorkingFiles = append(selectedWorkingFiles, types.WorkingFile{Path: f})
		}
	}
	cirApp.workingSession.WorkingFiles = selectedWorkingFiles
	if err := saveWorkingSession(cirApp.sessionFile, cirApp.workingSession); err != nil {
		panic(err)
	}
	cirApp.contextBar.Render(cirApp.workingSession.WorkingFiles)
}

func (cirApp *CirApplication) Run() error {
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(cirApp.chatHistory, 0, 5, false).
		AddItem(cirApp.contextBar, 0, 1, false).
		AddItem(cirApp.inputArea, 0, 2, true)
	if err := cirApp.
		SetRoot(flex, true).
		SetFocus(cirApp.inputArea).Run(); err != nil {
		panic(err)
	}
	defer func() {
		if err := saveWorkingSession(cirApp.sessionFile, cirApp.workingSession); err != nil {
			log.Println("Error saving session:", err)
		}
	}()
	return nil
}

// Update the checksums of the files that were submitted
// Checksums have already been calculated in the filesToSubmit
func (cirApp *CirApplication) updateWorkingFileChecksums(filesToSubmit []types.WorkingFile) {
	for i, wf := range cirApp.workingSession.WorkingFiles {
		for _, wfSubmit := range filesToSubmit {
			if wf.Path == wfSubmit.Path {
				cirApp.workingSession.WorkingFiles[i] = wfSubmit
			}
		}
	}
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

func (cirApp *CirApplication) handleChatSubmit(text string) {
	if text != "" {
		if len(cirApp.workingSession.Messages) == 0 {
			cirApp.workingSession.Messages = append(cirApp.workingSession.Messages,
				createSystemMessage())
		}

		filesToSubmit := getFilesToSubmitWithChecksums(cirApp.workingSession.WorkingFiles)

		userMessage := prepareUserMessage(filesToSubmit, text)
		cirApp.workingSession.Messages = append(
			cirApp.workingSession.Messages, userMessage,
		)
		cirApp.updateWorkingFileChecksums(filesToSubmit)
		cirApp.chatHistory.Render(cirApp.workingSession.Messages)
		cirApp.inputArea.SetText("", true)
		if err := saveWorkingSession(cirApp.sessionFile, cirApp.workingSession); err != nil {
			panic(err)
		}

		// Lock the text input area
		cirApp.inputArea.SetDisabled(true)

		// Add empty message for streaming response
		cirApp.workingSession.Messages = append(
			cirApp.workingSession.Messages,
			types.Message{
				AiServiceMessage:     types.AiServiceMessage{Role: "assistant", Content: ""},
				Question:             "",
				IncludedWorkingFiles: []types.WorkingFile{},
			},
		)

		serviceMessages := getServiceMessages(cirApp.workingSession.Messages)

		// Start streaming
		resultChan, errChan := streamOpenAI(serviceMessages)

		// Create a goroutine to handle streaming updates
		go cirApp.handleStreamResponse(resultChan, errChan)
	}
}

func (cirApp *CirApplication) handleStreamResponse(resultChan chan string, errChan chan error) {
	accumulated := ""
	lastIdx := len(cirApp.workingSession.Messages) - 1
	for {
		select {
		case chunk, ok := <-resultChan:
			if !ok {
				// Stream completed
				if err := saveWorkingSession(cirApp.sessionFile, cirApp.workingSession); err != nil {
					panic(err)
				}
				cirApp.inputArea.SetDisabled(false)
				return
			}
			accumulated += chunk
			cirApp.workingSession.Messages[lastIdx].AiServiceMessage.Content = accumulated
			cirApp.chatHistory.Render(cirApp.workingSession.Messages)
		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				cirApp.workingSession.Messages[lastIdx].Content = fmt.Sprintf("Error: %v", err)
				cirApp.chatHistory.Render(cirApp.workingSession.Messages)
				cirApp.inputArea.SetDisabled(false)
				if err := saveWorkingSession(cirApp.sessionFile, cirApp.workingSession); err != nil {
					panic(err)
				}
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

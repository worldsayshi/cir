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
	chatHistory    *tview.TextView
	inputArea      *components.InputArea
	contextBar     *components.ContextBar
	workingSession *types.WorkingSession
	sessionFile    string
}

// From: https://github.com/rivo/tview/issues/100#issuecomment-763131391
func cycleFocus(cirApp *CirApplication, elements []tview.Primitive, reverse bool) {
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

func NewCirApplication(sessionFile string) *CirApplication {
	workingSession, err := loadWorkingSession(sessionFile)
	if err != nil {
		log.Println("Error loading session from file:", sessionFile)
		panic(fmt.Sprintf("Error loading session from file: %v\n%v", sessionFile, err))
	}

	// Chat history
	chatHistory := components.InitChatHistory(workingSession)

	// Context bar
	contextBar := components.NewContextBar(&workingSession.WorkingFiles)

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
			cycleFocus(cirApp, focusableElements, false)
			return nil
		case tcell.KeyBacktab:
			cycleFocus(cirApp, focusableElements, true)
			return nil
		// Ctrl+O to edit context files
		case tcell.KeyCtrlO:
			cirApp.editContextFiles()
			return nil
		}
		return event
	})

	return cirApp
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

// Add WorkingFiles to the content iff checksum is nill or changed
func (cirApp *CirApplication) getFilesToSubmit() []types.WorkingFile {
	filesToSubmit := []types.WorkingFile{}
	for _, wf := range cirApp.workingSession.WorkingFiles {
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

var promptTemplate string = `{{- range .workingFiles -}}
<context file="{{.Path}}">
{{ printf "%s" .FileContent }}
</context>
{{- end }}
<question>
{{.question}}
</question>`

// Add WorkingFiles to the content iff checksum is nill or changed
func prepareUserMessage(filesToSubmit []types.WorkingFile, question string) string {
	var buf bytes.Buffer
	templ := template.Must(template.New("promptTemplate").Parse(promptTemplate))
	templ.Execute(&buf, map[string]interface{}{
		"workingFiles": filesToSubmit,
		"question":     question,
	})
	return buf.String()
}

// Update the checksums of the files that were submitted
func (cirApp *CirApplication) updateWorkingFileChecksums(filesToSubmit []types.WorkingFile) {
	for i, wf := range cirApp.workingSession.WorkingFiles {
		for _, wfSubmit := range filesToSubmit {
			if wf.Path == wfSubmit.Path {
				cirApp.workingSession.WorkingFiles[i] = wfSubmit
			}
		}
	}
}

func (cirApp *CirApplication) handleChatSubmit(text string) {
	if text != "" {
		filesToSubmit := cirApp.getFilesToSubmit()
		content := prepareUserMessage(filesToSubmit, text)
		cirApp.workingSession.Messages = append(
			cirApp.workingSession.Messages,
			types.Message{
				AiServiceMessage:     types.AiServiceMessage{Role: "user", Content: content},
				Question:             text,
				IncludedWorkingFiles: filesToSubmit,
			})
		cirApp.updateWorkingFileChecksums(filesToSubmit)
		components.RenderChatHistory(cirApp.chatHistory, cirApp.workingSession.Messages)
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
				AiServiceMessage:     types.AiServiceMessage{Role: "system", Content: ""},
				Question:             "",
				IncludedWorkingFiles: []types.WorkingFile{},
			},
		)
		lastIdx := len(cirApp.workingSession.Messages) - 1

		serviceMessages := []types.AiServiceMessage{}
		for _, msg := range cirApp.workingSession.Messages[:lastIdx] {
			serviceMessages = append(serviceMessages, msg.AiServiceMessage)
		}

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
			components.RenderChatHistory(cirApp.chatHistory, cirApp.workingSession.Messages)
		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				cirApp.workingSession.Messages[lastIdx].Content = fmt.Sprintf("Error: %v", err)
				components.RenderChatHistory(cirApp.chatHistory, cirApp.workingSession.Messages)
				cirApp.inputArea.SetDisabled(false)
				if err := saveWorkingSession(cirApp.sessionFile, cirApp.workingSession); err != nil {
					panic(err)
				}
				return
			}
		}
	}
}

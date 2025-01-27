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
)

type AiServiceMessage struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
}

type Message struct {
	AiServiceMessage     `json:"aiServiceMessage,omitempty" yaml:"aiServiceMessage,omitempty"`
	Question             string        `json:"question,omitempty" yaml:"question,omitempty"`
	IncludedWorkingFiles []WorkingFile `json:"included_working_files,omitempty" yaml:"included_working_files,omitempty"`
}

type WorkingFile struct {
	Path                  string  `json:"path" yaml:"path"`
	LastSubmittedChecksum *string `json:"last_submitted_checksum,omitempty" yaml:"last_submitted_checksum,omitempty"`
	FileContent           []byte  `json:"-" yaml:"-"` // Don't serialize this field
}

type WorkingSession struct {
	Messages     []Message     `json:"messages" yaml:"messages"`
	WorkingFiles []WorkingFile `json:"working_files" yaml:"working_files"`
	InputText    string        `json:"input_text" yaml:"input_text"`
}

type CirApplication struct {
	app            *tview.Application
	chatHistory    *tview.TextView
	textInputArea  *tview.TextArea
	contextBar     *tview.TextView
	workingSession *WorkingSession
	sessionFile    string
}

// From: https://github.com/rivo/tview/issues/100#issuecomment-763131391
func cycleFocus(app *tview.Application, elements []tview.Primitive, reverse bool) {
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

		app.SetFocus(elements[i])
		return
	}
}

func initChatHistory(workingSession *WorkingSession) *tview.TextView {
	chatHistory := tview.NewTextView()
	chatHistory.
		SetBorder(true).
		SetTitle("History")
	renderChatHistory(chatHistory, workingSession.Messages)
	return chatHistory
}

func renderChatHistory(chatHistory *tview.TextView, messages []Message) {
	msgsString := []string{}
	for _, msg := range messages {
		if msg.Role == "user" {
			msgsString = append(msgsString, msg.Question)
		} else {
			msgsString = append(msgsString, msg.Content)
		}
	}
	chatHistory.SetText(strings.Join(msgsString, "\n\n---\n"))
	chatHistory.ScrollToEnd()
}

func initContextBar(workingSession *WorkingSession) *tview.TextView {
	contextBar := tview.NewTextView()
	contextBar.
		SetBorder(true).
		SetTitle("Context")
	renderContextBar(contextBar, workingSession.WorkingFiles)
	return contextBar
}

func renderContextBar(contextBar *tview.TextView, wf []WorkingFile) {
	s := []string{}
	for _, f := range wf {
		s = append(s, f.Path)
	}
	contextBar.SetText(strings.Join(s, " | "))
}

func (app *CirApplication) editContextFiles() {
	cmd := "find . -type f -not -path '*/.*' | fzf-tmux -h -m"
	out, err := exec.Command(
		"bash", "-c", cmd,
	).CombinedOutput()
	if err != nil {
		log.Println(err)
	}

	contextFiles := strings.Split(string(out), "\n")
	// filter out empty strings
	filteredWorkingFiles := []WorkingFile{}
	for _, f := range contextFiles {
		if f != "" {
			filteredWorkingFiles = append(filteredWorkingFiles, WorkingFile{Path: f})
		}
	}
	app.workingSession.WorkingFiles = filteredWorkingFiles
	if err := saveWorkingSession(app.sessionFile, app.workingSession); err != nil {
		panic(err)
	}
	renderContextBar(app.contextBar, app.workingSession.WorkingFiles)
}

func initTextInputArea(workingSession *WorkingSession) *tview.TextArea {
	textInputArea := tview.NewTextArea().
		SetPlaceholder("Write here")
	textInputArea.
		SetBorder(true).
		SetTitle("Input")
	textInputArea.
		SetText(workingSession.InputText, true)
	return textInputArea
}

func NewCirApplication(sessionFile string) *CirApplication {
	app := tview.NewApplication()

	workingSession, err := loadWorkingSession(sessionFile)
	if err != nil {
		log.Println("Error loading session from file:", sessionFile)
		panic(fmt.Sprintf("Error loading session from file: %v\n%v", sessionFile, err))
	}

	// Chat history
	chatHistory := initChatHistory(workingSession)

	// Context bar
	contextBar := initContextBar(workingSession)

	// Text input area
	textInputArea := initTextInputArea(workingSession)

	cirApp := &CirApplication{
		app:            app,
		chatHistory:    chatHistory,
		textInputArea:  textInputArea,
		contextBar:     contextBar,
		workingSession: workingSession,
		sessionFile:    sessionFile,
	}

	// Redraw chat history when it changes
	chatHistory.SetChangedFunc(func() {
		app.Draw()
	})

	// Update input text in working session
	textInputArea.SetChangedFunc(func() {
		cirApp.workingSession.InputText = textInputArea.GetText()
	})

	// Ctrl+S to submit
	textInputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			cirApp.handleChatSubmit()
			return nil
		}
		if event.Key() == tcell.KeyCtrlO {
			cirApp.editContextFiles()
			return nil
		}
		return event
	})

	// Tab and Shift+Tab to cycle focus
	focusableElements := []tview.Primitive{chatHistory, textInputArea}
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			cycleFocus(app, focusableElements, false)
			return nil
		case tcell.KeyBacktab:
			cycleFocus(app, focusableElements, true)
			return nil
		}
		return event
	})

	return cirApp
}

func (app *CirApplication) Run() error {
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.chatHistory, 0, 5, false).
		AddItem(app.contextBar, 0, 1, false).
		AddItem(app.textInputArea, 0, 2, true)
	if err := app.app.
		SetRoot(flex, true).
		SetFocus(app.textInputArea).Run(); err != nil {
		panic(err)
	}
	defer func() {
		if err := saveWorkingSession(app.sessionFile, app.workingSession); err != nil {
			log.Println("Error saving session:", err)
		}
	}()
	return nil
}

// Add WorkingFiles to the content iff checksum is nill or changed
func (app *CirApplication) getFilesToSubmit() []WorkingFile {
	filesToSubmit := []WorkingFile{}
	for _, wf := range app.workingSession.WorkingFiles {
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
func prepareUserMessage(filesToSubmit []WorkingFile, question string) string {
	var buf bytes.Buffer
	templ := template.Must(template.New("promptTemplate").Parse(promptTemplate))
	templ.Execute(&buf, map[string]interface{}{
		"workingFiles": filesToSubmit,
		"question":     question,
	})
	return buf.String()
}

// Update the checksums of the files that were submitted
func (app *CirApplication) updateWorkingFileChecksums(filesToSubmit []WorkingFile) {
	for i, wf := range app.workingSession.WorkingFiles {
		for _, wfSubmit := range filesToSubmit {
			if wf.Path == wfSubmit.Path {
				app.workingSession.WorkingFiles[i] = wfSubmit
			}
		}
	}
}

func (app *CirApplication) handleChatSubmit() {
	text := app.textInputArea.GetText()
	if text != "" {
		filesToSubmit := app.getFilesToSubmit()
		content := prepareUserMessage(filesToSubmit, text)
		app.workingSession.Messages = append(
			app.workingSession.Messages,
			Message{
				AiServiceMessage:     AiServiceMessage{Role: "user", Content: content},
				Question:             text,
				IncludedWorkingFiles: filesToSubmit,
			})
		app.updateWorkingFileChecksums(filesToSubmit)
		renderChatHistory(app.chatHistory, app.workingSession.Messages)
		app.textInputArea.SetText("", true)
		if err := saveWorkingSession(app.sessionFile, app.workingSession); err != nil {
			panic(err)
		}

		// Lock the text input area
		app.textInputArea.SetDisabled(true)

		// Add empty message for streaming response
		app.workingSession.Messages = append(
			app.workingSession.Messages,
			Message{
				AiServiceMessage:     AiServiceMessage{Role: "system", Content: ""},
				Question:             "",
				IncludedWorkingFiles: []WorkingFile{},
			},
		)
		lastIdx := len(app.workingSession.Messages) - 1

		serviceMessages := []AiServiceMessage{}
		for _, msg := range app.workingSession.Messages[:lastIdx] {
			serviceMessages = append(serviceMessages, msg.AiServiceMessage)
		}

		// Start streaming
		resultChan, errChan := streamOpenAI(serviceMessages)

		// Create a goroutine to handle streaming updates
		go app.handleStreamResponse(resultChan, errChan)
	}
}

func (app *CirApplication) handleStreamResponse(resultChan chan string, errChan chan error) {
	accumulated := ""
	lastIdx := len(app.workingSession.Messages) - 1
	for {
		select {
		case chunk, ok := <-resultChan:
			if !ok {
				// Stream completed
				if err := saveWorkingSession(app.sessionFile, app.workingSession); err != nil {
					panic(err)
				}
				app.textInputArea.SetDisabled(false)
				return
			}
			accumulated += chunk
			app.workingSession.Messages[lastIdx].AiServiceMessage.Content = accumulated
			renderChatHistory(app.chatHistory, app.workingSession.Messages)
		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				app.workingSession.Messages[lastIdx].Content = fmt.Sprintf("Error: %v", err)
				renderChatHistory(app.chatHistory, app.workingSession.Messages)
				app.textInputArea.SetDisabled(false)
				if err := saveWorkingSession(app.sessionFile, app.workingSession); err != nil {
					panic(err)
				}
				return
			}
		}
	}
}

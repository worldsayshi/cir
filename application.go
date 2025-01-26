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

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type WorkingFile struct {
	Path                  string  `json:"path"`
	LastSubmittedChecksum *string `json:"last_submitted_checksum,omitempty"`
	FileContent           []byte  `json:"file_content,omitempty"`
}

type WorkingSession struct {
	Messages     []Message     `json:"messages"`
	WorkingFiles []WorkingFile `json:"working_files"`
}

type CirApplication struct {
	app            *tview.Application
	chatHistory    *tview.TextView
	textInputArea  *tview.TextArea
	contextBar     *tview.TextView
	workingSession *WorkingSession
	sessionFile    string
}

func NewCirApplication(sessionFile string) *CirApplication {
	app := tview.NewApplication()
	newPrimitive := func(text string) tview.Primitive {
		p := tview.NewTextView()
		p.
			SetBorder(true).
			SetTitle(text)
		return p
	}

	workingSession, err := loadWorkingSession(sessionFile)
	if err != nil {
		panic(err)
	}

	// Chat history
	chatHistory := newPrimitive("History").(*tview.TextView)
	renderMessages(chatHistory, workingSession.Messages)

	// Context bar
	contextBar := newPrimitive("Context").(*tview.TextView)
	renderContextFiles(contextBar, workingSession.WorkingFiles)

	// Text input area
	textInputArea := tview.NewTextArea().
		SetPlaceholder("Write here")
	textInputArea.
		SetBorder(true).
		SetTitle("Input")
	chatHistory.
		SetChangedFunc(func() {
			app.Draw()
		})

	cirApp := &CirApplication{
		app:            app,
		chatHistory:    chatHistory,
		textInputArea:  textInputArea,
		contextBar:     contextBar,
		workingSession: workingSession,
		sessionFile:    sessionFile,
	}

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

	return cirApp
}

func renderMessages(chatHistory *tview.TextView, messages []Message) {
	msgsString := []string{}
	for _, msg := range messages {
		msgsString = append(msgsString, msg.Content)
	}
	chatHistory.SetText(strings.Join(msgsString, "\n\n---\n"))
	chatHistory.ScrollToEnd()
}

func renderContextFiles(contextBar *tview.TextView, wf []WorkingFile) {
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
	renderContextFiles(app.contextBar, app.workingSession.WorkingFiles)
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

var promptTemplate string = `
{{range .workingFiles}}
<context file="{{.Path}}">
{{.FileContent}}
</context>
{{end}}
<question>
{{.Question}}
</question>
`

// Add WorkingFiles to the content iff checksum is nill or changed
func (app *CirApplication) prepareUserMessage(question string) string {
	filesToSubmit := app.getFilesToSubmit()
	var buf bytes.Buffer
	templ := template.Must(template.New("promptTemplate").Parse(promptTemplate))
	templ.Execute(&buf, map[string]interface{}{
		"workingFiles": filesToSubmit,
		"question":     question,
	})
	return buf.String()
}

func (app *CirApplication) handleChatSubmit() {
	text := app.textInputArea.GetText()
	if text != "" {
		content := app.prepareUserMessage(text)
		app.workingSession.Messages = append(app.workingSession.Messages, Message{Role: "user", Content: content})
		renderMessages(app.chatHistory, app.workingSession.Messages)
		app.textInputArea.SetText("", true)
		if err := saveWorkingSession(app.sessionFile, app.workingSession); err != nil {
			panic(err)
		}

		// Lock the text input area
		app.textInputArea.SetDisabled(true)

		// Add empty message for streaming response
		app.workingSession.Messages = append(app.workingSession.Messages, Message{Role: "system", Content: ""})
		lastIdx := len(app.workingSession.Messages) - 1

		// Start streaming
		resultChan, errChan := streamOpenAI(app.workingSession.Messages[:lastIdx])

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
			app.workingSession.Messages[lastIdx].Content = accumulated
			renderMessages(app.chatHistory, app.workingSession.Messages)
		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				app.workingSession.Messages[lastIdx].Content = fmt.Sprintf("Error: %v", err)
				renderMessages(app.chatHistory, app.workingSession.Messages)
				app.textInputArea.SetDisabled(false)
				if err := saveWorkingSession(app.sessionFile, app.workingSession); err != nil {
					panic(err)
				}
				return
			}
		}
	}
}

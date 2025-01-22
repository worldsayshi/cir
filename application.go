package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CirApplication struct {
	app           *tview.Application
	chatHistory   *tview.TextView
	textInputArea *tview.TextArea
	contextBar    tview.Primitive
	messages      []Message
	contextFiles  []string
}

func NewCirApplication() *CirApplication {
	app := tview.NewApplication()
	newPrimitive := func(text string) tview.Primitive {
		p := tview.NewTextView()
		p.
			SetBorder(true).
			SetTitle(text)
		return p
	}

	messages, err := loadMessages()
	if err != nil {
		panic(err)
	}

	chatHistory := newPrimitive("History").(*tview.TextView)

	renderMessages(chatHistory, messages)
	contextBar := newPrimitive("Context")
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
		app:           app,
		chatHistory:   chatHistory,
		textInputArea: textInputArea,
		contextBar:    contextBar,
		messages:      messages,
	}

	textInputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			cirApp.handleChatSubmit()
			return nil
		}
		if event.Key() == tcell.KeyCtrlO {
			cirApp.addContextFile()
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

func (app *CirApplication) addContextFile() {
	cmd := "find . -type f -not -path '*/.*' | fzf-tmux -h"
	out, err := exec.Command(
		"bash", "-c", cmd,
		// "find", ".", "-type", "f", "-not", "-path", "'*/.*'", "|", "fzf-tmux", "-h"
	).CombinedOutput() // "50%", "--preview", "'bat --color=always {}'")
	if err != nil {
		log.Println(err)
	}
	//file, err := os.Open(string(out))
	app.contextFiles = append(app.contextFiles, string(out))
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

func (app *CirApplication) handleChatSubmit() {
	text := app.textInputArea.GetText()
	if text != "" {
		app.messages = append(app.messages, Message{Role: "user", Content: text})
		renderMessages(app.chatHistory, app.messages)
		app.textInputArea.SetText("", true)
		if err := saveMessages(app.messages); err != nil {
			panic(err)
		}

		// Lock the text input area
		app.textInputArea.SetDisabled(true)

		// Add empty message for streaming response
		app.messages = append(app.messages, Message{Role: "system", Content: ""})
		lastIdx := len(app.messages) - 1

		// Start streaming
		resultChan, errChan := streamOpenAI(app.messages[:lastIdx])

		// Create a goroutine to handle streaming updates
		go app.handleStreamResponse(resultChan, errChan)
	}
}

func (app *CirApplication) handleStreamResponse(resultChan chan string, errChan chan error) {
	accumulated := ""
	lastIdx := len((*app).messages) - 1
	for {
		select {
		case chunk, ok := <-resultChan:
			if !ok {
				// Stream completed
				if err := saveMessages(app.messages); err != nil {
					panic(err)
				}
				app.textInputArea.SetDisabled(false)
				return
			}
			accumulated += chunk
			app.messages[lastIdx].Content = accumulated
			renderMessages(app.chatHistory, app.messages)
		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				app.messages[lastIdx].Content = fmt.Sprintf("Error: %v", err)
				renderMessages(app.chatHistory, app.messages)
				app.textInputArea.SetDisabled(false)
				if err := saveMessages((*app).messages); err != nil {
					panic(err)
				}
				return
			}
		}
	}
}

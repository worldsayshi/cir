package main

import (
	"fmt"
	"log"
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
	messages      []string
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
	chatHistory.SetText(strings.Join(messages, "\n\n---\n"))
	chatHistory.ScrollToEnd()
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
			addContextFile()
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
	return nil
}

func (app *CirApplication) handleChatSubmit() {
	text := app.textInputArea.GetText()
	if text != "" {
		(*app).messages = append(app.messages, text)
		(*app).chatHistory.SetText(strings.Join((*app).messages, "\n\n---\n"))
		(*app).textInputArea.SetText("", true)
		if err := saveMessages(app.messages); err != nil {
			panic(err)
		}

		// Lock the text input area
		(*app).textInputArea.SetDisabled(true)

		// Add empty message for streaming response
		(*app).messages = append((*app).messages, "")
		lastIdx := len((*app).messages) - 1

		// Start streaming
		resultChan, errChan := streamOpenAI((app.messages)[:lastIdx])

		// Create a goroutine to handle streaming updates
		go handleStreamResponse(app, resultChan, errChan)
	}
}

func handleStreamResponse(app *CirApplication, resultChan chan string, errChan chan error) {
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
			app.messages[lastIdx] = accumulated
			app.chatHistory.SetText(strings.Join((*app).messages, "\n\n---\n"))
			app.chatHistory.ScrollToEnd()
		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				app.messages[lastIdx] = fmt.Sprintf("Error: %v", err)
				app.chatHistory.SetText(strings.Join((*app).messages, "\n\n---\n"))
				app.textInputArea.SetDisabled(false)
				if err := saveMessages((*app).messages); err != nil {
					panic(err)
				}
				return
			}
		}
	}
}

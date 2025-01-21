package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setupLogging() (f *os.File, err error) {
	f, err = os.OpenFile("go-coder.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return nil, err
	}

	log.SetOutput(f)
	return f, nil
}

func main() {
	var app *tview.Application
	logfile, err := setupLogging()
	if err != nil {
		panic(err)
	}
	defer logfile.Close()
	log.Println("Starting Go Coder")

	messages, err := loadMessages()
	if err != nil {
		panic(err)
	}

	newPrimitive := func(text string) tview.Primitive {
		p := tview.NewTextView()
		p.
			SetBorder(true).
			SetTitle(text)
		return p
	}

	chatHistory := newPrimitive("History").(*tview.TextView)
	chatHistory.SetText(strings.Join(messages, "\n\n---\n"))
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

	textInputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			text := textInputArea.GetText()
			if text != "" {
				messages = append(messages, text)
				chatHistory.SetText(strings.Join(messages, "\n\n---\n"))
				textInputArea.SetText("", true)
				if err := saveMessages(messages); err != nil {
					panic(err)
				}

				// Lock the text input area
				textInputArea.SetDisabled(true)

				// Add empty message for streaming response
				messages = append(messages, "")
				lastIdx := len(messages) - 1

				// Start streaming
				resultChan, errChan := streamOpenAI(messages[:lastIdx])

				// Create a goroutine to handle streaming updates
				go func() {
					accumulated := ""
					for {
						select {
						case chunk, ok := <-resultChan:
							log.Printf("Received chunk: %v", chunk)
							if !ok {
								// Stream completed
								log.Println("Stream completed")
								textInputArea.SetDisabled(false)
								return
							}
							accumulated += chunk
							messages[lastIdx] = accumulated
							chatHistory.SetText(strings.Join(messages, "\n\n---\n"))
							chatHistory.ScrollToEnd()
						case err := <-errChan:
							log.Printf("Error: %v", err)
							if err != nil {
								messages[lastIdx] = fmt.Sprintf("Error: %v", err)
								chatHistory.SetText(strings.Join(messages, "\n\n---\n"))
								textInputArea.SetDisabled(false)
								return
							}
						}
					}
				}()
			}
			return nil
		}
		return event
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatHistory, 0, 5, false).
		AddItem(contextBar, 0, 1, false).
		AddItem(textInputArea, 0, 2, true)
	app = tview.NewApplication()
	if err := app.
		SetRoot(flex, true).
		SetFocus(textInputArea).Run(); err != nil {
		panic(err)
	}
}

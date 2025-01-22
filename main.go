package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

func handleStreamResponse(resultChan chan string, errChan chan error, messages *[]string, lastIdx int, chatHistory *tview.TextView, textInputArea *tview.TextArea) {
	accumulated := ""
	for {
		select {
		case chunk, ok := <-resultChan:
			if !ok {
				// Stream completed
				if err := saveMessages(*messages); err != nil {
					panic(err)
				}
				textInputArea.SetDisabled(false)
				return
			}
			accumulated += chunk
			(*messages)[lastIdx] = accumulated
			chatHistory.SetText(strings.Join(*messages, "\n\n---\n"))
			chatHistory.ScrollToEnd()
		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				(*messages)[lastIdx] = fmt.Sprintf("Error: %v", err)
				chatHistory.SetText(strings.Join(*messages, "\n\n---\n"))
				textInputArea.SetDisabled(false)
				if err := saveMessages(*messages); err != nil {
					panic(err)
				}
				return
			}
		}
	}
}

func handleChatSubmit(messages *[]string, chatHistory *tview.TextView, textInputArea *tview.TextArea) {
	text := textInputArea.GetText()
	if text != "" {
		*messages = append(*messages, text)
		chatHistory.SetText(strings.Join(*messages, "\n\n---\n"))
		textInputArea.SetText("", true)
		if err := saveMessages(*messages); err != nil {
			panic(err)
		}

		// Lock the text input area
		textInputArea.SetDisabled(true)

		// Add empty message for streaming response
		*messages = append(*messages, "")
		lastIdx := len(*messages) - 1

		// Start streaming
		resultChan, errChan := streamOpenAI((*messages)[:lastIdx])

		// Create a goroutine to handle streaming updates
		go handleStreamResponse(resultChan, errChan, messages, lastIdx, chatHistory, textInputArea)
	}
}

func addContextFile() {
	cmd := "find . -type f -not -path '*/.*' | fzf-tmux -h"
	out, err := exec.Command(
		"bash", "-c", cmd,
		// "find", ".", "-type", "f", "-not", "-path", "'*/.*'", "|", "fzf-tmux", "-h"
	).CombinedOutput() // "50%", "--preview", "'bat --color=always {}'")
	//out, err := cmd.CombinedOutput()
	log.Printf("Output1: %s", out)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Output: %s", out)
}

func main() {
	var app *tview.Application
	logfile, err := setupLogging()
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

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

	textInputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			handleChatSubmit(&messages, chatHistory, textInputArea)
			return nil
		}
		if event.Key() == tcell.KeyCtrlO {
			addContextFile()
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

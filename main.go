package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const sessionFile = "./.go-coder/default-session.yaml"

func loadMessages() ([]string, error) {
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		log.Println("Session file not found, creating a new one")
		sessionFileDir := filepath.Dir(sessionFile)
		os.MkdirAll(sessionFileDir, 0755)
		return []string{}, os.WriteFile(sessionFile, []byte("[]"), 0644)
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, err
	}

	var messages []string
	if err := yaml.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func saveMessages(messages []string) error {
	data, err := yaml.Marshal(messages)
	if err != nil {
		return err
	}

	return os.WriteFile(sessionFile, data, 0644)
}

func main() {
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
			}
			return nil
		}
		return event
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatHistory, 0, 5, false).
		AddItem(contextBar, 0, 1, false).
		AddItem(textInputArea, 0, 2, true)
	if err := tview.NewApplication().
		SetRoot(flex, true).
		SetFocus(textInputArea).Run(); err != nil {
		panic(err)
	}
}

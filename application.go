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
	"github.com/worldsayshi/cir/internal/state"
	"github.com/worldsayshi/cir/internal/storage"
	"github.com/worldsayshi/cir/internal/types"
)

// CirApplication manages UI components and application lifecycle
type CirApplication struct {
	*tview.Application
	appState      *state.AppState
	chatHistory   *components.ChatHistory
	inputArea     *components.InputArea
	contextBar    *components.ContextBar
	helpPopup     *components.HelpPopup
	rootContainer *tview.Flex
	keyMappings   []types.KeyMapping
	pages         *tview.Pages
}

func NewCirApplication(sessionFile string) *CirApplication {
	// Initialize state
	workingSession, err := storage.LoadWorkingSession(sessionFile)
	if err != nil {
		log.Println("Error loading session from file:", sessionFile)
		panic(fmt.Sprintf("Error loading session from file: %v\n%v", sessionFile, err))
	}

	// Initialize centralized AppState
	appState := state.NewAppState(workingSession, sessionFile)

	// Set up automatic persistence
	appState.AddPersistenceSubscriber()

	// Initialize UI components
	chatHistory := components.NewChatHistory(nil) // Will be populated during render
	contextBar := components.NewContextBar(nil)   // Will be populated during render
	inputArea := components.NewInputArea()
	pages := tview.NewPages()

	cirApp := &CirApplication{
		Application:   tview.NewApplication(),
		appState:      appState,
		chatHistory:   chatHistory,
		inputArea:     inputArea,
		contextBar:    contextBar,
		rootContainer: tview.NewFlex().SetDirection(tview.FlexRow),
		pages:         pages,
	}

	// Define key mappings
	cirApp.keyMappings = createKeyMappings(cirApp)
	cirApp.helpPopup = components.NewHelpPopup(
		cirApp.keyMappings,
		cirApp.pages.HasPage,
		func(name string, item tview.Primitive, resize, visible bool) {
			cirApp.pages.AddPage(name, item, resize, visible)
		},
		func(name string) {
			cirApp.pages.RemovePage(name)
		},
	)

	// Set up UI event handlers
	inputArea.SetChangedFunc(func() {
		appState.Update(func(data *state.StateData) {
			data.WorkingSession.InputText = inputArea.GetText()
		})
	})

	inputArea.SetSubmitFunc(cirApp.handleChatSubmit)

	// Subscribe to state changes for UI updates
	appState.Subscribe(cirApp.render)

	// Setup keyboard handlers
	cirApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// If the help popup is open, any key press should close it
		if cirApp.pages.HasPage("help") {
			cirApp.pages.RemovePage("help")
			return nil
		}

		// Special case for the '?' key
		if event.Key() == tcell.KeyRune && event.Rune() == '?' {
			cirApp.helpPopup.ShowHelpPopup()
			return nil
		}

		// Check for specific key mappings (Tab, Ctrl+E, etc.)
		for _, mapping := range cirApp.keyMappings {
			if event.Key() == mapping.Key {
				log.Println("Key pressed:",
					components.GetKeyName(event.Key()), "Action:", mapping.Description,
				)
				mapping.Action()
				return nil
			}
		}

		// Let all other keys pass through to the focused primitive
		return event
	})

	// Initial render of the UI
	cirApp.render()

	return cirApp
}

func createKeyMappings(cirApp *CirApplication) []types.KeyMapping {
	return []types.KeyMapping{
		{
			Key:         tcell.KeyTab,
			Description: "Cycle focus forward",
			Action: func() {
				focusableElements := []tview.Primitive{cirApp.chatHistory, cirApp.inputArea}
				cirApp.cycleFocus(focusableElements, false)
			},
		},
		{
			Key:         tcell.KeyBacktab,
			Description: "Cycle focus backward",
			Action: func() {
				focusableElements := []tview.Primitive{cirApp.chatHistory, cirApp.inputArea}
				cirApp.cycleFocus(focusableElements, true)
			},
		},
		{
			Key:         tcell.KeyCtrlE,
			Description: "Open session file",
			Action: func() {
				cirApp.openSessionFile()
			},
		},
		{
			Key:         tcell.KeyCtrlY,
			Description: "Edit context files",
			Action: func() {
				cirApp.editContextFiles()
			},
		},
	}
}

// render updates all UI components based on current state
// without triggering further state updates
func (cirApp *CirApplication) render() {
	// Get all state data we need without modifying state
	workingSession := cirApp.appState.GetWorkingSession()
	isProcessing := cirApp.appState.IsCurrentlyProcessing()

	// Update UI components with current state
	cirApp.chatHistory.Render(workingSession.Messages)
	cirApp.contextBar.Render(workingSession.WorkingFiles)

	// Only update if the text actually changed to prevent loops
	if cirApp.inputArea.GetText() != workingSession.InputText {
		cirApp.inputArea.SetText(workingSession.InputText, false)
	}

	cirApp.inputArea.SetDisabled(isProcessing)

	// Set up the layout if it hasn't been done yet
	if cirApp.rootContainer.GetItemCount() == 0 {
		cirApp.rootContainer.
			AddItem(cirApp.chatHistory, 0, 5, false).
			AddItem(cirApp.contextBar, 0, 1, false).
			AddItem(cirApp.inputArea, 0, 2, true)

		// Add the main UI to the pages
		cirApp.pages.AddPage("main", cirApp.rootContainer, true, true)
		cirApp.SetRoot(cirApp.pages, true)
		cirApp.SetFocus(cirApp.inputArea)
	}
}

// Pages returns the application's pages component
func (cirApp *CirApplication) Pages() *tview.Pages {
	return cirApp.pages
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

// openSessionFile loads a new session file and updates the app state
func (cirApp *CirApplication) openSessionFile() {
	sessionFindingCommand := `(dir=$(pwd); while [ "$dir" != "/" ]; do find "$dir" -maxdepth 1 \( -name "*.yaml" -o -name "*.yml" \) -exec grep -l "^kind: WorkingSession" {} \; 2>/dev/null; if [ -d "$dir/.cir" ]; then find "$dir/.cir" -maxdepth 1 \( -name "*.yaml" -o -name "*.yml" \) -exec grep -l "^kind: WorkingSession" {} \; 2>/dev/null; fi; dir=$(dirname "$dir"); done)`
	tmuxSessionFindingCommand := sessionFindingCommand + ` | fzf-tmux -h -m`
	out, err := exec.Command(
		"bash", "-c", tmuxSessionFindingCommand,
	).CombinedOutput()
	if err != nil {
		log.Println(err)
		return
	}

	filePath := strings.TrimSpace(string(out))
	if filePath == "" {
		return
	}

	// Load the selected session file
	newWorkingSession, err := storage.LoadWorkingSession(filePath)
	if err != nil {
		log.Printf("Error loading session from file: %v\n%v", filePath, err)
		return
	}

	// Update state with new session
	cirApp.appState.Update(func(data *state.StateData) {
		data.WorkingSession = newWorkingSession
		data.SessionFile = filePath
	})
}

func (cirApp *CirApplication) editContextFiles() {
	cmd := "find . -type f -not -path '*/.*' | fzf-tmux -h -m"
	out, err := exec.Command(
		"bash", "-c", cmd,
	).CombinedOutput()
	if err != nil {
		log.Println(err)
		return
	}

	contextFiles := strings.Split(string(out), "\n")
	// filter out empty strings
	selectedWorkingFiles := []types.WorkingFile{}
	for _, f := range contextFiles {
		if f != "" {
			selectedWorkingFiles = append(selectedWorkingFiles, types.WorkingFile{Path: f})
		}
	}

	cirApp.appState.Update(func(data *state.StateData) {
		data.WorkingSession.WorkingFiles = selectedWorkingFiles
	})
}

func (cirApp *CirApplication) Run() error {
	return cirApp.Application.Run()
}

func (cirApp *CirApplication) handleChatSubmit(text string) {
	if text == "" {
		return
	}

	cirApp.appState.Update(func(data *state.StateData) {
		// Initialize with system message if needed
		if len(data.WorkingSession.Messages) == 0 {
			data.WorkingSession.Messages = append(data.WorkingSession.Messages,
				createSystemMessage())
		}

		// Get checksums without risking deadlocks, all within the update function
		filesToSubmit := getFilesToSubmitWithChecksums(data.WorkingSession.WorkingFiles)
		userMessage := prepareUserMessage(filesToSubmit, text)

		// Add user message
		data.WorkingSession.Messages = append(
			data.WorkingSession.Messages, userMessage,
		)

		// Update file checksums
		updatedWorkingFiles := make([]types.WorkingFile, len(data.WorkingSession.WorkingFiles))
		copy(updatedWorkingFiles, data.WorkingSession.WorkingFiles)

		for i, wf := range updatedWorkingFiles {
			for _, wfSubmit := range filesToSubmit {
				if wf.Path == wfSubmit.Path {
					updatedWorkingFiles[i] = wfSubmit
				}
			}
		}
		data.WorkingSession.WorkingFiles = updatedWorkingFiles

		// Clear input and set processing state
		data.WorkingSession.InputText = ""
		data.IsProcessing = true

		// Add empty message for streaming response
		data.WorkingSession.Messages = append(
			data.WorkingSession.Messages,
			types.Message{
				AiServiceMessage: types.AiServiceMessage{Role: "assistant", Content: ""},
			},
		)
	})

	// Get service messages for API call
	serviceMessages := getServiceMessages(cirApp.appState.GetWorkingSession().Messages)

	// Start streaming
	resultChan, errChan := streamOpenAI(serviceMessages)

	// Create a goroutine to handle streaming updates
	go cirApp.handleStreamResponse(resultChan, errChan)
}

func (cirApp *CirApplication) handleStreamResponse(resultChan chan string, errChan chan error) {
	accumulated := ""

	for {
		select {
		case chunk, ok := <-resultChan:
			if !ok {
				// Stream completed
				cirApp.appState.Update(func(data *state.StateData) {
					data.IsProcessing = false
				})
				return
			}

			accumulated += chunk

			cirApp.appState.Update(func(data *state.StateData) {
				lastIdx := len(data.WorkingSession.Messages) - 1
				if lastIdx >= 0 {
					data.WorkingSession.Messages[lastIdx].AiServiceMessage.Content = accumulated
				}
			})

		case err := <-errChan:
			log.Printf("Error: %v", err)
			if err != nil {
				cirApp.appState.Update(func(data *state.StateData) {
					lastIdx := len(data.WorkingSession.Messages) - 1
					if lastIdx >= 0 {
						data.WorkingSession.Messages[lastIdx].Content = fmt.Sprintf("Error: %v", err)
					}
					data.IsProcessing = false
				})
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

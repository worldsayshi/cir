package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/worldsayshi/cir/internal/types"
)

// Model represents the UI state and components
type Model struct {
	// State
	state         *AppState
	activeElement ActiveElement
	help          helpModel
	showHelp      bool

	// UI Components
	chatHistory viewport.Model
	inputArea   textarea.Model
	statusBar   StatusBarModel

	// Window size
	width  int
	height int

	// Application callbacks for actions that require application logic
	appCallbacks AppCallbacks
}

// AppState holds all application state
type AppState struct {
	WorkingSession *types.WorkingSession
	SessionFile    string
	IsProcessing   bool
}

// ActiveElement tracks which UI element currently has focus
type ActiveElement int

const (
	ChatHistoryElement ActiveElement = iota
	InputAreaElement
)

// Application callback functions that will be injected
type AppCallbacks struct {
	SubmitMessage    func(text string) tea.Cmd
	OpenSessionFile  func() tea.Cmd
	EditContextFiles func() tea.Cmd
}

// NewModel creates a new UI model
func NewModel(state *AppState) Model {
	// Create the input textarea
	ta := textarea.New()
	ta.Placeholder = "Write here..."
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	ta.SetHeight(4)
	ta.Focus() // Make sure the textarea is focused from the start

	// Set up the chat history viewport
	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Create help model
	help := newHelpModel()

	// Create status bar
	statusBar := NewStatusBarModel()

	return Model{
		state:         state,
		activeElement: InputAreaElement,
		chatHistory:   vp,
		inputArea:     ta,
		statusBar:     statusBar,
		help:          help,
		showHelp:      false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// UpdateChatHistory updates the chat history content
func (m *Model) UpdateChatHistory() {
	if m.state == nil || m.state.WorkingSession == nil {
		return
	}

	// Use the current width for proper text wrapping
	width := m.chatHistory.Width
	content := formatChatHistory(m.state.WorkingSession.Messages, width)
	m.chatHistory.SetContent(content)
	m.chatHistory.GotoBottom()
}

// UpdateContextBar updates the context bar with working files
func (m *Model) UpdateContextBar() {
	if m.state == nil || m.state.WorkingSession == nil {
		return
	}

	m.statusBar.UpdateWorkingFiles(m.state.WorkingSession.WorkingFiles)
}

// GetInputText returns the current input text
func (m *Model) GetInputText() string {
	return m.inputArea.Value()
}

// SetInputText sets the input text
func (m *Model) SetInputText(text string) {
	m.inputArea.SetValue(text)
}

// SetAppCallbacks sets the callback functions for interacting with the main application
func (m *Model) SetAppCallbacks(callbacks AppCallbacks) {
	m.appCallbacks = callbacks
}

// Update handles UI events and state changes
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// First check if help is being displayed - if so, any key closes it
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// Handle global key shortcuts
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Toggle focus between chat history and input area
			if m.activeElement == ChatHistoryElement {
				m.activeElement = InputAreaElement
				m.inputArea.Focus()
			} else {
				m.activeElement = ChatHistoryElement
				m.inputArea.Blur()
			}
			return m, nil

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "ctrl+s":
			if m.activeElement == InputAreaElement && !m.state.IsProcessing && m.appCallbacks.SubmitMessage != nil {
				text := m.inputArea.Value()
				if text != "" {
					// Clear the input area immediately for better UX
					m.inputArea.Reset()
					return m, m.appCallbacks.SubmitMessage(text)
				}
			}
			return m, nil

		case "ctrl+e":
			// Open session file
			if m.appCallbacks.OpenSessionFile != nil {
				return m, m.appCallbacks.OpenSessionFile()
			}
			return m, nil

		case "ctrl+y":
			// Edit context files
			if m.appCallbacks.EditContextFiles != nil {
				return m, m.appCallbacks.EditContextFiles()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		// Handle window resize
		m.width = msg.Width
		m.height = msg.Height

		// Resize chat history viewport
		m.chatHistory.Width = msg.Width - 2
		m.chatHistory.Height = msg.Height - 10

		// Resize input area
		m.inputArea.SetWidth(msg.Width - 2)

		return m, nil

	case SubmitMessageMsg:
		// Handle message submission - UI already updated by the app
		return m, nil

	case StreamResponseChunkMsg:
		// Update UI with streamed response chunk
		m.UpdateChatHistory()
		return m, nil

	case StreamResponseDoneMsg:
		// Handle completion of response streaming
		m.state.IsProcessing = false
		m.UpdateChatHistory()
		return m, nil

	case StreamResponseErrorMsg:
		// Handle error in response streaming
		m.state.IsProcessing = false
		m.UpdateChatHistory()
		m.statusBar.UpdateStatus(fmt.Sprintf("Error: %v", msg.Err))
		return m, nil

	case SessionUpdatedMsg:
		// Update UI after session changes
		m.UpdateChatHistory()
		m.UpdateContextBar()
		return m, nil
	}

	// Handle active element updates
	if m.activeElement == ChatHistoryElement {
		newViewport, cmd := m.chatHistory.Update(msg)
		m.chatHistory = newViewport
		cmds = append(cmds, cmd)
	} else {
		newTextarea, cmd := m.inputArea.Update(msg)
		m.inputArea = newTextarea
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if m.showHelp {
		return m.help.View()
	}

	// Build the main layout
	chatView := m.chatHistory.View()
	inputView := m.inputArea.View()
	statusView := m.statusBar.View(m.width)

	// Apply styles based on focus
	if m.activeElement == ChatHistoryElement {
		chatView = focusedBorderStyle.Render(chatView)
		inputView = blurredBorderStyle.Render(inputView)
	} else {
		chatView = blurredBorderStyle.Render(chatView)
		inputView = focusedBorderStyle.Render(inputView)
	}

	// Combine the views
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Cir - Chat Interface"),
		chatView,
		statusView,
		inputView,
	)
}

// Custom message types for our application
type SubmitMessageMsg struct {
	Text string
}

type StreamResponseChunkMsg struct {
	Chunk string
}

type StreamResponseDoneMsg struct{}

type StreamResponseErrorMsg struct {
	Err error
}

type SessionUpdatedMsg struct{}

// Commands
func submitMessage(text string) tea.Cmd {
	return func() tea.Msg {
		return SubmitMessageMsg{Text: text}
	}
}

func openSessionFile() tea.Cmd {
	return func() tea.Msg {
		// This will be implemented in the application logic
		return nil
	}
}

func editContextFiles() tea.Cmd {
	return func() tea.Msg {
		// This will be implemented in the application logic
		return nil
	}
}

// Styling
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	focusedBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62"))

	blurredBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))
)

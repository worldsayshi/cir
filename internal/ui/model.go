package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/worldsayshi/cir/internal/types"
)

// Layout constants
const (
	// Common layout constants
	TitleHeight     = 1
	StatusBarHeight = 3
	BorderHeight    = 2 // Account for borders (1 top + 1 bottom)

	// Mode-specific layout constants
	NormalModeExtraSpace = 6  // Total space used in normal mode (title + borders + status + mode indicator)
	InsertModeExtraSpace = 12 // Total space used in insert mode (includes input area + extra padding)
)

// Model represents the UI state and components
type Model struct {
	// State
	state            *AppState
	activeElement    ActiveElement
	help             helpModel
	showHelp         bool
	insertMode       bool // Flag to track if we're in insert mode
	selectedMsgIndex int  // Currently selected message in chat history

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
	ta.Blur() // Initially blurred since we're starting in normal mode

	// Set up the chat history viewport
	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Create help model
	help := newHelpModel()

	// Create status bar
	statusBar := NewStatusBarModel()

	return Model{
		state:            state,
		activeElement:    ChatHistoryElement, // Start with focus on chat history
		chatHistory:      vp,
		inputArea:        ta,
		statusBar:        statusBar,
		help:             help,
		showHelp:         false,
		insertMode:       false, // Start in normal mode (not insert mode)
		selectedMsgIndex: -1,    // Initialize with no selected message
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
	content := formatChatHistory(m.state.WorkingSession.Messages, width, m.selectedMsgIndex)
	m.chatHistory.SetContent(content)

	// If no message is selected and we have messages, select the latest one
	if m.selectedMsgIndex == -1 && len(m.state.WorkingSession.Messages) > 0 {
		m.selectedMsgIndex = len(m.state.WorkingSession.Messages) - 1
		// Re-render with the selection
		content = formatChatHistory(m.state.WorkingSession.Messages, width, m.selectedMsgIndex)
		m.chatHistory.SetContent(content)
	}

	m.chatHistory.GotoBottom()
}

// scrollToSelectedMessage scrolls the viewport to make the selected message visible
func (m *Model) scrollToSelectedMessage() {
	if m.selectedMsgIndex < 0 || len(m.state.WorkingSession.Messages) == 0 {
		return
	}

	// Calculate approximate position of selected message
	content := m.chatHistory.View()
	lines := strings.Split(content, "\n")

	// Find the position by looking for the message indicator
	messageStartLine := 0
	currentMessage := -1 // Start at -1 so we properly count the first message as 0

	log.Println("lines:", len(lines))

	// SOMETHING IS WRONG HERE
	// We need to find the line number of the selected message
	for i, line := range lines {
		// Look for the indicator prefix or role headers at the start of lines
		trimmedLine := strings.TrimSpace(line)
		log.Println(trimmedLine, strings.HasPrefix(trimmedLine, "Assistant"))
		// Check if this is a header line (either with the indicator or as a regular header)
		if strings.HasPrefix(trimmedLine, "User") ||
			strings.HasPrefix(trimmedLine, "Assistant") ||
			strings.HasPrefix(trimmedLine, "System") {

			log.Printf("Found message header at line %d: %s", i, trimmedLine)
			// If we found a header, we're at a new message
			currentMessage += 1
			log.Printf("Current message index: %d", currentMessage)

			log.Println("", i, m.selectedMsgIndex)
			// If this is our target message, record its position
			if currentMessage == m.selectedMsgIndex {
				messageStartLine = i
				break
			}
		}

	}

	log.Printf("Selected message index: %d, Start line: %d", m.selectedMsgIndex, messageStartLine)

	// Scroll to position with some context
	if messageStartLine > 0 {
		// Adjust the viewport to show the selected message with some context
		targetLine := max(0, messageStartLine-2) // 2 lines above for context
		m.chatHistory.SetYOffset(targetLine)
	}
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// First check if help is being displayed - if so, any key closes it
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// Different key handling based on insert mode
		if m.insertMode {
			// In insert mode
			switch msg.String() {
			case "esc":
				// Exit insert mode
				m.insertMode = false
				m.activeElement = ChatHistoryElement
				m.inputArea.Blur()

				// Adjust chat history height for normal mode
				m.chatHistory.Height = m.height - NormalModeExtraSpace
				m.UpdateChatHistory()

				return m, nil

			case "ctrl+s":
				if !m.state.IsProcessing && m.appCallbacks.SubmitMessage != nil {
					text := m.inputArea.Value()
					if text != "" {
						// Clear the input area immediately for better UX
						m.inputArea.Reset()
						// Also exit insert mode after sending message
						m.insertMode = false
						m.activeElement = ChatHistoryElement

						// Adjust chat history height for normal mode
						m.chatHistory.Height = m.height - NormalModeExtraSpace

						return m, m.appCallbacks.SubmitMessage(text)
					}
				}
				return m, nil

			default:
				// Handle all other keys in textarea
				newTextarea, cmd := m.inputArea.Update(msg)
				m.inputArea = newTextarea
				return m, cmd
			}
		} else {
			// Normal mode (not insert mode)
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit

			case "i":
				// Enter insert mode
				m.insertMode = true
				m.activeElement = InputAreaElement
				m.inputArea.Focus()

				// Adjust chat history height for insert mode
				m.chatHistory.Height = m.height - InsertModeExtraSpace
				m.UpdateChatHistory()

				return m, nil

			case "?":
				m.showHelp = !m.showHelp
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

			case "h":
				// Previous message
				if !m.insertMode && len(m.state.WorkingSession.Messages) > 0 {
					// Go to previous message if not at the first one
					if m.selectedMsgIndex > 0 {
						m.selectedMsgIndex--
						m.UpdateChatHistory()
						m.scrollToSelectedMessage() // Scroll to make the selected message visible
					}
				}
				return m, nil

			case "l":
				// Next message
				if !m.insertMode && len(m.state.WorkingSession.Messages) > 0 {
					// Go to next message if not at the last one
					if m.selectedMsgIndex < len(m.state.WorkingSession.Messages)-1 {
						m.selectedMsgIndex++
						m.UpdateChatHistory()
						m.scrollToSelectedMessage() // Scroll to make the selected message visible
					}
				}
				return m, nil

			default:
				// In normal mode, handle viewport navigation
				newViewport, cmd := m.chatHistory.Update(msg)
				m.chatHistory = newViewport
				return m, cmd
			}
		}

	case tea.WindowSizeMsg:
		// Handle window resize
		m.width = msg.Width
		m.height = msg.Height

		// Resize chat history viewport - make it larger when not in insert mode
		m.chatHistory.Width = msg.Width - BorderHeight
		if m.insertMode {
			m.chatHistory.Height = msg.Height - InsertModeExtraSpace
		} else {
			m.chatHistory.Height = msg.Height - NormalModeExtraSpace
		}

		// Resize input area
		m.inputArea.SetWidth(msg.Width - BorderHeight)

		// Re-render chat history with the new width to ensure text wrapping
		m.UpdateChatHistory()

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

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.showHelp {
		return m.help.View()
	}

	// Build the main layout
	chatView := m.chatHistory.View()
	statusView := m.statusBar.View(m.width)

	// In normal mode, don't show the input area
	if !m.insertMode {
		// When not in insert mode, just show chat history and status bar
		return lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("Cir - Chat Interface"),
			focusedBorderStyle.Render(chatView), // Always focused in normal mode
			statusView,
			modeIndicatorStyle.Render("NORMAL"), // Show mode indicator
		)
	}

	// In insert mode, show input area too
	inputView := m.inputArea.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Cir - Chat Interface"),
		blurredBorderStyle.Render(chatView), // Always blurred in insert mode
		statusView,
		focusedBorderStyle.Render(inputView), // Always focused in insert mode
		modeIndicatorStyle.Render("INSERT"),  // Show mode indicator
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

	modeIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#444444")).
				Padding(0, 1)
)

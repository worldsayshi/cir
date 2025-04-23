package state

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/worldsayshi/cir/internal/storage"
	"github.com/worldsayshi/cir/internal/types"
)

// AppState holds the application state and manages subscriptions
type AppState struct {
	mu             sync.RWMutex
	workingSession *types.WorkingSession
	sessionFile    string
	isProcessing   bool
	subscribers    []func()
	updating       atomic.Bool // Flag to prevent recursive updates
}

// NewAppState creates a new AppState with the given working session
func NewAppState(workingSession *types.WorkingSession, sessionFile string) *AppState {
	return &AppState{
		workingSession: workingSession,
		sessionFile:    sessionFile,
		isProcessing:   false,
		subscribers:    []func(){},
	}
}

// Update applies a state change function in a thread-safe manner
// and notifies all subscribers of the change
func (a *AppState) Update(updateFunc func(*AppState)) {
	// If we're already inside an update, don't trigger subscribers again
	isAlreadyUpdating := a.updating.Swap(true)

	// Apply the update under lock
	a.mu.Lock()
	updateFunc(a)
	a.mu.Unlock()

	// Only notify subscribers if this is the outermost update call
	if !isAlreadyUpdating {
		// Call subscribers
		for _, subscriber := range a.subscribers {
			subscriber()
		}

		// Reset the updating flag
		a.updating.Store(false)
	}
}

// GetWorkingSession returns the working session
func (a *AppState) GetWorkingSession() *types.WorkingSession {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.workingSession
}

// SetWorkingSession sets the working session
func (a *AppState) SetWorkingSession(ws *types.WorkingSession) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workingSession = ws
}

// GetSessionFile returns the session file path
func (a *AppState) GetSessionFile() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessionFile
}

// SetSessionFile sets the session file path
func (a *AppState) SetSessionFile(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sessionFile = path
}

// IsCurrentlyProcessing returns whether the application is currently processing a request
func (a *AppState) IsCurrentlyProcessing() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isProcessing
}

// SetIsProcessing sets the processing state
func (a *AppState) SetIsProcessing(isProcessing bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.isProcessing = isProcessing
}

// GetMessages returns the messages from the working session
func (a *AppState) GetMessages() []types.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.workingSession.Messages
}

// AddMessage adds a message to the working session
func (a *AppState) AddMessage(message types.Message) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workingSession.Messages = append(a.workingSession.Messages, message)
}

// UpdateLastMessage updates the last message in the working session
func (a *AppState) UpdateLastMessage(content string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	lastIdx := len(a.workingSession.Messages) - 1
	if lastIdx >= 0 {
		a.workingSession.Messages[lastIdx].AiServiceMessage.Content = content
	}
}

// SetInputText sets the input text in the working session
func (a *AppState) SetInputText(text string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workingSession.InputText = text
}

// GetInputText gets the input text from the working session
func (a *AppState) GetInputText() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.workingSession.InputText
}

// SetWorkingFiles sets the working files in the working session
func (a *AppState) SetWorkingFiles(files []types.WorkingFile) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workingSession.WorkingFiles = files
}

// GetWorkingFiles gets the working files from the working session
func (a *AppState) GetWorkingFiles() []types.WorkingFile {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.workingSession.WorkingFiles
}

// Subscribe adds a subscriber function that will be called when state changes
func (a *AppState) Subscribe(subscriber func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.subscribers = append(a.subscribers, subscriber)
}

// AddPersistenceSubscriber adds a subscriber that persists the state to disk
// whenever state changes occur
func (a *AppState) AddPersistenceSubscriber() {
	a.Subscribe(func() {
		// Thread-safe access to state
		a.mu.RLock()
		sessionFile := a.sessionFile
		workingSession := a.workingSession
		a.mu.RUnlock()

		// Save state to disk without triggering additional updates
		err := storage.SaveWorkingSession(sessionFile, workingSession)
		if err != nil {
			log.Printf("Error saving session: %v", err)
		}
	})
}

package state

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/worldsayshi/cir/internal/storage"
	"github.com/worldsayshi/cir/internal/types"
)

// StateData holds all the application state that can be mutated
// Fields are exported so they can be directly accessed within update functions
type StateData struct {
	WorkingSession *types.WorkingSession
	SessionFile    string
	IsProcessing   bool
}

// AppState manages state and subscriptions
type AppState struct {
	mu          sync.RWMutex
	data        StateData
	subscribers []func()
	updating    atomic.Bool // Flag to prevent recursive updates
}

// NewAppState creates a new AppState with the given working session
func NewAppState(workingSession *types.WorkingSession, sessionFile string) *AppState {
	return &AppState{
		data: StateData{
			WorkingSession: workingSession,
			SessionFile:    sessionFile,
			IsProcessing:   false,
		},
		subscribers: []func(){},
	}
}

// Update applies a state change function in a thread-safe manner
// and notifies all subscribers of the change
func (a *AppState) Update(updateFunc func(*StateData)) {
	// If we're already inside an update, don't trigger subscribers again
	isAlreadyUpdating := a.updating.Swap(true)

	// Apply the update
	a.mu.Lock()
	updateFunc(&a.data)
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

// GetWorkingSession returns the current working session (thread-safe)
func (a *AppState) GetWorkingSession() *types.WorkingSession {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data.WorkingSession
}

// GetSessionFile returns the session file path (thread-safe)
func (a *AppState) GetSessionFile() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data.SessionFile
}

// IsCurrentlyProcessing returns whether a task is in progress (thread-safe)
func (a *AppState) IsCurrentlyProcessing() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data.IsProcessing
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
		sessionFile := a.data.SessionFile
		workingSession := a.data.WorkingSession
		a.mu.RUnlock()

		// Save state to disk without triggering additional updates
		err := storage.SaveWorkingSession(sessionFile, workingSession)
		if err != nil {
			log.Printf("Error saving session: %v", err)
		}
	})
}

// Might cause deadlock if not careful
// This should only be used for testing or debugging
func (a *AppState) UnsafeGetState() *StateData {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return &a.data
}

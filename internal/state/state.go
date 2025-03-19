package state

import (
	"sync"

	"github.com/worldsayshi/cir/internal/types"
)

/*
	Optional example of how state can be managed!
*/

// AppState holds all application state
type AppState struct {
	WorkingSession *types.WorkingSession
	SessionFile    string
	IsProcessing   bool

	mu          sync.RWMutex
	subscribers []func()
}

// NewAppState creates a new application state
func NewAppState(workingSession *types.WorkingSession, sessionFile string) *AppState {
	return &AppState{
		WorkingSession: workingSession,
		SessionFile:    sessionFile,
		IsProcessing:   false,
		subscribers:    make([]func(), 0),
	}
}

// Update modifies the state with the provided update function and notifies subscribers
func (s *AppState) Update(updateFunc func(*AppState)) {
	s.mu.Lock()
	updateFunc(s)
	s.mu.Unlock()

	s.notifySubscribers()
}

// Subscribe adds a listener that will be called when state changes
func (s *AppState) Subscribe(callback func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscribers = append(s.subscribers, callback)
}

// Unsubscribe removes a listener
func (s *AppState) Unsubscribe(callback func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, cb := range s.subscribers {
		if &cb == &callback {
			s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
			break
		}
	}
}

// notifySubscribers calls all subscribed listeners
func (s *AppState) notifySubscribers() {
	s.mu.RLock()
	subscribers := make([]func(), len(s.subscribers))
	copy(subscribers, s.subscribers)
	s.mu.RUnlock()

	for _, callback := range subscribers {
		callback()
	}
}

// GetWorkingSession returns the working session (thread-safe)
func (s *AppState) GetWorkingSession() *types.WorkingSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.WorkingSession
}

// GetSessionFile returns the session file path (thread-safe)
func (s *AppState) GetSessionFile() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SessionFile
}

// IsCurrentlyProcessing checks if the application is processing a request (thread-safe)
func (s *AppState) IsCurrentlyProcessing() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IsProcessing
}

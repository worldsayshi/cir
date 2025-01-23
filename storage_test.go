package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadWorkingSession(t *testing.T) {
	tmpDir := t.TempDir()
	testSessionFile := filepath.Join(tmpDir, "test-session.yaml")
	os.Setenv("TEST_SESSION_FILE", testSessionFile)

	// Prepare a session to save
	session := &WorkingSession{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
			{Role: "system", Content: "Hi there!"},
		},
		WorkingFiles: []WorkingFile{
			{Path: "file1.txt"},
			{Path: "file2.txt"},
		},
	}

	fmt.Println("Session:", session)

	// Try saving
	if err := saveWorkingSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Now load it back
	loadedSession, err := loadWorkingSession()
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Assert that loadedSession matches the original session
	if len(loadedSession.Messages) != len(session.Messages) ||
		len(loadedSession.WorkingFiles) != len(session.WorkingFiles) {
		t.Fatalf("Loaded session does not match saved session")
	}

	for i, msg := range loadedSession.Messages {
		if msg.Role != session.Messages[i].Role || msg.Content != session.Messages[i].Content {
			t.Fatalf("Loaded message does not match saved message")
		}
	}

	for i, file := range loadedSession.WorkingFiles {
		if file.Path != session.WorkingFiles[i].Path {
			t.Fatalf("Loaded file does not match saved file")
		}
	}
}

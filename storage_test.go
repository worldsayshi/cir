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

	// Try saving
	if err := saveWorkingSession(testSessionFile, session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Now load it back
	loadedSession, err := loadWorkingSession(testSessionFile)
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

func TestLoadNewWorkingSession(t *testing.T) {
	tmpDir := t.TempDir()
	testSessionFile := filepath.Join(tmpDir, "test-session.yaml")

	// Ensure no session file exists
	if _, err := os.Stat(testSessionFile); !os.IsNotExist(err) {
		t.Fatalf("Session file should not exist")
	}

	// Load session, which should create a new one
	workingSession, err := loadWorkingSession(testSessionFile)
	if err != nil {
		t.Fatalf("Failed to load new session: %v", err)
	}

	// Assert that the session is empty
	if len(workingSession.Messages) != 0 || len(workingSession.WorkingFiles) != 0 {
		fmt.Println("Session:", workingSession)
		t.Fatalf("New session should be empty")
	}

	// Ensure the session file was created
	if _, err := os.Stat(testSessionFile); os.IsNotExist(err) {
		t.Fatalf("Session file should have been created")
	}

	// Load the session again to ensure it is still empty
	loadedSession, err := loadWorkingSession(testSessionFile)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Assert that the loaded session is still empty
	if len(loadedSession.Messages) != 0 || len(loadedSession.WorkingFiles) != 0 {
		t.Fatalf("Loaded session should be empty")
	}
}

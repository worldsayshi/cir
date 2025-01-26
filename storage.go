package main

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func loadWorkingSession(sessionFile string) (*WorkingSession, error) {
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		log.Println("Session file not found, creating a new one at", sessionFile)
		sessionFileDir := filepath.Dir(sessionFile)
		os.MkdirAll(sessionFileDir, 0755)
		return &WorkingSession{}, saveWorkingSession(sessionFile, &WorkingSession{})
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, err
	}

	var workingSession WorkingSession
	if err := yaml.Unmarshal(data, &workingSession); err != nil {
		return nil, err
	}

	return &workingSession, nil
}

func saveWorkingSession(sessionFile string, workingSession *WorkingSession) error {
	// Strip out the file content from the working session before saving
	for i := range workingSession.WorkingFiles {
		workingSession.WorkingFiles[i].FileContent = nil
	}
	data, err := yaml.Marshal(workingSession)
	if err != nil {
		return err
	}

	return os.WriteFile(sessionFile, data, 0644)
}

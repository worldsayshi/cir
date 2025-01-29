package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/worldsayshi/cir/internal/types"
	"gopkg.in/yaml.v2"
)

func loadWorkingSession(sessionFile string) (*types.WorkingSession, error) {
	log.Println("Loading working session from", sessionFile)
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		log.Println("Session file not found, creating a new one at", sessionFile)
		sessionFileDir := filepath.Dir(sessionFile)
		os.MkdirAll(sessionFileDir, 0755)
		return &types.WorkingSession{}, saveWorkingSession(sessionFile, &types.WorkingSession{})
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, err
	}

	workingSession, err := types.UnmarshalWorkingSession(data)
	if err != nil {
		panic(err)
	}
	return workingSession, nil
}

func saveWorkingSession(sessionFile string, workingSession *types.WorkingSession) error {
	apiVersion := types.CurrentApiVersion
	workingSession.ApiVersion = &apiVersion
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

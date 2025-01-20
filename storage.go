package main

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const sessionFile = "./.go-coder/default-session.yaml"

func loadMessages() ([]string, error) {
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		log.Println("Session file not found, creating a new one at", sessionFile)
		sessionFileDir := filepath.Dir(sessionFile)
		os.MkdirAll(sessionFileDir, 0755)
		return []string{}, os.WriteFile(sessionFile, []byte("[]"), 0644)
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, err
	}

	var messages []string
	if err := yaml.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func saveMessages(messages []string) error {
	data, err := yaml.Marshal(messages)
	if err != nil {
		return err
	}

	return os.WriteFile(sessionFile, data, 0644)
}

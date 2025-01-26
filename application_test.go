package main

import (
	"os"
	"testing"
)

func TestPrepareUserMessage(t *testing.T) {
	// Create a temporary session file
	tmpSessionfile, err := os.CreateTemp("", "session-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpSessionfile.Name())

	// Write initial session data to the file
	initialSessionData := `
messages:
  - role: "user"
    content: "Hello"
working_files:
  - path: "testfile.txt"
    last_submitted_checksum: null
`
	if _, err := tmpSessionfile.Write([]byte(initialSessionData)); err != nil {
		t.Fatal(err)
	}
	if err := tmpSessionfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create a test file
	testFilePath := "testfile.txt"
	testFileContent := "This is a test file."
	if err := os.WriteFile(testFilePath, []byte(testFileContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFilePath)

	// Initialize CirApplication
	app := NewCirApplication(tmpSessionfile.Name())

	// Prepare user message
	question := "What is the content of the test file?"
	userMessage := app.prepareUserMessage(question)

	// Check if the user message contains the expected content
	expectedContent := `<context file="testfile.txt">
This is a test file.
</context>
<question>
What is the content of the test file?
</question>`
	if userMessage != expectedContent {
		t.Errorf("Expected user message to be:\n%s\nBut got:\n%s", expectedContent, userMessage)
	}
}

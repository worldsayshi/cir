package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalWorkingSessionV1(t *testing.T) {
	yamlData := `
messages:
- role: user
  content: |-
    <context file="./test.txt">
    Here's the not so secret message: New York

    </context>
    <question>
    Hi! Can you read the message from the test file?
    </question>
  question: Hi! Can you read the message from the test file?
- role: system
  content: 'Hello! Yes, the message from the test file is: "New York".'
working_files:
- path: ./test.txt
`

	workingSession, err := UnmarshalWorkingSession([]byte(yamlData))
	assert.NoError(t, err)
	assert.NotNil(t, workingSession)
	assert.Equal(t, 2, len(workingSession.Messages))
	messageToBeContained := "New York"
	assert.True(t, strings.Contains(workingSession.Messages[0].Content, messageToBeContained),
		"Expected message %q to contain %q",
		workingSession.Messages[0].Content,
		messageToBeContained,
	)
	assert.Equal(t, 1, len(workingSession.WorkingFiles))
	assert.Equal(t, "./test.txt", workingSession.WorkingFiles[0].Path)
}

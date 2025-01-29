package v2

import (
	"github.com/worldsayshi/cir/internal/types/versionedtype"
)

type AiServiceMessage struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
}

type WorkingFile struct {
	Path                  string  `json:"path" yaml:"path"`
	LastSubmittedChecksum *string `json:"last_submitted_checksum,omitempty" yaml:"last_submitted_checksum,omitempty"`
	FileContent           []byte  `json:"-" yaml:"-"` // Don't serialize this field
}

type Message struct {
	AiServiceMessage     `json:"aiServiceMessage,omitempty" yaml:"aiServiceMessage,omitempty"`
	Question             string        `json:"question,omitempty" yaml:"question,omitempty"`
	IncludedWorkingFiles []WorkingFile `json:"included_working_files,omitempty" yaml:"included_working_files,omitempty"`
}

type WorkingSession struct {
	*versionedtype.ApiVersion `json:"apiVersion" yaml:"apiVersion"`
	Messages                  []Message     `json:"messages" yaml:"messages"`
	WorkingFiles              []WorkingFile `json:"working_files" yaml:"working_files"`
	InputText                 string        `json:"input_text" yaml:"input_text"`
}

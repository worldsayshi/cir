package v1

type WorkingFile struct {
	Path        string `json:"path" yaml:"path"`
	FileContent []byte `json:"-" yaml:"-"` // Don't serialize this field
}

type Message struct {
	Role                 string        `json:"role" yaml:"role"`
	Content              string        `json:"content" yaml:"content"`
	Question             string        `json:"question,omitempty" yaml:"question,omitempty"`
	IncludedWorkingFiles []WorkingFile `json:"included_working_files,omitempty" yaml:"included_working_files,omitempty"`
}

type WorkingSession struct {
	Messages     []Message     `json:"messages" yaml:"messages"`
	WorkingFiles []WorkingFile `json:"working_files" yaml:"working_files"`
}

package types

import (
	v1 "github.com/worldsayshi/cir/internal/types/v1"
	v2 "github.com/worldsayshi/cir/internal/types/v2"
	// "gopkg.in/yaml.v2"
)

type ApiVersion string

const (
	V1 ApiVersion = "v1"
	V2 ApiVersion = "v2"
)

type VersionedType struct {
	*ApiVersion `json:"apiVersion" yaml:"apiVersion"`
}

type (
	Message          = v2.Message
	WorkingFile      = v2.WorkingFile
	WorkingSession   = v2.WorkingSession
	AiServiceMessage = v2.AiServiceMessage
)

// TODO data is a yaml file. Look for the key `apiVersion` and unmarshal using the corresponding versioned type. I.e. apiVersion == 'v1' means that the yaml should be unmarshaled using v1.WorkingSession. If apiVersion is missing assume that it is v1.
func UnmarshalWorkingSession(data []byte) (workingSession WorkingSession, err error) {
	// err = yaml.Unmarshal(data, &workingSession)
}

// TODO WorkingSession v1 -> v2
func ConvertWorkingSessionV1ToV2(workingSessionV1 v1.WorkingSession) (workingSessionV2 v2.WorkingSession, err error) {
}

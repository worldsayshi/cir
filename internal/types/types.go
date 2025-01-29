package types

import (
	v2 "github.com/worldsayshi/cir/internal/types/v2"
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

// func UnmarshalWorkingSession(data []byte) (workingSession WorkingSession, err error) {
// 	err = yaml.Unmarshal(data, &workingSession)
// 	return
// }

package types

import (
	v2 "github.com/worldsayshi/cir/internal/types/v2"
	"github.com/worldsayshi/cir/internal/types/versionedtype"
)

type (
	Message          = v2.Message
	WorkingFile      = v2.WorkingFile
	WorkingSession   = v2.WorkingSession
	AiServiceMessage = v2.AiServiceMessage
)

const (
	CurrentApiVersion = versionedtype.V2
)

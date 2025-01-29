package types

import (
	"gopkg.in/yaml.v2"

	v1 "github.com/worldsayshi/cir/internal/types/v1"
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

func UnmarshalWorkingSession(data []byte) (workingSession *v2.WorkingSession, err error) {
	var vt versionedtype.VersionedType
	if err = yaml.Unmarshal(data, &vt); err != nil {
		return nil, err
	}

	// If it's an unknown version or missing, assume v1
	var apiVersion versionedtype.ApiVersion
	if vt.ApiVersion == nil {
		apiVersion = versionedtype.V1
	} else {
		apiVersion = *vt.ApiVersion
	}

	switch apiVersion {
	case versionedtype.V1:
		var workingSessionV1 v1.WorkingSession
		if err = yaml.UnmarshalStrict(data, &workingSessionV1); err != nil {
			return nil, err
		}
		return ConvertWorkingSessionV1ToV2(&workingSessionV1)
	case versionedtype.V2:
		if err = yaml.Unmarshal(data, &workingSession); err != nil {
			return nil, err
		}
		return workingSession, nil
	default:
		panic("Unknown API version: " + apiVersion)
	}
}

func ConvertWorkingSessionV1ToV2(workingSessionV1 *v1.WorkingSession) (workingSessionV2 *v2.WorkingSession, err error) {
	var messagesV2 []v2.Message
	for _, msgV1 := range workingSessionV1.Messages {
		wf := convertWorkingFilesV1ToV2(&msgV1.IncludedWorkingFiles)
		msgV2 := v2.Message{
			Question:             msgV1.Question,
			IncludedWorkingFiles: *wf,
		}
		messagesV2 = append(messagesV2, msgV2)
	}

	workingSessionV2 = &v2.WorkingSession{
		Messages: messagesV2,
		WorkingFiles: *convertWorkingFilesV1ToV2(
			&workingSessionV1.WorkingFiles,
		),
	}

	return workingSessionV2, nil
}

func convertWorkingFilesV1ToV2(workingFilesV1 *[]v1.WorkingFile) *[]v2.WorkingFile {
	var workingFilesV2 []v2.WorkingFile
	for _, wfV1 := range *workingFilesV1 {
		wfV2 := &v2.WorkingFile{
			Path:        wfV1.Path,
			FileContent: wfV1.FileContent,
		}
		workingFilesV2 = append(workingFilesV2, *wfV2)
	}
	return &workingFilesV2
}

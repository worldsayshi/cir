package versionedtype

type ApiVersion string

const (
	V1 ApiVersion = "v1"
	V2 ApiVersion = "v2"
)

type VersionedType struct {
	*ApiVersion `json:"apiVersion" yaml:"apiVersion"`
}

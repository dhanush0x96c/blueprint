package resolver

import "io/fs"

// SourceType represents the type of a template source.
type SourceType string

const (
	SourceTypeBuiltin SourceType = "builtin"
	SourceTypeUser    SourceType = "user"
)

// Source represents a template source.
type Source struct {
	Name       string
	Type       SourceType
	Filesystem fs.FS
}

package resolver

import (
	"io/fs"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// TemplateRef represents a reference to a template.
type TemplateRef struct {
	Name string
	Type template.Type
}

// ResolvedTemplate represents a resolved template.
type ResolvedTemplate struct {
	FS   fs.FS
	Path string
}

// Resolver resolves a template reference.
type Resolver interface {
	Resolve(ref TemplateRef) (*ResolvedTemplate, error)
}

// Discoverer discovers templates available from a source.
type Discoverer interface {
	Discover() (map[string]*template.Template, error)
	Exists(path string) bool
}

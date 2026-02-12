package app

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

// Resolver is the interface for resolving templates.
type Resolver interface {
	Resolve(ctx *Context, ref TemplateRef) (*ResolvedTemplate, error)
}

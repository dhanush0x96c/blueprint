package template

import "io/fs"

// TemplateRef represents a reference to a template.
type TemplateRef struct {
	Name string
	Type Type
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

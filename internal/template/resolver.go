package template

import "io/fs"

// TemplateRef represents a reference to a template.
type TemplateRef struct {
	Name string
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

// DiscoverOptions contains options for template discovery.
type DiscoverOptions struct {
	Type         Type
	Tags         []string
	IgnoreErrors bool
}

// Discoverer discovers templates available from a source.
type Discoverer interface {
	Discover(opts DiscoverOptions) (map[string]*Template, error)
	Exists(name string) bool
}

// Loader loads a template from a filesystem.
type Loader interface {
	Load(fsys fs.FS, pth string) (*Template, error)
}

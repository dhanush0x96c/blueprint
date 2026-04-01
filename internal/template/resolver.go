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
	Discover(opts DiscoverOptions) (map[string]*Metadata, error)
	Exists(name string) bool
}

// MetadataLoader loads template metadata from a filesystem.
type MetadataLoader interface {
	LoadMetadata(fsys fs.FS, pth string) (*Metadata, error)
}

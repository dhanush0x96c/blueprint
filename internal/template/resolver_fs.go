package template

import (
	"io/fs"
	"path"
)

// ResolveFromFS resolves a template from a given file system.
func ResolveFromFS(rootFS fs.FS, ref TemplateRef) (*ResolvedTemplate, error) {
	templatePath := path.Join(ref.Type.Folder(), ref.Name)

	_, err := fs.Stat(rootFS, templatePath)
	if err != nil {
		return nil, &TemplateNotFoundError{Name: ref.Name}
	}

	return &ResolvedTemplate{
		FS:   rootFS,
		Path: templatePath,
	}, nil
}

// FSResolver resolves templates from a file system.
type FSResolver struct {
	rootFS fs.FS
}

// NewFSResolver creates a resolver backed by the provided file system.
func NewFSResolver(rootFS fs.FS) *FSResolver {
	return &FSResolver{rootFS: rootFS}
}

// Resolve resolves templates from the configured file system.
func (r *FSResolver) Resolve(ref TemplateRef) (*ResolvedTemplate, error) {
	return ResolveFromFS(r.rootFS, ref)
}

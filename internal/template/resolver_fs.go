package template

import (
	"fmt"
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

// Discover finds all templates and returns them keyed by template directory path.
func (r *FSResolver) Discover() (map[string]*Template, error) {
	templates := make(map[string]*Template)
	loader := NewLoader(r.rootFS)

	err := fs.WalkDir(r.rootFS, ".", func(pth string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != FileName {
			return nil
		}

		tmpl, err := loader.Load(pth)
		if err != nil {
			return nil
		}

		templates[path.Dir(pth)] = tmpl
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to discover templates: %w", err)
	}

	return templates, nil
}

// Exists checks if a template exists at the given path.
func (r *FSResolver) Exists(templatePath string) bool {
	_, err := fs.Stat(r.rootFS, resolveTemplatePath(templatePath))
	return err == nil
}

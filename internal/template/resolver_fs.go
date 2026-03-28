package template

import (
	"fmt"
	"io/fs"
	"path"
)

// FSResolver resolves templates from a file system.
type FSResolver struct {
	rootFS fs.FS
	loader *FileLoader
}

// NewFSResolver creates a resolver backed by the provided file system.
func NewFSResolver(rootFS fs.FS) *FSResolver {
	return &FSResolver{rootFS: rootFS, loader: NewLoader(rootFS)}
}

// Resolve resolves templates from the configured file system.
func (r *FSResolver) Resolve(ref TemplateRef) (*ResolvedTemplate, error) {
	templates, err := r.Discover()

	if err != nil {
		return nil, err
	}

	for path, tmpl := range templates {
		if tmpl.Name == ref.Name && tmpl.Type == ref.Type {
			return &ResolvedTemplate{
				Path: path,
				FS:   r.rootFS,
			}, nil
		}
	}

	return nil, &TemplateNotFoundError{Name: ref.Name}
}

// Discover finds all templates and returns them keyed by template directory path.
func (r *FSResolver) Discover() (map[string]*Template, error) {
	templates := make(map[string]*Template)

	err := fs.WalkDir(r.rootFS, ".", func(pth string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != FileName {
			return nil
		}

		tmpl, err := r.loader.Load(pth)
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

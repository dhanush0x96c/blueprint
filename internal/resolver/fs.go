package resolver

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// FSResolver resolves templates from a file system.
type FSResolver struct {
	rootFS fs.FS
	loader *template.FileLoader
}

// NewFSResolver creates a resolver backed by the provided file system.
func NewFSResolver(rootFS fs.FS) *FSResolver {
	return &FSResolver{rootFS: rootFS, loader: template.NewLoader()}
}

// Resolve resolves templates from the configured file system.
func (r *FSResolver) Resolve(ref template.TemplateRef) (*template.ResolvedTemplate, error) {
	templates, err := r.Discover()
	if err != nil {
		return nil, err
	}

	for path, tmpl := range templates {
		if tmpl.Name == ref.Name {
			return &template.ResolvedTemplate{
				Path: path,
				FS:   r.rootFS,
			}, nil
		}
	}

	return nil, &template.TemplateNotFoundError{Name: ref.Name}
}

// Discover finds all templates and returns them keyed by template directory path.
func (r *FSResolver) Discover() (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	err := fs.WalkDir(r.rootFS, ".", func(pth string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != template.FileName {
			return nil
		}

		tmpl, err := r.loader.Load(r.rootFS, pth)
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

func resolveTemplatePath(pth string) string {
	if path.Base(pth) == template.FileName {
		return pth
	}

	return path.Join(pth, template.FileName)
}

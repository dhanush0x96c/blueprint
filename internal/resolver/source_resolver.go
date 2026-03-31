package resolver

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// SourceResolver resolves templates from a source.
type SourceResolver struct {
	source Source
	loader *template.FileLoader
}

// NewSourceResolver creates a resolver backed by the provided source.
func NewSourceResolver(source Source) *SourceResolver {
	return &SourceResolver{source: source, loader: template.NewLoader()}
}

// Resolve resolves templates from the configured source.
func (r *SourceResolver) Resolve(ref template.TemplateRef) (*template.ResolvedTemplate, error) {
	templates, err := r.Discover()
	if err != nil {
		return nil, err
	}

	for pth, tmpl := range templates {
		if tmpl.Name == ref.Name {
			return &template.ResolvedTemplate{
				Path: pth,
				FS:   r.source.Filesystem,
			}, nil
		}
	}

	return nil, &template.TemplateNotFoundError{Name: ref.Name}
}

// Discover finds all templates and returns them keyed by template directory path.
func (r *SourceResolver) Discover() (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	err := fs.WalkDir(r.source.Filesystem, ".", func(pth string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != template.FileName {
			return nil
		}

		tmpl, err := r.loader.Load(r.source.Filesystem, pth)
		if err != nil {
			return nil
		}

		templates[path.Dir(pth)] = tmpl
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to discover templates from source %s: %w", r.source.Name, err)
	}

	return templates, nil
}

// Exists checks if a template exists at the given path.
func (r *SourceResolver) Exists(templatePath string) bool {
	_, err := fs.Stat(r.source.Filesystem, resolveTemplatePath(templatePath))
	return err == nil
}

func resolveTemplatePath(pth string) string {
	if path.Base(pth) == template.FileName {
		return pth
	}

	return path.Join(pth, template.FileName)
}

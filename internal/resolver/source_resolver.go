package resolver

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

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
	templates, err := r.Discover(template.DiscoverOptions{IgnoreErrors: true})
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
func (r *SourceResolver) Discover(opts template.DiscoverOptions) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	err := fs.WalkDir(r.source.Filesystem, ".", func(pth string, d fs.DirEntry, err error) error {
		if err != nil {
			if opts.IgnoreErrors {
				return nil
			}
			return err
		}

		if d.IsDir() || d.Name() != template.FileName {
			return nil
		}

		tmpl, err := r.loader.Load(r.source.Filesystem, pth)
		if err != nil {
			if opts.IgnoreErrors {
				return nil
			}
			return err
		}

		if opts.Type != "" && tmpl.Type != opts.Type {
			return nil
		}

		if len(opts.Tags) > 0 && !matchesAnyTag(tmpl, opts.Tags) {
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

// matchesAnyTag returns true if the template has at least one of the filter tags.
func matchesAnyTag(tmpl *template.Template, filterTags []string) bool {
	if len(tmpl.Tags) == 0 {
		return false
	}

	tagSet := make(map[string]struct{}, len(tmpl.Tags))
	for _, t := range tmpl.Tags {
		tagSet[strings.ToLower(t)] = struct{}{}
	}

	for _, ft := range filterTags {
		if _, ok := tagSet[strings.ToLower(ft)]; ok {
			return true
		}
	}

	return false
}

// Exists checks if a template exists with the given name.
func (r *SourceResolver) Exists(name string) bool {
	templates, err := r.Discover(template.DiscoverOptions{IgnoreErrors: true})
	if err != nil {
		return false
	}

	for _, tmpl := range templates {
		if tmpl.Name == name {
			return true
		}
	}

	return false
}

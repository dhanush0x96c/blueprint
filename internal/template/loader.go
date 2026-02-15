package template

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	FileName = "template.yaml"
)

// FileLoader handles loading templates from the filesystem
type FileLoader struct {
	fs       fs.FS
	validate *validator.Validate
}

// NewLoader creates a new template loader with the given base directory
func NewLoader(fs fs.FS) *FileLoader {
	return &FileLoader{
		fs:       fs,
		validate: validator.New(),
	}
}

// Load loads a template from the filesystem
//
// The path may refer to either a template.yaml file or a directory
// containing one. In the latter case, "<dir>/template.yaml" is used.
//
// The loaded template is validated, and all file source paths
// (Template.Files[].Src) are resolved relative to the template directory.
func (l *FileLoader) Load(path string) (*Template, error) {
	templatePath := l.resolveTemplatePath(path)

	data, err := fs.ReadFile(l.fs, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	tmplDir := filepath.Dir(templatePath)

	if err := l.validate.Struct(&tmpl); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	for i := range tmpl.Files {
		tmpl.Files[i].Src = filepath.Join(
			tmplDir, tmpl.Files[i].Src)
	}

	return &tmpl, nil
}

// LoadFromDir loads a template from a directory containing template.yaml
func (l *FileLoader) LoadFromDir(dir string) (*Template, error) {
	templatePath := filepath.Join(dir, FileName)
	return l.Load(templatePath)
}

// Discover finds all available templates in the base directory
// Returns a map of template path -> template name
func (l *FileLoader) Discover() (map[string]string, error) {
	templates := make(map[string]string)

	err := fs.WalkDir(l.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if d.Name() != FileName {
			return nil
		}

		tmpl, err := l.Load(path)
		if err != nil {
			return nil
		}

		relDir := filepath.Dir(path)

		templates[relDir] = tmpl.Name
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover templates: %w", err)
	}

	return templates, nil
}

// DiscoverByType finds all templates of a specific type
func (l *FileLoader) DiscoverByType(templateType Type) (map[string]string, error) {
	allTemplates, err := l.Discover()
	if err != nil {
		return nil, err
	}

	filtered := make(map[string]string)
	for path, name := range allTemplates {
		tmpl, err := l.LoadFromDir(path)
		if err != nil {
			continue
		}

		if tmpl.Type == templateType {
			filtered[path] = name
		}
	}

	return filtered, nil
}

// DiscoverAll finds all templates and returns the full Template structs.
// If filterType is non-empty, only templates of that type are returned.
func (l *FileLoader) DiscoverAll(filterType Type) ([]*Template, error) {
	var templates []*Template

	err := fs.WalkDir(l.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != FileName {
			return nil
		}

		tmpl, err := l.Load(path)
		if err != nil {
			return nil
		}

		if filterType != "" && tmpl.Type != filterType {
			return nil
		}

		templates = append(templates, tmpl)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover templates: %w", err)
	}

	return templates, nil
}

// Exists checks if a template exists at the given path
func (l *FileLoader) Exists(path string) bool {
	templatePath := l.resolveTemplatePath(path)
	_, err := fs.Stat(l.fs, templatePath)
	return err == nil
}

// resolveTemplatePath resolves a template path to an absolute path
func (l *FileLoader) resolveTemplatePath(path string) string {
	if filepath.Base(path) == FileName {
		return path
	}

	return filepath.Join(path, FileName)
}

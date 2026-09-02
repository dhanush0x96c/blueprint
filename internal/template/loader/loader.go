// Package loader handles loading and parsing template manifest files.
package loader

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/template/validator"
	"gopkg.in/yaml.v3"
)

// FileName is the default manifest filename for templates.
const (
	FileName = "template.yaml"
)

// LoadedTemplate represents a template along with its source information
type LoadedTemplate struct {
	Template *template.Template
	FS       fs.FS
	Path     string
}

// Loader handles loading templates from the filesystem
type Loader struct {
	validate *validator.Validator
}

// NewLoader creates a new template loader.
func NewLoader() *Loader {
	return &Loader{
		validate: validator.NewValidator(),
	}
}

// Load loads a template from the given filesystem.
//
// The path may refer to either a template.yaml file or a directory
// containing one. In the latter case, "<dir>/template.yaml" is used.
//
// The loaded template is validated.
func (l *Loader) Load(fsys fs.FS, pth string) (*LoadedTemplate, error) {
	templatePath := resolveTemplatePath(pth)

	data, err := fs.ReadFile(fsys, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var tmpl template.Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	if err := l.validate.Validate(&tmpl); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	return &LoadedTemplate{
		Template: &tmpl,
		FS:       fsys,
		Path:     path.Dir(templatePath),
	}, nil
}

// LoadMetadata loads template metadata from the given filesystem.
func (l *Loader) LoadMetadata(fsys fs.FS, pth string) (*template.Metadata, error) {
	templatePath := resolveTemplatePath(pth)

	data, err := fs.ReadFile(fsys, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var meta template.Metadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	if err := l.validate.ValidateMetadata(&meta); err != nil {
		return nil, fmt.Errorf("metadata validation failed: %w", err)
	}

	return &meta, nil
}

// resolveTemplatePath resolves a template path to a template manifest path.
func resolveTemplatePath(pth string) string {
	if path.Base(pth) == FileName {
		return pth
	}

	return path.Join(pth, FileName)
}

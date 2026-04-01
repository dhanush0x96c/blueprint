package template

import (
	"fmt"
	"io/fs"
	"path"

	"gopkg.in/yaml.v3"
)

const (
	FileName = "template.yaml"
)

// LoadedTemplate represents a template along with its source information
type LoadedTemplate struct {
	Template *Template
	FS       fs.FS
	Path     string
}

// Loader handles loading templates from the filesystem
type Loader interface {
	Load(fsys fs.FS, pth string) (*LoadedTemplate, error)
}

// FileLoader handles loading templates from the filesystem
type FileLoader struct {
	validate *Validator
}

// NewLoader creates a new template loader.
func NewLoader() *FileLoader {
	return &FileLoader{
		validate: NewValidator(),
	}
}

// Load loads a template from the given filesystem.
//
// The path may refer to either a template.yaml file or a directory
// containing one. In the latter case, "<dir>/template.yaml" is used.
//
// The loaded template is validated.
func (l *FileLoader) Load(fsys fs.FS, pth string) (*LoadedTemplate, error) {
	templatePath := resolveTemplatePath(pth)

	data, err := fs.ReadFile(fsys, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var tmpl Template
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
func (l *FileLoader) LoadMetadata(fsys fs.FS, pth string) (*Metadata, error) {
	templatePath := resolveTemplatePath(pth)

	data, err := fs.ReadFile(fsys, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var meta Metadata
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

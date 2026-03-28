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

// FileLoader handles loading templates from the filesystem
type FileLoader struct {
	fs       fs.FS
	validate *Validator
}

// NewLoader creates a new template loader with the given base directory
func NewLoader(fs fs.FS) *FileLoader {
	return &FileLoader{
		fs:       fs,
		validate: NewValidator(),
	}
}

// Load loads a template from the filesystem
//
// The path may refer to either a template.yaml file or a directory
// containing one. In the latter case, "<dir>/template.yaml" is used.
//
// The loaded template is validated, and all file source paths
// (Template.Files[].Src) are resolved relative to the template directory.
func (l *FileLoader) Load(pth string) (*Template, error) {
	templatePath := resolveTemplatePath(pth)

	data, err := fs.ReadFile(l.fs, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	tmplDir := path.Dir(templatePath)

	if err := l.validate.Validate(&tmpl); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	for i := range tmpl.Files {
		tmpl.Files[i].Src = path.Join(
			tmplDir, tmpl.Files[i].Src)
	}

	return &tmpl, nil
}

// LoadFromDir loads a template from a directory containing template.yaml
func (l *FileLoader) LoadFromDir(dir string) (*Template, error) {
	templatePath := path.Join(dir, FileName)
	return l.Load(templatePath)
}

// resolveTemplatePath resolves a template path to a template manifest path.
func resolveTemplatePath(pth string) string {
	if path.Base(pth) == FileName {
		return pth
	}

	return path.Join(pth, FileName)
}

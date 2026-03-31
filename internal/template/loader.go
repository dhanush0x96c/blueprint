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
// The loaded template is validated, and all file source paths
// (Template.Files[].Src) are resolved relative to the template directory.
func (l *FileLoader) Load(fsys fs.FS, pth string) (*Template, error) {
	templatePath := resolveTemplatePath(pth)

	data, err := fs.ReadFile(fsys, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	tmplDir := path.Dir(templatePath)

	for i := range tmpl.Files {
		tmpl.Files[i].Src = path.Join(
			tmplDir, tmpl.Files[i].Src)
		tmpl.Files[i].FS = fsys
	}

	if err := l.validate.Validate(&tmpl); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	return &tmpl, nil
}

// resolveTemplatePath resolves a template path to a template manifest path.
func resolveTemplatePath(pth string) string {
	if path.Base(pth) == FileName {
		return pth
	}

	return path.Join(pth, FileName)
}

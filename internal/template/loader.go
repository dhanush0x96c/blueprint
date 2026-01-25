package template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	FileName = "template.yaml"
)

// FileLoader handles loading templates from the filesystem
type FileLoader struct {
	baseDir  string
	validate *validator.Validate
}

// NewLoader creates a new template loader with the given base directory
func NewLoader(baseDir string) (*FileLoader, error) {
	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path of %s: %w", baseDir, err)
	}
	return &FileLoader{
		baseDir:  abs,
		validate: validator.New(),
	}, nil
}

// Load loads a template from the given path
// The path can be either:
// - An absolute path to a template.yaml file
// - A relative path from the base directory (e.g., "projects/go-cli")
func (l *FileLoader) Load(path string) (*Template, error) {
	templatePath := l.resolveTemplatePath(path)

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	if err := l.validate.Struct(&tmpl); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
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

	err := filepath.Walk(l.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == FileName {
			tmpl, err := l.Load(path)
			if err != nil {
				return nil
			}

			relPath, err := filepath.Rel(l.baseDir, filepath.Dir(path))
			if err != nil {
				return nil
			}

			templates[relPath] = tmpl.Name
		}

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
		tmpl, err := l.LoadFromDir(filepath.Join(l.baseDir, path))
		if err != nil {
			continue
		}

		if tmpl.Type == templateType {
			filtered[path] = name
		}
	}

	return filtered, nil
}

// Exists checks if a template exists at the given path
func (l *FileLoader) Exists(path string) bool {
	templatePath := l.resolveTemplatePath(path)
	_, err := os.Stat(templatePath)
	return err == nil
}

// GetBaseDir returns the base directory of the loader
func (l *FileLoader) GetBaseDir() string {
	return l.baseDir
}

// resolveTemplatePath resolves a template path to an absolute path
func (l *FileLoader) resolveTemplatePath(path string) string {
	if filepath.IsAbs(path) && filepath.Base(path) == FileName {
		return path
	}

	if filepath.IsAbs(path) {
		return filepath.Join(path, FileName)
	}

	if filepath.Base(path) == FileName {
		return filepath.Join(l.baseDir, path)
	}

	return filepath.Join(l.baseDir, path, FileName)
}

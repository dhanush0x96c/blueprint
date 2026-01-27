package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Renderer handles rendering template files with variables
type Renderer struct {
	baseDir string
	funcMap template.FuncMap
}

// NewRenderer creates a new template renderer with the given base directory
func NewRenderer(baseDir string) *Renderer {
	r := &Renderer{
		baseDir: baseDir,
	}
	r.funcMap = r.defaultFuncMap()
	return r
}

// Render renders a template file with the given context
// The templatePath is relative to the renderer's base directory
func (r *Renderer) Render(templatePath string, ctx *Context) (string, error) {
	fullPath := filepath.Join(r.baseDir, templatePath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	return r.RenderString(string(content), ctx, templatePath)
}

// RenderString renders a template string with the given context
func (r *Renderer) RenderString(content string, ctx *Context, name string) (string, error) {
	tmpl, err := template.New(name).Funcs(r.funcMap).Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx.Variables); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderPath renders a destination path template with the given context
// This allows dynamic file paths like "{{ .package_name }}/main.go"
func (r *Renderer) RenderPath(pathTemplate string, ctx *Context) (string, error) {
	return r.RenderString(pathTemplate, ctx, "path")
}

// RenderAll renders all files from a template with the given context
// Returns a map of destination path -> rendered content
func (r *Renderer) RenderAll(tmpl *Template, ctx *Context) (map[string]string, error) {
	results := make(map[string]string)

	for _, file := range tmpl.Files {
		destPath, err := r.RenderPath(file.Dest, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to render destination path for %s: %w", file.Src, err)
		}

		content, err := r.Render(file.Src, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to render template %s: %w", file.Src, err)
		}

		results[destPath] = content
	}

	return results, nil
}

// AddFunc adds a custom function to the template function map
func (r *Renderer) AddFunc(name string, fn any) {
	r.funcMap[name] = fn
}

// defaultFuncMap returns the default set of template functions
func (r *Renderer) defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// String manipulation
		"toLower":   strings.ToLower,
		"toUpper":   strings.ToUpper,
		"title":     strings.ToTitle,
		"trim":      strings.TrimSpace,
		"trimLeft":  strings.TrimLeft,
		"trimRight": strings.TrimRight,
		"replace":   strings.ReplaceAll,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"split":     strings.Split,
		"join":      strings.Join,

		// Path manipulation
		"base":     filepath.Base,
		"dir":      filepath.Dir,
		"ext":      filepath.Ext,
		"joinPath": filepath.Join,

		// Type conversions
		"toString": toString,
		"toInt":    toInt,
		"toBool":   toBool,

		// Utility
		"default":  defaultValue,
		"empty":    isEmpty,
		"coalesce": coalesce,
	}
}

// Helper functions for template rendering

func toString(v any) string {
	return fmt.Sprintf("%v", v)
}

func toInt(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		return i, err
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "1" || val == "yes"
	case int, int64:
		return val != 0
	default:
		return false
	}
}

func defaultValue(defaultVal, val any) any {
	if isEmpty(val) {
		return defaultVal
	}
	return val
}

func isEmpty(val any) bool {
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

func coalesce(vals ...any) any {
	for _, val := range vals {
		if !isEmpty(val) {
			return val
		}
	}
	return nil
}

package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
)

// Renderer handles rendering template files with variables
type Renderer struct {
	fs      fs.FS
	funcMap template.FuncMap
}

// NewRenderer creates a new template renderer with the given base directory
func NewRenderer(fs fs.FS) *Renderer {
	r := &Renderer{fs: fs}
	r.funcMap = r.defaultFuncMap()
	return r
}

// Render renders a template file with the given context
// The templatePath is relative to the renderer's base directory
func (r *Renderer) Render(templatePath string, ctx *Context) (string, error) {
	content, err := fs.ReadFile(r.fs, templatePath)
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

// Copy reads a file and returns its content without template processing
func (r *Renderer) Copy(filePath string) (string, error) {
	content, err := fs.ReadFile(r.fs, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

// RenderAll renders all files from a template with the given context
// Returns a map of destination path -> rendered content
// Files with .tmpl extension are rendered and the extension is stripped
// Other files are copied as-is
func (r *Renderer) RenderAll(tmpl *Template, ctx *Context) (map[string]string, error) {
	results := make(map[string]string)

	for _, file := range tmpl.Files {
		if err := r.processPath(file.Src, file.Dest, ctx, results); err != nil {
			return nil, err
		}
	}

	return results, nil
}

// processPath processes a file or directory path recursively
func (r *Renderer) processPath(srcPath, destPath string, ctx *Context, results map[string]string) error {
	info, err := fs.Stat(r.fs, srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", srcPath, err)
	}

	if info.IsDir() {
		return r.processDirectory(srcPath, destPath, ctx, results)
	}

	return r.processFile(srcPath, destPath, ctx, results)
}

// processDirectory recursively processes all files in a directory
func (r *Renderer) processDirectory(srcDir, destDir string, ctx *Context, results map[string]string) error {
	entries, err := fs.ReadDir(r.fs, srcDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if err := r.processPath(srcPath, destPath, ctx, results); err != nil {
			return err
		}
	}

	return nil
}

// isTemplateFile checks if the path has a .tmpl extension
func isTemplateFile(path string) bool {
	return strings.HasSuffix(path, ".tmpl")
}

// stripTemplateExt removes the .tmpl extension from a path
func stripTemplateExt(path string) string {
	return strings.TrimSuffix(path, ".tmpl")
}

// processFile processes a single file - renders .tmpl files, copies others
func (r *Renderer) processFile(srcPath, destPath string, ctx *Context, results map[string]string) error {
	// Render destination path template
	renderedDestPath, err := r.RenderPath(destPath, ctx)
	if err != nil {
		return fmt.Errorf("failed to render destination path for %s: %w", srcPath, err)
	}

	var content string

	if isTemplateFile(srcPath) {
		renderedDestPath = stripTemplateExt(renderedDestPath)

		content, err = r.Render(srcPath, ctx)
		if err != nil {
			return err
		}
	} else {
		content, err = r.Copy(srcPath)
		if err != nil {
			return err
		}
	}

	results[renderedDestPath] = content

	return nil
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

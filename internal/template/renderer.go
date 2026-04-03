package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"text/template"
)

// Renderer handles rendering template files with variables
type Renderer struct {
	funcMap template.FuncMap
}

// NewRenderer creates a new template renderer
func NewRenderer() *Renderer {
	r := &Renderer{}
	r.funcMap = r.defaultFuncMap()
	return r
}

// Render renders a template file with the given context
func (r *Renderer) Render(fsys fs.FS, templatePath string, ctx *Context) ([]byte, error) {
	content, err := fs.ReadFile(fsys, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	return r.RenderString(string(content), ctx, templatePath)
}

// RenderString renders a template string with the given context
func (r *Renderer) RenderString(content string, ctx *Context, name string) ([]byte, error) {
	tmpl, err := template.New(name).Funcs(r.funcMap).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx.Variables); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

// RenderPath renders a destination path template with the given context
// This allows dynamic file paths like "{{ .package_name }}/main.go"
func (r *Renderer) RenderPath(pathTemplate string, ctx *Context) (string, error) {
	rendered, err := r.RenderString(pathTemplate, ctx, "path")
	if err != nil {
		return "", err
	}
	return string(rendered), nil
}

// Copy reads a file and returns its content without template processing
func (r *Renderer) Copy(fsys fs.FS, filePath string) ([]byte, error) {
	content, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return content, nil
}

// RenderAll renders all files from a template tree with the given contexts.
// It walks the tree and renders files for each node with its corresponding context.
func (r *Renderer) RenderAll(node *TemplateNode, contexts RenderContexts) (*RenderResult, error) {
	result := &RenderResult{
		Files: make(map[string][]RenderedFile),
	}
	if err := r.renderNode(node, contexts, result); err != nil {
		return nil, err
	}

	return result, nil
}

// renderNode recursively renders a node and its children.
func (r *Renderer) renderNode(node *TemplateNode, contexts RenderContexts, result *RenderResult) error {
	ctx, ok := contexts[node.ID]
	if !ok {
		return fmt.Errorf("no context found for template %s (ID: %s)", node.Template.Name, node.ID)
	}

	var nodeFiles []RenderedFile
	for _, file := range node.Template.Files {
		srcPath := path.Join(node.Path, file.Src)

		destPath, err := r.RenderPath(file.Dest, ctx)
		if err != nil {
			return fmt.Errorf("failed to render destination path for %s: %w", srcPath, err)
		}

		if err := r.processPath(node.FS, srcPath, destPath, ctx, &nodeFiles); err != nil {
			return err
		}
	}

	if len(nodeFiles) > 0 {
		result.Files[node.ID] = nodeFiles
	}

	for _, child := range node.Children {
		if err := r.renderNode(child, contexts, result); err != nil {
			return err
		}
	}

	return nil
}

// processPath processes a file or directory path recursively
func (r *Renderer) processPath(fsys fs.FS, srcPath, destPath string, ctx *Context, results *[]RenderedFile) error {
	info, err := fs.Stat(fsys, srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", srcPath, err)
	}

	if info.IsDir() {
		return r.processDirectory(fsys, srcPath, destPath, ctx, results)
	}

	return r.processFile(fsys, srcPath, destPath, ctx, results)
}

// processDirectory recursively processes all files in a directory
func (r *Renderer) processDirectory(fsys fs.FS, srcDir, destDir string, ctx *Context, results *[]RenderedFile) error {
	entries, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		srcPath := path.Join(srcDir, entry.Name())
		destPath := path.Join(destDir, entry.Name())

		if err := r.processPath(fsys, srcPath, destPath, ctx, results); err != nil {
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
func (r *Renderer) processFile(fsys fs.FS, srcPath, destPath string, ctx *Context, results *[]RenderedFile) error {
	var content []byte
	var err error

	if isTemplateFile(srcPath) {
		destPath = stripTemplateExt(destPath)

		content, err = r.Render(fsys, srcPath, ctx)
		if err != nil {
			return err
		}
	} else {
		content, err = r.Copy(fsys, srcPath)
		if err != nil {
			return err
		}
	}

	*results = append(*results, RenderedFile{
		Path:    destPath,
		Content: content,
	})

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
		"base":     path.Base,
		"dir":      path.Dir,
		"ext":      path.Ext,
		"joinPath": path.Join,

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

package template

import (
	"fmt"
	"io/fs"
)

// Engine is the unified template engine that orchestrates loading, composing, and rendering
type Engine struct {
	loader   *FileLoader
	composer *Composer
	renderer *Renderer
}

// NewEngine creates a new template engine with the given template base directory
func NewEngine(templatesFS fs.FS) *Engine {
	loader := NewLoader(templatesFS)

	composer := NewComposer(loader)
	renderer := NewRenderer(templatesFS)

	return &Engine{
		loader:   loader,
		composer: composer,
		renderer: renderer,
	}
}

// LoadTemplate loads a template from the given path
func (e *Engine) LoadTemplate(path string) (*Template, error) {
	return e.loader.Load(path)
}

// ComposeTemplate resolves all includes and returns a fully composed template
func (e *Engine) ComposeTemplate(tmpl *Template) (*Template, error) {
	return e.composer.Compose(tmpl)
}

// ComposeTemplateWithIncludes resolves includes based on user selection
func (e *Engine) ComposeTemplateWithIncludes(tmpl *Template, enabledIncludes map[string]bool) (*Template, error) {
	return e.composer.ComposeWithEnabledIncludes(tmpl, enabledIncludes)
}

// GetAllIncludes returns all includes (direct and transitive) for a template
func (e *Engine) GetAllIncludes(tmpl *Template) ([]Include, error) {
	return e.composer.GetAllIncludes(tmpl)
}

// RenderTemplate renders all files from a composed template with the given context
// Returns a map of destination path -> rendered content
func (e *Engine) RenderTemplate(tmpl *Template, ctx *Context) (map[string]string, error) {
	return e.renderer.RenderAll(tmpl, ctx)
}

// ProcessTemplate is the complete end-to-end flow: load, compose, and render
// This is a convenience method that combines the three main operations
func (e *Engine) ProcessTemplate(templatePath string, ctx *Context) (map[string]string, error) {
	// Load the template
	tmpl, err := e.LoadTemplate(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Compose (resolve includes)
	composed, err := e.ComposeTemplate(tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to compose template: %w", err)
	}

	// Render all files
	rendered, err := e.RenderTemplate(composed, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return rendered, nil
}

// ProcessTemplateWithIncludes is like ProcessTemplate but allows selective includes
func (e *Engine) ProcessTemplateWithIncludes(templatePath string, ctx *Context, enabledIncludes map[string]bool) (map[string]string, error) {
	// Load the template
	tmpl, err := e.LoadTemplate(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Compose with selected includes
	composed, err := e.ComposeTemplateWithIncludes(tmpl, enabledIncludes)
	if err != nil {
		return nil, fmt.Errorf("failed to compose template: %w", err)
	}

	// Render all files
	rendered, err := e.RenderTemplate(composed, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return rendered, nil
}

// DiscoverTemplates finds all available templates in the base directory
func (e *Engine) DiscoverTemplates() (map[string]string, error) {
	return e.loader.Discover()
}

// DiscoverTemplatesByType finds all templates of a specific type
func (e *Engine) DiscoverTemplatesByType(templateType Type) (map[string]string, error) {
	return e.loader.DiscoverByType(templateType)
}

// TemplateExists checks if a template exists at the given path
func (e *Engine) TemplateExists(path string) bool {
	return e.loader.Exists(path)
}

// GetComposedTemplate returns the fully composed template without rendering
// Useful for inspecting what variables, dependencies, and files will be generated
func (e *Engine) GetComposedTemplate(templatePath string) (*Template, error) {
	tmpl, err := e.LoadTemplate(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	composed, err := e.ComposeTemplate(tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to compose template: %w", err)
	}

	return composed, nil
}

// GetTemplateVariables returns all variables needed for a template (including from includes)
func (e *Engine) GetTemplateVariables(templatePath string) ([]Variable, error) {
	composed, err := e.GetComposedTemplate(templatePath)
	if err != nil {
		return nil, err
	}

	return composed.Variables, nil
}

// GetTemplateDependencies returns all dependencies for a template (including from includes)
func (e *Engine) GetTemplateDependencies(templatePath string) ([]string, error) {
	composed, err := e.GetComposedTemplate(templatePath)
	if err != nil {
		return nil, err
	}

	return composed.Dependencies, nil
}

// AddTemplateFunc adds a custom function to the template renderer
func (e *Engine) AddTemplateFunc(name string, fn any) {
	e.renderer.AddFunc(name, fn)
}

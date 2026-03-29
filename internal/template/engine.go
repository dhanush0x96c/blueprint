package template

import (
	"fmt"
	"io/fs"
)

// Engine is the unified template engine that orchestrates loading, composing, and rendering
type Engine struct {
	resolver Resolver
	loader   *FileLoader
	composer *Composer
	renderer *Renderer
}

// NewEngine creates a new template engine with the given resolver
func NewEngine(resolver Resolver) *Engine {
	loader := NewLoader()
	composer := NewComposer(resolver, loader)
	renderer := NewRenderer()

	return &Engine{
		resolver: resolver,
		loader:   loader,
		composer: composer,
		renderer: renderer,
	}
}

// LoadTemplate loads a template from the given reference
func (e *Engine) LoadTemplate(ref TemplateRef) (*Template, error) {
	resolved, err := e.resolver.Resolve(ref)
	if err != nil {
		return nil, err
	}
	return e.loader.Load(resolved.FS, resolved.Path)
}

// LoadTemplateByPath loads a template from a specific path on a filesystem
func (e *Engine) LoadTemplateByPath(fsys fs.FS, path string) (*Template, error) {
	return e.loader.Load(fsys, path)
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
func (e *Engine) ProcessTemplate(ref TemplateRef, ctx *Context) (map[string]string, error) {
	// Load the template
	tmpl, err := e.LoadTemplate(ref)
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
func (e *Engine) ProcessTemplateWithIncludes(ref TemplateRef, ctx *Context, enabledIncludes map[string]bool) (map[string]string, error) {
	// Load the template
	tmpl, err := e.LoadTemplate(ref)
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

// GetComposedTemplate returns the fully composed template without rendering
func (e *Engine) GetComposedTemplate(ref TemplateRef) (*Template, error) {
	tmpl, err := e.LoadTemplate(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	composed, err := e.ComposeTemplate(tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to compose template: %w", err)
	}

	return composed, nil
}

// GetTemplateVariables returns all variables needed for a template
func (e *Engine) GetTemplateVariables(ref TemplateRef) ([]Variable, error) {
	composed, err := e.GetComposedTemplate(ref)
	if err != nil {
		return nil, err
	}

	return composed.Variables, nil
}

// GetTemplateDependencies returns all dependencies for a template
func (e *Engine) GetTemplateDependencies(ref TemplateRef) ([]string, error) {
	composed, err := e.GetComposedTemplate(ref)
	if err != nil {
		return nil, err
	}

	return composed.Dependencies, nil
}

// AddTemplateFunc adds a custom function to the template renderer
func (e *Engine) AddTemplateFunc(name string, fn any) {
	e.renderer.AddFunc(name, fn)
}

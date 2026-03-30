package template

import (
	"fmt"
	"io/fs"
)

// Engine is the unified template engine that orchestrates loading, composing, and rendering
type Engine struct {
	resolver  Resolver
	loader    *FileLoader
	composer  *Composer
	renderer  *Renderer
	validator *Validator
}

// NewEngine creates a new template engine with the given resolver
func NewEngine(resolver Resolver) *Engine {
	loader := NewLoader()
	composer := NewComposer(resolver, loader)
	renderer := NewRenderer()
	validator := NewValidator()

	return &Engine{
		resolver:  resolver,
		loader:    loader,
		composer:  composer,
		renderer:  renderer,
		validator: validator,
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

// Compose resolves all includes for a template recursively and builds a TemplateNode tree.
// It calls confirm for all includes of a template to decide which ones should be loaded.
func (e *Engine) Compose(tmpl *Template, confirm ConfirmIncludes) (*TemplateNode, error) {
	return e.composer.Compose(tmpl, confirm)
}

// RenderNode renders all files from a template tree with the given contexts.
func (e *Engine) RenderNode(node *TemplateNode, contexts RenderContexts) ([]RenderedFile, error) {
	return e.renderer.RenderAll(node, contexts)
}

// GetFullTree returns a TemplateNode tree with ALL includes enabled.
func (e *Engine) GetFullTree(ref TemplateRef) (*TemplateNode, error) {
	tmpl, err := e.LoadTemplate(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Always return all includes as enabled.
	return e.composer.Compose(tmpl, func(includes []Include) ([]Include, error) {
		return includes, nil
	})
}

// ValidateTree recursively validates a template tree.
func (e *Engine) ValidateTree(node *TemplateNode) error {
	return e.validator.ValidateTree(node)
}

// ValidateContexts recursively validates that all required variables are present
// in the provided contexts for the entire tree.
func (e *Engine) ValidateContexts(node *TemplateNode, contexts RenderContexts) error {
	return e.validator.ValidateTreeContexts(node, contexts)
}

// AddTemplateFunc adds a custom function to the template renderer
func (e *Engine) AddTemplateFunc(name string, fn any) {
	e.renderer.AddFunc(name, fn)
}

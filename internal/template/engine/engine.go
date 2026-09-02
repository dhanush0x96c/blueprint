// Package engine provides the unified template engine that orchestrates loading, composing, validating, and rendering templates.
package engine

import (
	"fmt"
	"io/fs"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/template/composer"
	"github.com/dhanush0x96c/blueprint/internal/template/loader"
	"github.com/dhanush0x96c/blueprint/internal/template/renderer"
	"github.com/dhanush0x96c/blueprint/internal/template/validator"
)

// Engine is the unified template engine that orchestrates loading, composing, and rendering
type Engine struct {
	resolver  composer.Resolver
	loader    *loader.Loader
	composer  *composer.Composer
	renderer  *renderer.Renderer
	validator *validator.Validator
}

// NewEngine creates a new template engine with the given resolver
func NewEngine(resolver composer.Resolver) *Engine {
	l := loader.NewLoader()
	c := composer.NewComposer(resolver, l)
	r := renderer.NewRenderer()
	v := validator.NewValidator()

	return &Engine{
		resolver:  resolver,
		loader:    l,
		composer:  c,
		renderer:  r,
		validator: v,
	}
}

// LoadTemplate loads a template from the given reference
func (e *Engine) LoadTemplate(ref template.Ref) (*loader.LoadedTemplate, error) {
	resolved, err := e.resolver.Resolve(ref)
	if err != nil {
		return nil, err
	}
	return e.loader.Load(resolved.FS, resolved.Path)
}

// LoadTemplateByPath loads a template from a specific path on a filesystem
func (e *Engine) LoadTemplateByPath(fsys fs.FS, path string) (*loader.LoadedTemplate, error) {
	return e.loader.Load(fsys, path)
}

// Compose resolves all includes for a template recursively and builds a Node tree.
// It calls confirm for all includes of a template to decide which ones should be loaded.
func (e *Engine) Compose(loaded *loader.LoadedTemplate, confirm template.ConfirmIncludes) (*template.Node, error) {
	return e.composer.Compose(loaded, confirm)
}

// RenderNode renders all files from a template tree with the given contexts.
func (e *Engine) RenderNode(node *template.Node, contexts template.RenderContexts) (*template.RenderResult, error) {
	return e.renderer.RenderAll(node, contexts)
}

// GetFullTree loads a template, resolves all includes using the provided confirm function,
// and validates the resulting tree.
func (e *Engine) GetFullTree(ref template.Ref, confirm template.ConfirmIncludes) (*template.Node, error) {
	loaded, err := e.LoadTemplate(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tree, err := e.composer.Compose(loaded, confirm)
	if err != nil {
		return nil, err
	}

	if err := e.ValidateTree(tree); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return tree, nil
}

// ValidateTree recursively validates a template tree.
func (e *Engine) ValidateTree(node *template.Node) error {
	return e.validator.ValidateTree(node)
}

// ValidateContexts recursively validates that all required variables are present
// in the provided contexts for the entire tree.
func (e *Engine) ValidateContexts(node *template.Node, contexts template.RenderContexts) error {
	return e.validator.ValidateTreeContexts(node, contexts)
}

// AddTemplateFunc adds a custom function to the template renderer
func (e *Engine) AddTemplateFunc(name string, fn any) {
	e.renderer.AddFunc(name, fn)
}

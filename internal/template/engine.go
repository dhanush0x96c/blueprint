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
func (e *Engine) LoadTemplate(ref TemplateRef) (*LoadedTemplate, error) {
	resolved, err := e.resolver.Resolve(ref)
	if err != nil {
		return nil, err
	}
	return e.loader.Load(resolved.FS, resolved.Path)
}

// LoadTemplateByPath loads a template from a specific path on a filesystem
func (e *Engine) LoadTemplateByPath(fsys fs.FS, path string) (*LoadedTemplate, error) {
	return e.loader.Load(fsys, path)
}

// Compose resolves all includes for a template recursively and builds a TemplateNode tree.
// It calls confirm for all includes of a template to decide which ones should be loaded.
func (e *Engine) Compose(loaded *LoadedTemplate, confirm ConfirmIncludes) (*TemplateNode, error) {
	return e.composer.Compose(loaded, confirm)
}

// RenderNode renders all files from a template tree with the given contexts.
func (e *Engine) RenderNode(node *TemplateNode, contexts RenderContexts) (*RenderResult, error) {
	return e.renderer.RenderAll(node, contexts)
}

// GetFullTree loads a template, resolves all includes using the provided confirm function,
// and validates the resulting tree.
func (e *Engine) GetFullTree(ref TemplateRef, confirm ConfirmIncludes) (*TemplateNode, error) {
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
func (e *Engine) ValidateTree(node *TemplateNode) error {
	return e.validator.ValidateTree(node)
}

// ValidateContexts recursively validates that all required variables are present
// in the provided contexts for the entire tree.
func (e *Engine) ValidateContexts(node *TemplateNode, contexts RenderContexts) error {
	return e.validator.ValidateTreeContexts(node, contexts)
}

// BuildOptions holds options for building render contexts from a template tree.
type BuildOptions struct {
	GlobalVars       map[string]string
	NameSpecificVars map[string]map[string]string
	NodeSpecificVars map[string]map[string]string
}

// BuildContext recursively builds the render contexts from the template tree.
// Variables are applied in order: defaults -> inheritance -> global -> name-specific -> node-specific.
func (e *Engine) BuildContext(node *TemplateNode, opts BuildOptions) RenderContexts {
	contexts := make(RenderContexts)
	e.fillContexts(node, "", contexts, opts)
	return contexts
}

func (e *Engine) fillContexts(node *TemplateNode, parentID string, contexts RenderContexts, opts BuildOptions) {
	if _, ok := contexts[node.ID]; !ok {
		ctx := NewTemplateContext(make(map[string]any))

		e.applyDefaults(node, ctx)
		e.applyInheritance(node, parentID, contexts, ctx)

		for k, v := range opts.GlobalVars {
			ctx.Set(k, v)
		}

		if nameVars, ok := opts.NameSpecificVars[node.Template.Name]; ok {
			for k, v := range nameVars {
				ctx.Set(k, v)
			}
		}

		if nodeVars, ok := opts.NodeSpecificVars[node.ID]; ok {
			for k, v := range nodeVars {
				ctx.Set(k, v)
			}
		}

		contexts[node.ID] = ctx
	}

	for _, child := range node.Children {
		e.fillContexts(child, node.ID, contexts, opts)
	}
}

func (e *Engine) applyInheritance(node *TemplateNode, parentID string, contexts RenderContexts, ctx *Context) {
	if parentID == "" || len(node.Inherited) == 0 {
		return
	}

	parentCtx, ok := contexts[parentID]
	if !ok {
		return
	}

	for childVar, parentVar := range node.Inherited {
		if val, ok := parentCtx.Get(parentVar); ok {
			ctx.Set(childVar, val)
		}
	}
}

func (e *Engine) applyDefaults(node *TemplateNode, ctx *Context) {
	for _, v := range node.Template.Variables {
		if _, inherited := node.Inherited[v.Name]; !inherited {
			if v.Default != nil {
				ctx.Set(v.Name, v.Default)
			}
		}
	}
}

// AddTemplateFunc adds a custom function to the template renderer
func (e *Engine) AddTemplateFunc(name string, fn any) {
	e.renderer.AddFunc(name, fn)
}

package template

import (
	"fmt"
	"slices"
)

// Composer handles building the TemplateNode tree from a root Template.
type Composer struct {
	resolver Resolver
	loader   Loader
}

// NewComposer creates a new template composer with the given resolver and loader.
func NewComposer(resolver Resolver, loader Loader) *Composer {
	return &Composer{
		resolver: resolver,
		loader:   loader,
	}
}

// Compose resolves all includes for a template recursively and builds a TemplateNode tree.
// It calls confirm for all includes of a template to decide which ones should be loaded.
func (c *Composer) Compose(tmpl *Template, confirm ConfirmIncludes) (*TemplateNode, error) {
	return c.doCompose(tmpl, []string{tmpl.Name}, confirm)
}

// doCompose is the internal recursive composition function that tracks the stack
// to detect circular dependencies and builds the TemplateNode tree.
func (c *Composer) doCompose(tmpl *Template, stack []string, confirm ConfirmIncludes) (*TemplateNode, error) {
	node := &TemplateNode{
		Template: tmpl,
		Children: make([]*TemplateNode, 0),
	}

	if len(tmpl.Includes) == 0 {
		return node, nil
	}

	enabledIncludes, err := confirm(tmpl.Includes)
	if err != nil {
		return nil, err
	}

	for _, inc := range enabledIncludes {
		if slices.Contains(stack, inc.Name) {
			return nil, fmt.Errorf("circular dependency detected: %v -> %s", stack, inc.Name)
		}

		ref := TemplateRef{
			Name: inc.Name,
		}

		resolved, err := c.resolver.Resolve(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve included template '%s': %w", inc.Name, err)
		}

		includedTmpl, err := c.loader.Load(resolved.FS, resolved.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to load included template '%s' from %s: %w", inc.Name, resolved.Path, err)
		}

		newStack := append(slices.Clone(stack), inc.Name)
		childNode, err := c.doCompose(includedTmpl, newStack, confirm)
		if err != nil {
			return nil, err
		}

		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

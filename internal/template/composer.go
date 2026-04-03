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
func (c *Composer) Compose(loaded *LoadedTemplate, confirm ConfirmIncludes) (*TemplateNode, error) {
	return c.doCompose(loaded, []string{loaded.Template.Name}, confirm, "0")
}

// doCompose is the internal recursive composition function that tracks the stack
// to detect circular dependencies and builds the TemplateNode tree.
func (c *Composer) doCompose(loaded *LoadedTemplate, stack []string, confirm ConfirmIncludes, id string) (*TemplateNode, error) {
	node := &TemplateNode{
		ID:       id,
		Template: loaded.Template,
		FS:       loaded.FS,
		Path:     loaded.Path,
		Children: make([]*TemplateNode, 0),
	}

	if len(loaded.Template.Includes) == 0 {
		return node, nil
	}

	enabledIncludes, err := confirm(loaded.Template.Includes)
	if err != nil {
		return nil, err
	}

	for i, inc := range enabledIncludes {
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
		childID := fmt.Sprintf("%s.%d", id, i)
		childNode, err := c.doCompose(includedTmpl, newStack, confirm, childID)
		if err != nil {
			return nil, err
		}
		childNode.Mount = inc.Mount

		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

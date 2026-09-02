// Package composer handles building and resolving template node trees from includes.
package composer

import (
	"fmt"
	"io/fs"
	"slices"

	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/template/loader"
)

// Loader handles loading templates from the filesystem.
type Loader interface {
	Load(fsys fs.FS, pth string) (*loader.LoadedTemplate, error)
}

// Composer handles building the Node tree from a root Template.
type Composer struct {
	resolver template.Resolver
	loader   Loader
}

// NewComposer creates a new template composer with the given resolver and loader.
func NewComposer(resolver template.Resolver, loader Loader) *Composer {
	return &Composer{
		resolver: resolver,
		loader:   loader,
	}
}

// Compose resolves all includes for a template recursively and builds a Node tree.
// It calls confirm for all includes of a template to decide which ones should be loaded.
func (c *Composer) Compose(loaded *loader.LoadedTemplate, confirm template.ConfirmIncludes) (*template.Node, error) {
	return c.doCompose(loaded, []string{loaded.Template.Name}, confirm, "0")
}

// doCompose is the internal recursive composition function that tracks the stack
// to detect circular dependencies and builds the Node tree.
func (c *Composer) doCompose(loaded *loader.LoadedTemplate, stack []string, confirm template.ConfirmIncludes, id string) (*template.Node, error) {
	node := &template.Node{
		ID:       id,
		Template: loaded.Template,
		FS:       loaded.FS,
		Path:     loaded.Path,
		Children: make([]*template.Node, 0),
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

		ref := template.TemplateRef{
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
		childNode.Inherited = inc.Inherits

		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

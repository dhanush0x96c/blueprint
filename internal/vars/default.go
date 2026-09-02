package vars

import "github.com/dhanush0x96c/blueprint/internal/template"

// DefaultCollector populates default variable values into render contexts.
type DefaultCollector struct {
	tree *template.Node
}

// NewDefaultCollector creates a new DefaultCollector.
func NewDefaultCollector(tree *template.Node) *DefaultCollector {
	return &DefaultCollector{tree: tree}
}

// Collect applies default variable values to render contexts for all nodes in the tree.
func (c *DefaultCollector) Collect(contexts template.RenderContexts) error {
	return walk(c.tree, func(node *template.Node) error {
		ctx := ensureContext(contexts, node.ID)
		for _, variable := range node.RequiredVariables() {
			if variable.Default != nil {
				ctx.Set(variable.Name, variable.Default)
			}
		}
		return nil
	})
}

package vars

import "github.com/dhanush0x96c/blueprint/internal/template"

type DefaultCollector struct {
	tree *template.Node
}

func NewDefaultCollector(tree *template.Node) *DefaultCollector {
	return &DefaultCollector{tree: tree}
}

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

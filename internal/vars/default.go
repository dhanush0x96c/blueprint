package vars

import "github.com/dhanush0x96c/blueprint/internal/template"

type DefaultCollector struct {
	tree *template.TemplateNode
}

func NewDefaultCollector(tree *template.TemplateNode) *DefaultCollector {
	return &DefaultCollector{tree: tree}
}

func (c *DefaultCollector) Collect(contexts template.RenderContexts) error {
	walk(c.tree, func(node *template.TemplateNode) error {
		ctx := ensureContext(contexts, node.ID)
		for _, variable := range node.RequiredVariables() {
			if variable.Default != nil {
				ctx.Set(variable.Name, variable.Default)
			}
		}
		return nil
	})

	return nil
}

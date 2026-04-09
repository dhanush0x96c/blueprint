package vars

import "github.com/dhanush0x96c/blueprint/internal/template"

type CLICollector struct {
	tree *template.TemplateNode
	args Variables
}

func NewCLICollector(tree *template.TemplateNode, args Variables) *CLICollector {
	return &CLICollector{
		tree: tree,
		args: args,
	}
}

func (c *CLICollector) Collect(contexts template.RenderContexts) error {
	walk(c.tree, func(node *template.TemplateNode) error {
		ctx := ensureContext(contexts, node.ID)

		for key, value := range c.args.Global {
			ctx.Set(key, value)
		}

		if nameVars, ok := c.args.NameSpecific[node.Template.Name]; ok {
			for key, value := range nameVars {
				ctx.Set(key, value)
			}
		}

		if nodeVars, ok := c.args.NodeSpecific[node.ID]; ok {
			for key, value := range nodeVars {
				ctx.Set(key, value)
			}
		}

		return nil
	})

	return nil
}

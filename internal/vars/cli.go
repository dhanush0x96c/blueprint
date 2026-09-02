package vars

import "github.com/dhanush0x96c/blueprint/internal/template"

// CLICollector collects variables from CLI arguments into render contexts.
type CLICollector struct {
	tree *template.Node
	args Variables
}

// NewCLICollector creates a new CLICollector.
func NewCLICollector(tree *template.Node, args Variables) *CLICollector {
	return &CLICollector{
		tree: tree,
		args: args,
	}
}

// Collect applies CLI arguments to render contexts for all nodes in the tree.
func (c *CLICollector) Collect(contexts template.RenderContexts) error {
	return walk(c.tree, func(node *template.Node) error {
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
}

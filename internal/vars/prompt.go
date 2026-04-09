package vars

import (
	"fmt"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

type PromptEngine interface {
	PromptVariables([]template.Variable) (*template.Context, error)
}

type PromptCollector struct {
	tree   *template.TemplateNode
	engine PromptEngine
}

func NewPromptCollector(tree *template.TemplateNode, engine PromptEngine) *PromptCollector {
	return &PromptCollector{
		tree:   tree,
		engine: engine,
	}
}

func (c *PromptCollector) Collect(contexts template.RenderContexts) error {
	return walk(c.tree, func(node *template.TemplateNode) error {
		ctx := ensureContext(contexts, node.ID)
		variablesToPrompt := node.RequiredVariables()
		if len(variablesToPrompt) == 0 {
			return nil
		}

		fmt.Printf("\n--- Variables for %s (ID: %s) ---\n", node.Template.Name, node.ID)
		prompted, err := c.engine.PromptVariables(variablesToPrompt)
		if err != nil {
			return fmt.Errorf("failed to collect variables for %s: %w", node.Template.Name, err)
		}

		ctx.Merge(prompted)
		return nil
	})
}

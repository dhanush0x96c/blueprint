package vars

import (
	"fmt"

	"github.com/dhanush0x96c/blueprint/internal/prompt"
	"github.com/dhanush0x96c/blueprint/internal/template"
)

type PromptCollector struct {
	tree   *template.TemplateNode
	engine *prompt.Engine
}

func NewPromptCollector(tree *template.TemplateNode, engine *prompt.Engine) *PromptCollector {
	return &PromptCollector{
		tree:   tree,
		engine: engine,
	}
}

func (c *PromptCollector) Collect(contexts template.RenderContexts) error {
	return walk(c.tree, func(node *template.TemplateNode) error {
		ctx := ensureContext(contexts, node.ID)
		group := c.variableGroup(node, ctx)
		if len(group.Variables) == 0 {
			return nil
		}

		prompted, err := c.engine.PromptVariables(group)
		if err != nil {
			return fmt.Errorf("failed to collect variables for %s: %w", node.Template.Name, err)
		}

		ctx.Merge(prompted)
		return nil
	})
}

func (c *PromptCollector) variableGroup(node *template.TemplateNode, ctx *template.Context) prompt.VariableGroup {
	variables := node.RequiredVariables()
	group := prompt.VariableGroup{
		Title:     fmt.Sprintf("Variables for %s (ID: %s)", node.Template.Name, node.ID),
		Variables: make([]prompt.Variable, 0, len(variables)),
	}

	for _, variable := range variables {
		promptVariable := prompt.Variable{Variable: variable}
		if value, ok := ctx.Get(variable.Name); ok {
			promptVariable.Value = value
		}

		group.Variables = append(group.Variables, promptVariable)
	}

	return group
}

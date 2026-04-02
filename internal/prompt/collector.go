package prompt

import (
	"fmt"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// Collector collects variables from templates and their includes
type Collector struct {
	engine *Engine
}

// NewCollector creates a new variable collector
func NewCollector() *Collector {
	return &Collector{
		engine: NewEngine(),
	}
}

// ConfirmIncludes satisfies the template.ConfirmIncludes signature.
// It prompts the user for which includes to enable and returns the selected slice.
func (c *Collector) ConfirmIncludes(includes []template.Include) ([]template.Include, error) {
	// Skip prompting if there are no includes
	if len(includes) == 0 {
		return nil, nil
	}

	enabledMap, err := c.engine.PromptIncludes(includes)
	if err != nil {
		return nil, err
	}

	var enabled []template.Include
	for _, inc := range includes {
		if enabledMap[inc.Name] {
			enabled = append(enabled, inc)
		}
	}

	return enabled, nil
}

// CollectTreeVariables walks the tree and prompts for variables for each node.
// It returns a RenderContexts map (templateName -> Context).
func (c *Collector) CollectTreeVariables(node *template.TemplateNode) (template.RenderContexts, error) {
	contexts := make(template.RenderContexts)
	if err := c.collectNodeVariables(node, contexts); err != nil {
		return nil, err
	}
	return contexts, nil
}

// collectNodeVariables recursively collects variables for a node and its children.
func (c *Collector) collectNodeVariables(node *template.TemplateNode, contexts template.RenderContexts) error {
	// If we already have a context for this node, skip it.
	if _, ok := contexts[node.ID]; !ok {
		fmt.Printf("\n--- Variables for %s (ID: %s) ---\n", node.Template.Name, node.ID)
		ctx, err := c.engine.PromptVariablesAsForm(node.Template.Variables)
		if err != nil {
			return fmt.Errorf("failed to collect variables for %s: %w", node.Template.Name, err)
		}
		contexts[node.ID] = ctx
	}

	for _, child := range node.Children {
		if err := c.collectNodeVariables(child, contexts); err != nil {
			return err
		}
	}

	return nil
}

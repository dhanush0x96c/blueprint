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
	if err := c.collectNodeVariables(node, "", contexts); err != nil {
		return nil, err
	}
	return contexts, nil
}

// collectNodeVariables recursively collects variables for a node and its children.
func (c *Collector) collectNodeVariables(node *template.TemplateNode, parentID string, contexts template.RenderContexts) error {
	// If we already have a context for this node, skip it.
	if _, ok := contexts[node.ID]; !ok {
		ctx := c.buildNodeContext(node, parentID, contexts)

		if err := c.promptForVariables(node, ctx); err != nil {
			return err
		}

		contexts[node.ID] = ctx
	}

	for _, child := range node.Children {
		if err := c.collectNodeVariables(child, node.ID, contexts); err != nil {
			return err
		}
	}

	return nil
}

// buildNodeContext initializes the context for a node, handling inheritance.
func (c *Collector) buildNodeContext(node *template.TemplateNode, parentID string, contexts template.RenderContexts) *template.Context {
	ctx := template.NewTemplateContext(make(map[string]any))

	// Inherit variables from parent context if any
	if parentID != "" && len(node.Inherited) > 0 {
		parentCtx := contexts[parentID]
		for childVar, parentVar := range node.Inherited {
			if val, ok := parentCtx.Get(parentVar); ok {
				ctx.Set(childVar, val)
			}
		}
	}

	return ctx
}

// promptForVariables filters inherited variables and prompts for the remaining ones.
func (c *Collector) promptForVariables(node *template.TemplateNode, ctx *template.Context) error {
	variablesToPrompt := node.RequiredVariables()

	if len(variablesToPrompt) > 0 {
		fmt.Printf("\n--- Variables for %s (ID: %s) ---\n", node.Template.Name, node.ID)
		promptedCtx, err := c.engine.PromptVariables(variablesToPrompt)
		if err != nil {
			return fmt.Errorf("failed to collect variables for %s: %w", node.Template.Name, err)
		}
		// Merge prompted values into our context
		ctx.Merge(promptedCtx)
	}

	return nil
}

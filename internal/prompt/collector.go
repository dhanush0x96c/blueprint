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

// CollectFromTemplate collects all variables from a template
// This is a simple wrapper around the engine's PromptVariablesAsForm
func (c *Collector) CollectFromTemplate(tmpl *template.Template) (*template.Context, error) {
	if len(tmpl.Variables) == 0 {
		return template.NewTemplateContext(make(map[string]any)), nil
	}

	return c.engine.PromptVariablesAsForm(tmpl.Variables)
}

// CollectWithIncludes collects variables and include selections from a template
// Returns the context and a map of enabled includes
func (c *Collector) CollectWithIncludes(tmpl *template.Template, allIncludes []template.Include) (*template.Context, map[string]bool, error) {
	// First, prompt for which includes to enable
	enabledIncludes := make(map[string]bool)
	if len(allIncludes) > 0 {
		selected, err := c.engine.PromptIncludes(allIncludes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to collect includes: %w", err)
		}
		enabledIncludes = selected
	}

	// Then collect variables from the main template
	ctx, err := c.CollectFromTemplate(tmpl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect variables: %w", err)
	}

	return ctx, enabledIncludes, nil
}

// CollectInteractive provides a complete interactive flow:
// 1. Show available includes
// 2. Prompt user to select includes
// 3. Collect variables from template and selected includes
func (c *Collector) CollectInteractive(tmpl *template.Template, allIncludes []template.Include) (*template.Context, map[string]bool, error) {
	return c.CollectWithIncludes(tmpl, allIncludes)
}

// MergeContexts merges multiple template contexts into one
// Later contexts take precedence over earlier ones for duplicate keys
func (c *Collector) MergeContexts(contexts ...*template.Context) *template.Context {
	merged := template.NewTemplateContext(make(map[string]any))

	for _, ctx := range contexts {
		merged.Merge(ctx)
	}

	return merged
}

// CollectWithDefaults collects variables but uses provided defaults for missing values
func (c *Collector) CollectWithDefaults(tmpl *template.Template, defaults map[string]any) (*template.Context, error) {
	// Create a context with defaults
	ctx := template.NewTemplateContext(defaults)

	// If there are no variables to collect, return the defaults
	if len(tmpl.Variables) == 0 {
		return ctx, nil
	}

	// Collect user input
	userCtx, err := c.engine.PromptVariablesAsForm(tmpl.Variables)
	if err != nil {
		return nil, err
	}

	// Merge user input over defaults
	ctx.Merge(userCtx)

	return ctx, nil
}

// ValidateContext validates that all required variables are present in the context
func (c *Collector) ValidateContext(tmpl *template.Template, ctx *template.Context) error {
	for _, variable := range tmpl.Variables {
		_, exists := ctx.Get(variable.Name)
		if !exists && variable.Default == nil {
			return fmt.Errorf("required variable %s is missing", variable.Name)
		}
	}
	return nil
}

// GetMissingVariables returns a list of variables that are not in the context
func (c *Collector) GetMissingVariables(tmpl *template.Template, ctx *template.Context) []template.Variable {
	missing := make([]template.Variable, 0)

	for _, variable := range tmpl.Variables {
		_, exists := ctx.Get(variable.Name)
		if !exists {
			missing = append(missing, variable)
		}
	}

	return missing
}

// CollectMissing collects only the variables that are missing from the context
func (c *Collector) CollectMissing(tmpl *template.Template, ctx *template.Context) error {
	missing := c.GetMissingVariables(tmpl, ctx)

	if len(missing) == 0 {
		return nil
	}

	missingCtx, err := c.engine.PromptVariablesAsForm(missing)
	if err != nil {
		return fmt.Errorf("failed to collect missing variables: %w", err)
	}

	ctx.Merge(missingCtx)
	return nil
}

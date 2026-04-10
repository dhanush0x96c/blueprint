package prompt

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/dhanush0x96c/blueprint/internal/template"
)

// Engine handles interactive prompts for collecting template variables
type Engine struct {
	theme *huh.Theme
}

// NewEngine creates a new prompt engine
func NewEngine() *Engine {
	return &Engine{
		theme: huh.ThemeCharm(),
	}
}

// PromptVariables prompts for all variables as a single form
// This provides a better UX than individual prompts
func (e *Engine) PromptVariables(group VariableGroup) (*template.Context, error) {
	if len(group.Variables) == 0 {
		return template.NewTemplateContext(make(map[string]any)), nil
	}

	fields := make([]huh.Field, 0, len(group.Variables))
	values := make(map[string]any)

	for _, variable := range group.Variables {
		field, valuePtr := e.createFormField(variable)
		if field != nil {
			fields = append(fields, field)
			values[variable.Name] = valuePtr
		}
	}

	form := huh.NewForm(
		huh.NewGroup(fields...).Title(group.Title),
	).WithTheme(e.theme)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("form prompt failed: %w", err)
	}

	// Extract actual values from pointers
	ctx := template.NewTemplateContext(make(map[string]any))
	for _, variable := range group.Variables {
		valuePtr := values[variable.Name]
		ctx.Set(variable.Name, extractValue(valuePtr, variable.Type))
	}

	return ctx, nil
}

// PromptIncludes prompts the user to select which includes to enable
func (e *Engine) PromptIncludes(includes []template.Include) (map[string]bool, error) {
	if len(includes) == 0 {
		return make(map[string]bool), nil
	}

	options := make([]huh.Option[string], len(includes))
	selected := make([]string, 0)

	// Pre-select includes that are enabled by default
	for i, inc := range includes {
		options[i] = huh.NewOption(inc.Name, inc.Name)
		if inc.EnabledByDefault {
			selected = append(selected, inc.Name)
		}
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select features to include").
				Description("Use space to select/deselect, enter to confirm").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(e.theme).Run()

	if err != nil {
		return nil, fmt.Errorf("include selection failed: %w", err)
	}

	// Convert to map
	enabledIncludes := make(map[string]bool)
	for _, incName := range selected {
		enabledIncludes[incName] = true
	}

	return enabledIncludes, nil
}

// createFormField creates a huh form field for a variable
func (e *Engine) createFormField(variable Variable) (huh.Field, any) {
	switch variable.Type {
	case template.VariableTypeString:
		value := CastValue[string](variable.Value)
		return huh.NewInput().
			Title(variable.Prompt).
			Value(&value), &value

	case template.VariableTypeInt:
		var value string
		if variable.Value != nil {
			value = fmt.Sprintf("%v", variable.Value)
		}
		return huh.NewInput().
			Title(variable.Prompt).
			Value(&value).
			Validate(func(s string) error {
				if s == "" && variable.Value != nil {
					return nil
				}
				_, err := strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("must be a valid integer")
				}
				return nil
			}), &value

	case template.VariableTypeBool:
		value := CastValue[bool](variable.Value)
		return huh.NewConfirm().
			Title(variable.Prompt).
			Value(&value), &value

	case template.VariableTypeSelect:
		value := CastValue[string](variable.Value)
		options := make([]huh.Option[string], len(variable.Options))
		for i, opt := range variable.Options {
			options[i] = huh.NewOption(opt, opt)
		}
		return huh.NewSelect[string]().
			Title(variable.Prompt).
			Options(options...).
			Value(&value), &value

	case template.VariableTypeMultiSelect:
		value := CastValue[[]string](variable.Value)
		options := make([]huh.Option[string], len(variable.Options))
		for i, opt := range variable.Options {
			options[i] = huh.NewOption(opt, opt)
		}
		return huh.NewMultiSelect[string]().
			Title(variable.Prompt).
			Options(options...).
			Value(&value), &value

	default:
		return nil, nil
	}
}

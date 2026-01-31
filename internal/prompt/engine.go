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

// NewEngineWithTheme creates a new prompt engine with a custom theme
func NewEngineWithTheme(theme *huh.Theme) *Engine {
	return &Engine{
		theme: theme,
	}
}

// PromptVariable prompts the user for a single variable and returns its value
func (e *Engine) PromptVariable(variable template.Variable) (any, error) {
	switch variable.Type {
	case template.VariableTypeString:
		return e.promptString(variable)
	case template.VariableTypeInt:
		return e.promptInt(variable)
	case template.VariableTypeBool:
		return e.promptBool(variable)
	case template.VariableTypeSelect:
		return e.promptSelect(variable)
	case template.VariableTypeMultiSelect:
		return e.promptMultiSelect(variable)
	default:
		return nil, fmt.Errorf("unsupported variable type: %s", variable.Type)
	}
}

// PromptVariables prompts the user for all variables and returns a context
func (e *Engine) PromptVariables(variables []template.Variable) (*template.Context, error) {
	ctx := template.NewTemplateContext(make(map[string]any))

	for _, variable := range variables {
		value, err := e.PromptVariable(variable)
		if err != nil {
			return nil, fmt.Errorf("failed to prompt for variable %s: %w", variable.Name, err)
		}
		ctx.Set(variable.Name, value)
	}

	return ctx, nil
}

// PromptVariablesAsForm prompts for all variables as a single form
// This provides a better UX than individual prompts
func (e *Engine) PromptVariablesAsForm(variables []template.Variable) (*template.Context, error) {
	if len(variables) == 0 {
		return template.NewTemplateContext(make(map[string]any)), nil
	}

	fields := make([]huh.Field, 0, len(variables))
	values := make(map[string]any)

	for _, variable := range variables {
		field, valuePtr := e.createFormField(variable, values)
		if field != nil {
			fields = append(fields, field)
			values[variable.Name] = valuePtr
		}
	}

	form := huh.NewForm(
		huh.NewGroup(fields...),
	).WithTheme(e.theme)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("form prompt failed: %w", err)
	}

	// Extract actual values from pointers
	ctx := template.NewTemplateContext(make(map[string]any))
	for _, variable := range variables {
		valuePtr := values[variable.Name]
		ctx.Set(variable.Name, e.extractValue(valuePtr, variable.Type))
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
		options[i] = huh.NewOption(inc.Template, inc.Template)
		if inc.EnabledByDefault {
			selected = append(selected, inc.Template)
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
func (e *Engine) createFormField(variable template.Variable, values map[string]any) (huh.Field, any) {
	switch variable.Type {
	case template.VariableTypeString:
		var value string
		if variable.Default != nil {
			if defaultStr, ok := variable.Default.(string); ok {
				value = defaultStr
			}
		}
		values[variable.Name] = &value
		return huh.NewInput().
			Title(e.getPromptText(variable)).
			Value(&value).
			Placeholder(e.getPlaceholder(variable)), &value

	case template.VariableTypeInt:
		var value string
		if variable.Default != nil {
			value = fmt.Sprintf("%v", variable.Default)
		}
		values[variable.Name] = &value
		return huh.NewInput().
			Title(e.getPromptText(variable)).
			Value(&value).
			Placeholder(e.getPlaceholder(variable)).
			Validate(func(s string) error {
				if s == "" && variable.Default != nil {
					return nil
				}
				_, err := strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("must be a valid integer")
				}
				return nil
			}), &value

	case template.VariableTypeBool:
		var value bool
		if variable.Default != nil {
			if defaultBool, ok := variable.Default.(bool); ok {
				value = defaultBool
			}
		}
		values[variable.Name] = &value
		return huh.NewConfirm().
			Title(e.getPromptText(variable)).
			Value(&value), &value

	case template.VariableTypeSelect:
		var value string
		if variable.Default != nil {
			if defaultStr, ok := variable.Default.(string); ok {
				value = defaultStr
			}
		}
		options := make([]huh.Option[string], len(variable.Options))
		for i, opt := range variable.Options {
			options[i] = huh.NewOption(opt, opt)
		}
		values[variable.Name] = &value
		return huh.NewSelect[string]().
			Title(e.getPromptText(variable)).
			Options(options...).
			Value(&value), &value

	case template.VariableTypeMultiSelect:
		var value []string
		if variable.Default != nil {
			if defaultSlice, ok := variable.Default.([]string); ok {
				value = defaultSlice
			}
		}
		options := make([]huh.Option[string], len(variable.Options))
		for i, opt := range variable.Options {
			options[i] = huh.NewOption(opt, opt)
		}
		values[variable.Name] = &value
		return huh.NewMultiSelect[string]().
			Title(e.getPromptText(variable)).
			Options(options...).
			Value(&value), &value

	default:
		return nil, nil
	}
}

// extractValue extracts the actual value from the pointer used in the form
func (e *Engine) extractValue(valuePtr any, varType template.VariableType) any {
	switch varType {
	case template.VariableTypeString:
		if ptr, ok := valuePtr.(*string); ok {
			return *ptr
		}
	case template.VariableTypeInt:
		if ptr, ok := valuePtr.(*string); ok {
			if *ptr == "" {
				return 0
			}
			val, _ := strconv.Atoi(*ptr)
			return val
		}
	case template.VariableTypeBool:
		if ptr, ok := valuePtr.(*bool); ok {
			return *ptr
		}
	case template.VariableTypeSelect:
		if ptr, ok := valuePtr.(*string); ok {
			return *ptr
		}
	case template.VariableTypeMultiSelect:
		if ptr, ok := valuePtr.(*[]string); ok {
			return *ptr
		}
	}
	return valuePtr
}

// getPromptText returns the prompt text for a variable
func (e *Engine) getPromptText(variable template.Variable) string {
	if variable.Prompt != "" {
		return variable.Prompt
	}
	return variable.Name
}

// getPlaceholder returns a placeholder for a variable input
func (e *Engine) getPlaceholder(variable template.Variable) string {
	if variable.Default != nil {
		return fmt.Sprintf("(default: %v)", variable.Default)
	}
	return ""
}

// promptString prompts for a string value
func (e *Engine) promptString(variable template.Variable) (string, error) {
	var value string
	if variable.Default != nil {
		if defaultStr, ok := variable.Default.(string); ok {
			value = defaultStr
		}
	}

	err := huh.NewInput().
		Title(e.getPromptText(variable)).
		Value(&value).
		Placeholder(e.getPlaceholder(variable)).
		Run()

	if err != nil {
		return "", err
	}
	return value, nil
}

// promptInt prompts for an integer value
func (e *Engine) promptInt(variable template.Variable) (int, error) {
	var valueStr string
	if variable.Default != nil {
		valueStr = fmt.Sprintf("%v", variable.Default)
	}

	err := huh.NewInput().
		Title(e.getPromptText(variable)).
		Value(&valueStr).
		Placeholder(e.getPlaceholder(variable)).
		Validate(func(s string) error {
			if s == "" && variable.Default != nil {
				return nil
			}
			_, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be a valid integer")
			}
			return nil
		}).
		Run()

	if err != nil {
		return 0, err
	}

	if valueStr == "" && variable.Default != nil {
		if defaultInt, ok := variable.Default.(int); ok {
			return defaultInt, nil
		}
	}

	return strconv.Atoi(valueStr)
}

// promptBool prompts for a boolean value
func (e *Engine) promptBool(variable template.Variable) (bool, error) {
	var value bool
	if variable.Default != nil {
		if defaultBool, ok := variable.Default.(bool); ok {
			value = defaultBool
		}
	}

	err := huh.NewConfirm().
		Title(e.getPromptText(variable)).
		Value(&value).
		Run()

	if err != nil {
		return false, err
	}
	return value, nil
}

// promptSelect prompts for a single selection
func (e *Engine) promptSelect(variable template.Variable) (string, error) {
	var value string
	if variable.Default != nil {
		if defaultStr, ok := variable.Default.(string); ok {
			value = defaultStr
		}
	}

	options := make([]huh.Option[string], len(variable.Options))
	for i, opt := range variable.Options {
		options[i] = huh.NewOption(opt, opt)
	}

	err := huh.NewSelect[string]().
		Title(e.getPromptText(variable)).
		Options(options...).
		Value(&value).
		Run()

	if err != nil {
		return "", err
	}
	return value, nil
}

// promptMultiSelect prompts for multiple selections
func (e *Engine) promptMultiSelect(variable template.Variable) ([]string, error) {
	var value []string
	if variable.Default != nil {
		if defaultSlice, ok := variable.Default.([]string); ok {
			value = defaultSlice
		}
	}

	options := make([]huh.Option[string], len(variable.Options))
	for i, opt := range variable.Options {
		options[i] = huh.NewOption(opt, opt)
	}

	err := huh.NewMultiSelect[string]().
		Title(e.getPromptText(variable)).
		Options(options...).
		Value(&value).
		Run()

	if err != nil {
		return nil, err
	}
	return value, nil
}

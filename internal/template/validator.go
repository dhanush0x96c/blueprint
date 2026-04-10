package template

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Validator handles template validation with struct tags and semantic rules.
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new template validator.
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateTree recursively validates a template tree.
func (v *Validator) ValidateTree(node *TemplateNode) error {
	var errs []error

	if err := v.Validate(node.Template); err != nil {
		errs = append(errs, err)
	}

	if err := v.validateIncludes(node); err != nil {
		errs = append(errs, err)
	}

	errs = append(errs, v.validateNodeFiles(node)...)

	for _, child := range node.Children {
		if err := v.ValidateTree(child); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// validateNodeFiles validates that all source files exist for a node.
func (v *Validator) validateNodeFiles(node *TemplateNode) []error {
	var errs []error

	for i, file := range node.Template.Files {
		srcPath := path.Join(node.Path, file.Src)
		_, err := fs.Stat(node.FS, srcPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				errs = append(errs, fmt.Errorf("file[%d]: source file %q does not exist", i, srcPath))
			} else {
				errs = append(errs, fmt.Errorf("file[%d]: failed to stat source file %q: %w", i, srcPath, err))
			}
		}
	}

	return errs
}

// ValidateTreeContexts recursively validates that all template variables are present
// in the provided contexts for the entire tree.
func (v *Validator) ValidateTreeContexts(node *TemplateNode, contexts RenderContexts) error {
	ctx, ok := contexts[node.ID]
	if !ok {
		return fmt.Errorf("no context found for template %s (ID: %s)", node.Template.Name, node.ID)
	}

	if err := v.ValidateContext(node.Template, ctx); err != nil {
		return fmt.Errorf("template %s (ID: %s): %w", node.Template.Name, node.ID, err)
	}

	for _, child := range node.Children {
		if err := v.ValidateTreeContexts(child, contexts); err != nil {
			return err
		}
	}

	return nil
}

// ValidateContext validates that all template variables are present in the context.
func (v *Validator) ValidateContext(tmpl *Template, ctx *Context) error {
	for _, variable := range tmpl.Variables {
		value, exists := ctx.Get(variable.Name)
		if !exists {
			return fmt.Errorf("variable %s is missing", variable.Name)
		}

		if err := v.validateVariableValue(variable, value); err != nil {
			return fmt.Errorf("variable %s is invalid: %w", variable.Name, err)
		}
	}
	return nil
}

// Validate validates a template and returns all validation errors.
// Returns nil if the template is valid.
func (v *Validator) Validate(tmpl *Template) error {
	var errs []error

	// Struct tag validation
	if err := v.validate.Struct(tmpl); err != nil {
		errs = append(errs, err)
	}

	// Semantic validation
	errs = append(errs, v.validateVariables(tmpl.Variables)...)

	if err := v.validateProjectNameRole(tmpl); err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// ValidateMetadata validates a template metadata and returns all validation errors.
func (v *Validator) ValidateMetadata(meta *Metadata) error {
	return v.validate.Struct(meta)
}

// validateVariables validates variable-specific rules.
func (v *Validator) validateVariables(vars []Variable) []error {
	var errs []error

	seen := make(map[string]bool)
	for i, variable := range vars {
		// Check for duplicate names
		if seen[variable.Name] {
			errs = append(errs, fmt.Errorf("variable[%d]: duplicate variable name %q", i, variable.Name))
		}
		seen[variable.Name] = true

		if err := v.validateVariableOptions(i, variable); err != nil {
			errs = append(errs, err)
		}

		if variable.Default != nil {
			if err := v.validateVariableValue(variable, variable.Default); err != nil {
				errs = append(errs, fmt.Errorf("variable[%d] %q: invalid default value: %w", i, variable.Name, err))
			}
		}
	}

	return errs
}

func (v *Validator) validateVariableOptions(index int, variable Variable) error {
	if variable.Type != VariableTypeSelect && variable.Type != VariableTypeMultiSelect {
		if len(variable.Options) > 0 {
			return fmt.Errorf("variable[%d] %q: options are only allowed for select and multiselect types", index, variable.Name)
		}
		return nil
	}

	if len(variable.Options) == 0 {
		return fmt.Errorf("variable[%d] %q: options required for type %s", index, variable.Name, variable.Type)
	}

	seen := make(map[string]struct{}, len(variable.Options))
	for optionIndex, option := range variable.Options {
		if option == "" {
			return fmt.Errorf("variable[%d] %q: option[%d] must not be empty", index, variable.Name, optionIndex)
		}

		if _, ok := seen[option]; ok {
			return fmt.Errorf("variable[%d] %q: duplicate option %q", index, variable.Name, option)
		}
		seen[option] = struct{}{}
	}

	return nil
}

func (v *Validator) validateVariableValue(variable Variable, value any) error {
	switch variable.Type {
	case VariableTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected type %s, got %T", variable.Type, value)
		}
		return nil

	case VariableTypeInt:
		if !isIntegerValue(value) {
			return fmt.Errorf("expected type %s, got %T", variable.Type, value)
		}
		return nil

	case VariableTypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected type %s, got %T", variable.Type, value)
		}
		return nil

	case VariableTypeSelect:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected type %s, got %T", variable.Type, value)
		}
		if !containsOption(variable.Options, s) {
			return fmt.Errorf("contains invalid option %q", s)
		}
		return nil

	case VariableTypeMultiSelect:
		values, ok := normalizeStringSlice(value)
		if !ok {
			return fmt.Errorf("expected type %s, got %T", variable.Type, value)
		}
		for _, item := range values {
			if !containsOption(variable.Options, item) {
				return fmt.Errorf("contains invalid option %q", item)
			}
		}
		return nil

	default:
		return fmt.Errorf("unsupported variable type %q", variable.Type)
	}
}

func isIntegerValue(value any) bool {
	if value == nil {
		return false
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func containsOption(options []string, value string) bool {
	for _, option := range options {
		if option == value {
			return true
		}
	}
	return false
}

func normalizeStringSlice(value any) ([]string, bool) {
	switch val := value.(type) {
	case []string:
		return val, true
	case []any:
		result := make([]string, len(val))
		for i, item := range val {
			s, ok := item.(string)
			if !ok {
				return nil, false
			}
			result[i] = s
		}
		return result, true
	default:
		return nil, false
	}
}

// validateIncludes validates that features and components do not include projects.
func (v *Validator) validateIncludes(node *TemplateNode) error {
	if node.Template.Type == TypeProject {
		return nil
	}

	for _, child := range node.Children {
		if child.Template.Type == TypeProject {
			return fmt.Errorf("%s %q cannot include project %q", node.Template.Type, node.Template.Name, child.Template.Name)
		}
	}

	return nil
}

// validateProjectNameRole validates that project templates have exactly one
// variable with role: project_name.
func (v *Validator) validateProjectNameRole(tmpl *Template) error {
	// Only project templates require a project_name role
	if tmpl.Type != TypeProject {
		return nil
	}

	count := 0
	for _, variable := range tmpl.Variables {
		if variable.Role == RoleProjectName {
			count++
		}
	}

	switch count {
	case 0:
		return fmt.Errorf("project template %q must have exactly one variable with role %q", tmpl.Name, RoleProjectName)
	case 1:
		return nil
	default:
		return fmt.Errorf("project template %q has %d variables with role %q, but must have exactly one", tmpl.Name, count, RoleProjectName)
	}
}

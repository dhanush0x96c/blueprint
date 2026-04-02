package template

import (
	"errors"
	"fmt"
	"io/fs"
	"path"

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

// ValidateTreeContexts recursively validates that all required variables are present
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

// ValidateContext validates that all required variables are present in the context.
func (v *Validator) ValidateContext(tmpl *Template, ctx *Context) error {
	for _, variable := range tmpl.Variables {
		_, exists := ctx.Get(variable.Name)
		if !exists && variable.Default == nil {
			return fmt.Errorf("required variable %s is missing", variable.Name)
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

		// Options must be non-empty for select/multiselect
		if variable.Type == VariableTypeSelect || variable.Type == VariableTypeMultiSelect {
			if len(variable.Options) == 0 {
				errs = append(errs, fmt.Errorf("variable[%d] %q: options required for type %s", i, variable.Name, variable.Type))
			}
		}
	}

	return errs
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

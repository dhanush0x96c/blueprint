package template

import (
	"errors"
	"fmt"

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

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
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

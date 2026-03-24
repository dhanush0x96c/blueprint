package app

import "fmt"

// TemplateNotFoundError is returned when a template is not found.
type TemplateNotFoundError struct {
	Name string
}

func (e *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template not found: %s", e.Name)
}

// InvalidTemplateTypeError is returned when an invalid template type is provided.
type InvalidTemplateTypeError struct {
	Type string
}

func (e *InvalidTemplateTypeError) Error() string {
	return fmt.Sprintf("invalid template type: %s", e.Type)
}

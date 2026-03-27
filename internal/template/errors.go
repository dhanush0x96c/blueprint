package template

import "fmt"

// TemplateNotFoundError is returned when a template is not found.
type TemplateNotFoundError struct {
	Name string
}

func (e *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template not found: %s", e.Name)
}

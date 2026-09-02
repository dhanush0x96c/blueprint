package template

import "fmt"

// NotFoundError is returned when a template is not found.
type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("template not found: %s", e.Name)
}

package app

import "fmt"

// InvalidTemplateTypeError is returned when an invalid template type is provided.
type InvalidTemplateTypeError struct {
	Type string
}

func (e *InvalidTemplateTypeError) Error() string {
	return fmt.Sprintf("invalid template type: %s", e.Type)
}

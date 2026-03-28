package cli

import (
	"strings"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// ValidateTemplateTypeArg checks if the given argument is a valid template type filter.
func ValidateTemplateTypeArg(arg string) (template.Type, error) {
	switch strings.ToLower(arg) {
	case "projects":
		return template.TypeProject, nil
	case "features":
		return template.TypeFeature, nil
	case "components":
		return template.TypeComponent, nil
	default:
		return "", &InvalidTemplateTypeError{Type: arg}
	}
}

package prompt

import (
	"strconv"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// Variable extends a template variable with a collected value.
type Variable struct {
	template.Variable
	Value any
}

// VariableGroup is a set of variables prompted together.
type VariableGroup struct {
	Title     string
	Variables []Variable
}

// CastValue safely casts a validated variable value to the requested type.
func CastValue[T any](value any) T {
	var zero T
	if value == nil {
		return zero
	}

	return value.(T)
}

func extractValue(valuePtr any, varType template.VariableType) any {
	switch varType {
	case template.VariableTypeString, template.VariableTypeSelect:
		return *CastValue[*string](valuePtr)
	case template.VariableTypeInt:
		value := CastValue[*string](valuePtr)
		if *value == "" {
			return 0
		}

		parsed, _ := strconv.Atoi(*value)
		return parsed
	case template.VariableTypeBool:
		return *CastValue[*bool](valuePtr)
	case template.VariableTypeMultiSelect:
		return *CastValue[*[]string](valuePtr)
	default:
		return valuePtr
	}
}

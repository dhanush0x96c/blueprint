package app

import "errors"

// ErrTemplateNotFound is returned when a template is not found.
var ErrTemplateNotFound = errors.New("template not found")

// ErrInvalidTemplateType is returned when an invalid template type is provided.
var ErrInvalidTemplateType = errors.New("invalid template type")

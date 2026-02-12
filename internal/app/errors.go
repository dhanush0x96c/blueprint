package app

import "errors"

// ErrTemplateNotFound is returned when a template is not found.
var ErrTemplateNotFound = errors.New("template not found")

package templates

import "embed"

// Templates is an embedded file system containing the builtin templates.
//
//go:embed all:*
var Templates embed.FS

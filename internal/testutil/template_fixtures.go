package testutil

const (
	ValidProjectTemplate = `
name: go-cli
type: project
version: "1.0.0"
description: "Go CLI project"
variables:
  - name: app_name
    prompt: "App name?"
    type: string
    role: project_name
`

	ValidFeatureTemplate = `
name: testing
type: feature
version: "1.0.0"
description: "Testing support"
`

	InvalidTemplate = `
name:
type: project
`

	ValidTemplateWithTags = `
name: go-api
type: project
version: "1.0.0"
description: "Go API project"
tags: ["go", "api"]
variables:
  - name: app_name
    prompt: "App name?"
    type: string
    role: project_name
`

	ValidFeatureTemplateWithTags = `
name: auth
type: feature
version: "1.0.0"
description: "Authentication"
tags: ["auth", "security"]
`

	TemplateWithTags = `
name: tagged-template
type: project
version: "1.0.0"
description: "Template with tags"
tags: ["go", "cli", "testing"]
variables:
  - name: app_name
    prompt: "App name?"
    type: string
    role: project_name
`

	TemplateWithoutTags = `
name: no-tags
type: feature
version: "1.0.0"
description: "Template without tags"
`
)

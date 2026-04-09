package cmd

import (
	"fmt"
	"strings"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/scaffold"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/ui"
	"github.com/dhanush0x96c/blueprint/internal/vars"
	"github.com/spf13/cobra"
)

func NewInitCmd(appCtx *app.Context) *cobra.Command {
	var (
		force        bool
		yes          bool
		varFlags     []string
		includeFlags []string
		excludeFlags []string
	)

	cmd := &cobra.Command{
		Use:   "init <template> [output-dir]",
		Short: "Initialize a new project",
		Long:  `Initialize a new project from a template.`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]

			var outputDir string
			if len(args) > 1 {
				outputDir = args[1]
			}

			vars, err := parseVarFlags(varFlags)
			if err != nil {
				return err
			}

			enabledIncludes, err := parseIncludeFlags(includeFlags, excludeFlags)
			if err != nil {
				return err
			}

			scaffolder := scaffold.NewScaffolder(appCtx.Resolver)
			result, err := scaffolder.Scaffold(scaffold.Options{
				TemplateRef: template.TemplateRef{
					Name: templateName,
				},
				OutputDir:       outputDir,
				Variables:       vars,
				EnabledIncludes: enabledIncludes,
				Interactive:     !yes,
				DryRun:          appCtx.Options.DryRun,
				Overwrite:       force,
			})

			if err != nil {
				return fmt.Errorf("init template %q: %w", templateName, err)
			}

			ui.RenderResult(result)

			return nil
		},
	}

	cmd.Flags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Overwrite existing files if they exist",
	)

	cmd.Flags().BoolVarP(
		&yes,
		"yes",
		"y",
		false,
		"Accept defaults and disable prompts",
	)

	cmd.Flags().StringArrayVar(
		&varFlags,
		"var",
		nil,
		`Set a template variable (format: key=value)`,
	)

	cmd.Flags().StringArrayVar(
		&includeFlags,
		"include",
		nil,
		`Include a template feature (format: template-name)`,
	)

	cmd.Flags().StringArrayVar(
		&excludeFlags,
		"exclude",
		nil,
		`Exclude a template feature (format: template-name)`,
	)

	return cmd
}

func parseVarFlags(flags []string) (vars.Variables, error) {
	vars := vars.Variables{
		Global:       make(map[string]string),
		NameSpecific: make(map[string]map[string]string),
		NodeSpecific: make(map[string]map[string]string),
	}

	if len(flags) == 0 {
		return vars, nil
	}

	for _, f := range flags {
		scope, key, value, err := parseVarFlag(f)
		if err != nil {
			return vars, err
		}

		if strings.HasPrefix(scope, "#") {
			nodeID := scope[1:]
			if vars.NodeSpecific[nodeID] == nil {
				vars.NodeSpecific[nodeID] = make(map[string]string)
			}
			vars.NodeSpecific[nodeID][key] = value
			continue
		}

		if scope != "" {
			if vars.NameSpecific[scope] == nil {
				vars.NameSpecific[scope] = make(map[string]string)
			}
			vars.NameSpecific[scope][key] = value
			continue
		}

		vars.Global[key] = value
	}

	return vars, nil
}

func parseVarFlag(flag string) (scope, key, value string, err error) {
	left, value, ok := strings.Cut(flag, "=")
	if !ok {
		return "", "", "", fmt.Errorf("invalid variable format %q: expected key=value", flag)
	}

	if !strings.Contains(left, ":") {
		return "", left, value, nil
	}

	scope, key, ok = strings.Cut(left, ":")
	if !ok || key == "" {
		return "", "", "", fmt.Errorf("invalid variable format %q: expected scope:key=value", flag)
	}

	return scope, key, value, nil
}

func parseIncludeFlags(includeFlags, excludeFlags []string) (map[string]bool, error) {
	if len(includeFlags) == 0 && len(excludeFlags) == 0 {
		return nil, nil
	}

	includes := make(map[string]bool)

	// Process include flags
	for _, name := range includeFlags {
		if name == "" {
			return nil, fmt.Errorf("invalid include flag: empty template name")
		}
		includes[name] = true
	}

	// Process exclude flags
	for _, name := range excludeFlags {
		if name == "" {
			return nil, fmt.Errorf("invalid exclude flag: empty template name")
		}
		if _, exists := includes[name]; exists {
			return nil, fmt.Errorf("template %q cannot be both included and excluded", name)
		}
		includes[name] = false
	}

	return includes, nil
}

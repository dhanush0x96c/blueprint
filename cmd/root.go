package cmd

import (
	"fmt"

	"github.com/dhanush0x96c/blueprint/internal/app"
	"github.com/dhanush0x96c/blueprint/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cfgLoader := config.Loader{
		EnvPrefix: "BLUEPRINT",
		CLIArgs:   map[string]string{},
	}
	appCtx := &app.Context{}

	cmd := &cobra.Command{
		Use:     "blueprint",
		Aliases: []string{"bp"},
		Short:   "Universal project scaffolding",
		Long:    "Blueprint scaffolds projects from composable templates.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			appCtx.Config = cfg

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(
		&cfgLoader.ConfigFile,
		"config",
		"",
		fmt.Sprintf("config file (default is %s)", config.DefaultPathHint()),
	)

	return cmd
}

func Execute() error {
	if err := NewRootCmd().Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

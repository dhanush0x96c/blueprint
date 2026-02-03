package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCmd() *cobra.Command {
	var cfgFile string
	cmd := &cobra.Command{
		Use:     "blueprint",
		Aliases: []string{"bp"},
		Short:   "Universal project scaffolding",
		Long:    "Blueprint scaffolds projects from composable templates.",
	}
	cobra.OnInitialize(func() {
		initConfig(cfgFile)
	})

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.blueprint.yaml)")

	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	return cmd
}

func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initConfig(cfgFile string) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".blueprint")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

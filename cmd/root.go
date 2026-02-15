package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SisyphusSQ/go-cli-starter/vars"
)

var rootCmd = &cobra.Command{
	Use:   vars.AppName,
	Short: "Generate Go CLI starter projects",
	Long:  "go-cli-starter generates a ready-to-use Go CLI project template.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func initAll() {
	initVersion()
	initInit()
	initNew()
}

func Execute() {
	initAll()
	rootCmd.SilenceErrors = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

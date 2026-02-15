package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SisyphusSQ/go-cli-starter/internal/scaffold"
)

var (
	initModuleNameFlag string
	initBinaryNameFlag string
	initMinimalFlag    bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project in the current empty directory",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := buildTemplateData(
			".",
			initModuleNameFlag,
			initBinaryNameFlag,
			initMinimalFlag,
		)
		if err != nil {
			return err
		}

		if err := scaffold.Generate(".", data); err != nil {
			return fmt.Errorf("initialize project: %w", err)
		}

		fmt.Println("Project initialized in current directory")
		return nil
	},
}

func initInit() {
	initCmd.Flags().StringVarP(
		&initModuleNameFlag,
		"module",
		"m",
		"",
		"Go module path (default: example.com/<directory-name>)",
	)
	initCmd.Flags().StringVarP(
		&initBinaryNameFlag,
		"binary",
		"b",
		"",
		"Binary name (default: inferred from directory name)",
	)
	initCmd.Flags().BoolVar(
		&initMinimalFlag,
		"minimal",
		false,
		"Initialize minimal project with only root/version commands",
	)

	rootCmd.AddCommand(initCmd)
}

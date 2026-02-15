package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SisyphusSQ/go-cli-starter/internal/scaffold"
)

var (
	moduleNameFlag string
	binaryNameFlag string
	minimalFlag    bool
)

var newCmd = &cobra.Command{
	Use:   "new <output-dir>",
	Short: "Generate a new CLI project from template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir := args[0]
		data, err := buildTemplateData(
			outputDir,
			moduleNameFlag,
			binaryNameFlag,
			minimalFlag,
		)
		if err != nil {
			return err
		}

		if err := scaffold.Generate(outputDir, data); err != nil {
			return fmt.Errorf("generate project: %w", err)
		}

		fmt.Printf("Project generated at %s\n", outputDir)
		return nil
	},
}

func initNew() {
	newCmd.Flags().StringVarP(
		&moduleNameFlag,
		"module",
		"m",
		"",
		"Go module path (default: example.com/<directory-name>)",
	)
	newCmd.Flags().StringVarP(
		&binaryNameFlag,
		"binary",
		"b",
		"",
		"Binary name (default: inferred from directory name)",
	)
	newCmd.Flags().BoolVar(
		&minimalFlag,
		"minimal",
		false,
		"Generate minimal project with only root/version commands",
	)

	rootCmd.AddCommand(newCmd)
}

func buildTemplateData(
	outputDir string,
	moduleValue string,
	binaryValue string,
	minimal bool,
) (scaffold.TemplateData, error) {
	projectName, err := inferProjectName(outputDir)
	if err != nil {
		return scaffold.TemplateData{}, err
	}

	moduleName := strings.TrimSpace(moduleValue)
	if moduleName == "" {
		moduleName = defaultModuleName(projectName)
	}

	binaryName := strings.TrimSpace(binaryValue)
	if binaryName == "" {
		binaryName = projectName
	}

	return scaffold.TemplateData{
		ModuleName:  moduleName,
		BinaryName:  binaryName,
		ProjectName: projectName,
		Minimal:     minimal,
	}, nil
}

func inferProjectName(outputDir string) (string, error) {
	cleanedOutputDir := filepath.Clean(strings.TrimSpace(outputDir))
	if cleanedOutputDir == "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current directory: %w", err)
		}
		return sanitizeDerivedProjectName(filepath.Base(wd), wd)
	}

	return sanitizeDerivedProjectName(
		filepath.Base(cleanedOutputDir),
		cleanedOutputDir,
	)
}

func sanitizeDerivedProjectName(projectName, source string) (string, error) {
	name := strings.TrimSpace(projectName)
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf(
			"cannot infer project name from %q: use a non-root/non-dot output directory or specify --module and --binary explicitly",
			source,
		)
	}
	if strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf(
			"cannot infer project name from %q: derived name %q is invalid",
			source,
			name,
		)
	}

	return name, nil
}

func defaultModuleName(projectName string) string {
	return fmt.Sprintf("example.com/%s", projectName)
}

package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/mod/module"
)

const templateRoot = "_template"
const fallbackGoVersion = "1.26.0"

type TemplateData struct {
	ModuleName  string
	BinaryName  string
	ProjectName string
	GoVersion   string
	Minimal     bool
}

var (
	binaryNamePattern    = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	projectNamePattern   = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	goVersionPattern     = regexp.MustCompile(`^\d+\.\d+(?:\.\d+)?$`)
	goVersionExtractExpr = regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)
)

func (d TemplateData) Validate() error {
	if err := validateModulePath(d.ModuleName); err != nil {
		return err
	}
	if err := validateBinaryName(d.BinaryName); err != nil {
		return err
	}
	if err := validateProjectName(d.ProjectName); err != nil {
		return err
	}
	goVersion := strings.TrimSpace(d.GoVersion)
	if goVersion == "" {
		goVersion = defaultGoVersion()
	}
	if err := validateGoVersion(goVersion); err != nil {
		return err
	}

	return nil
}

func Generate(outputDir string, data TemplateData) error {
	data.applyDefaults()
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid template data: %w", err)
	}

	if err := prepareOutputDir(outputDir); err != nil {
		return err
	}

	if err := fs.WalkDir(templateFS, templateRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk template path %s: %w", path, walkErr)
		}
		if path == templateRoot {
			return nil
		}
		if shouldSkipTemplate(path, data) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		relPath := strings.TrimPrefix(path, templateRoot+"/")
		outRelPath := strings.TrimSuffix(relPath, ".tmpl")
		outPath := filepath.Join(outputDir, filepath.FromSlash(outRelPath))

		if d.IsDir() {
			if err := os.MkdirAll(outPath, 0o755); err != nil {
				return fmt.Errorf("create directory %s: %w", outPath, err)
			}
			return nil
		}

		raw, err := templateFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template file %s: %w", path, err)
		}
		rendered, err := renderTemplate(path, raw, data)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("create output parent directory %s: %w", outPath, err)
		}
		if err := os.WriteFile(outPath, rendered, 0o644); err != nil {
			return fmt.Errorf("write output file %s: %w", outPath, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walk templates: %w", err)
	}

	return nil
}

func (d *TemplateData) applyDefaults() {
	if strings.TrimSpace(d.GoVersion) == "" {
		d.GoVersion = defaultGoVersion()
	}
}

func defaultGoVersion() string {
	matches := goVersionExtractExpr.FindStringSubmatch(runtime.Version())
	if len(matches) != 2 {
		return fallbackGoVersion
	}

	return matches[1]
}

func shouldSkipTemplate(path string, data TemplateData) bool {
	if !data.Minimal {
		return false
	}

	switch path {
	case templateRoot + "/cmd/greet.go.tmpl",
		templateRoot + "/cmd/mcp.go.tmpl":
		return true
	default:
		return path == templateRoot+"/internal" ||
			strings.HasPrefix(path, templateRoot+"/internal/")
	}
}

func prepareOutputDir(outputDir string) error {
	info, err := os.Stat(outputDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("output path is not a directory: %s", outputDir)
		}

		entries, err := os.ReadDir(outputDir)
		if err != nil {
			return fmt.Errorf("read output directory %s: %w", outputDir, err)
		}
		if hasVisibleEntries(entries) {
			return fmt.Errorf("output directory is not empty: %s", outputDir)
		}

		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("stat output directory %s: %w", outputDir, err)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory %s: %w", outputDir, err)
	}

	return nil
}

func hasVisibleEntries(entries []os.DirEntry) bool {
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		return true
	}

	return false
}

func validateModulePath(moduleName string) error {
	v := strings.TrimSpace(moduleName)
	if v == "" {
		return fmt.Errorf("module cannot be empty")
	}
	if strings.ContainsAny(v, " \t\r\n") {
		return fmt.Errorf("module cannot contain whitespace: %q", moduleName)
	}
	if strings.Contains(v, `\`) {
		return fmt.Errorf("module cannot contain backslash: %q", moduleName)
	}
	if strings.ContainsAny(v, `"'`) {
		return fmt.Errorf("module cannot contain quotes: %q", moduleName)
	}
	if err := module.CheckPath(v); err != nil {
		return fmt.Errorf("invalid module path %q: %w", moduleName, err)
	}

	return nil
}

func validateBinaryName(binaryName string) error {
	v := strings.TrimSpace(binaryName)
	if v == "" {
		return fmt.Errorf("binary cannot be empty")
	}
	if strings.ContainsAny(v, " \t\r\n") {
		return fmt.Errorf("binary cannot contain whitespace: %q", binaryName)
	}
	if strings.ContainsAny(v, `/\`) {
		return fmt.Errorf("binary cannot contain path separators: %q", binaryName)
	}
	if strings.ContainsAny(v, `"'`) {
		return fmt.Errorf("binary cannot contain quotes: %q", binaryName)
	}
	if v == "." || v == ".." {
		return fmt.Errorf("binary cannot be %q", binaryName)
	}
	if !binaryNamePattern.MatchString(v) {
		return fmt.Errorf(
			"binary contains invalid characters: %q (allowed: letters, digits, dot, underscore, hyphen)",
			binaryName,
		)
	}

	return nil
}

func validateProjectName(projectName string) error {
	v := strings.TrimSpace(projectName)
	if v == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if v == "." || v == ".." {
		return fmt.Errorf("project name cannot be %q", projectName)
	}
	if strings.ContainsAny(v, `/\`) {
		return fmt.Errorf("project name cannot contain path separators: %q", projectName)
	}
	if strings.ContainsAny(v, " \t\r\n") {
		return fmt.Errorf("project name cannot contain whitespace: %q", projectName)
	}
	if strings.ContainsAny(v, `"'`) {
		return fmt.Errorf("project name cannot contain quotes: %q", projectName)
	}
	if !projectNamePattern.MatchString(v) {
		return fmt.Errorf(
			"project name contains invalid characters: %q (allowed: letters, digits, dot, underscore, hyphen)",
			projectName,
		)
	}

	return nil
}

func validateGoVersion(goVersion string) error {
	v := strings.TrimSpace(goVersion)
	if v == "" {
		return fmt.Errorf("go version cannot be empty")
	}
	if !goVersionPattern.MatchString(v) {
		return fmt.Errorf(
			"go version contains invalid format: %q (example: 1.26.0)",
			goVersion,
		)
	}

	return nil
}

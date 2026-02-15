package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInferProjectNameFromCurrentDirectory(t *testing.T) {
	baseDir := t.TempDir()
	projectDir := filepath.Join(baseDir, "demo-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir project dir: %v", err)
	}

	projectName, err := inferProjectName(".")
	if err != nil {
		t.Fatalf("inferProjectName(.) error = %v", err)
	}
	if projectName != "demo-app" {
		t.Fatalf("inferProjectName(.) = %q, want %q", projectName, "demo-app")
	}
}

func TestInferProjectNameRejectsIllegalDefaults(t *testing.T) {
	tests := []struct {
		name      string
		outputDir string
	}{
		{
			name:      "root path",
			outputDir: string(filepath.Separator),
		},
		{
			name:      "parent directory path",
			outputDir: "..",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := inferProjectName(tt.outputDir)
			if err == nil {
				t.Fatalf("inferProjectName(%q) expected error, got nil", tt.outputDir)
			}
			if !strings.Contains(err.Error(), "cannot infer project name") {
				t.Fatalf("inferProjectName(%q) error = %v", tt.outputDir, err)
			}
		})
	}
}

func TestDefaultModuleName(t *testing.T) {
	moduleName := defaultModuleName("demo-app")
	if moduleName != "example.com/demo-app" {
		t.Fatalf(
			"defaultModuleName() = %q, want %q",
			moduleName,
			"example.com/demo-app",
		)
	}
}

func TestBuildTemplateDataWithDefaultsAndMinimal(t *testing.T) {
	baseDir := t.TempDir()
	projectDir := filepath.Join(baseDir, "mini-cli")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	data, err := buildTemplateData(projectDir, "", "", true)
	if err != nil {
		t.Fatalf("buildTemplateData() error = %v", err)
	}

	if data.ProjectName != "mini-cli" {
		t.Fatalf("ProjectName = %q, want %q", data.ProjectName, "mini-cli")
	}
	if data.BinaryName != "mini-cli" {
		t.Fatalf("BinaryName = %q, want %q", data.BinaryName, "mini-cli")
	}
	if data.ModuleName != "example.com/mini-cli" {
		t.Fatalf("ModuleName = %q, want %q", data.ModuleName, "example.com/mini-cli")
	}
	if !data.Minimal {
		t.Fatal("Minimal = false, want true")
	}
}

func TestBuildTemplateDataWithCustomFlags(t *testing.T) {
	data, err := buildTemplateData(
		"./anything",
		"github.com/acme/cli",
		"acme-cli",
		false,
	)
	if err != nil {
		t.Fatalf("buildTemplateData() error = %v", err)
	}

	if data.ModuleName != "github.com/acme/cli" {
		t.Fatalf("ModuleName = %q, want %q", data.ModuleName, "github.com/acme/cli")
	}
	if data.BinaryName != "acme-cli" {
		t.Fatalf("BinaryName = %q, want %q", data.BinaryName, "acme-cli")
	}
	if data.Minimal {
		t.Fatal("Minimal = true, want false")
	}
}

func TestInitCommandInEmptyDir(t *testing.T) {
	baseDir := t.TempDir()
	projectDir := filepath.Join(baseDir, "init-demo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := os.Mkdir(filepath.Join(projectDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git dir: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir project dir: %v", err)
	}

	origModule := initModuleNameFlag
	origBinary := initBinaryNameFlag
	origMinimal := initMinimalFlag
	t.Cleanup(func() {
		initModuleNameFlag = origModule
		initBinaryNameFlag = origBinary
		initMinimalFlag = origMinimal
	})

	initModuleNameFlag = ""
	initBinaryNameFlag = ""
	initMinimalFlag = true

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("initCmd.RunE() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectDir, "main.go")); err != nil {
		t.Fatalf("expected main.go to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "cmd", "root.go")); err != nil {
		t.Fatalf("expected cmd/root.go to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "cmd", "greet.go")); !os.IsNotExist(err) {
		t.Fatalf("expected cmd/greet.go to be absent in minimal mode, got err=%v", err)
	}
}

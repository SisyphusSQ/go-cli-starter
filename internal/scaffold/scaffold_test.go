package scaffold

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const commandTimeout = 2 * time.Minute

func TestGenerate(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "my-new-cli")
	data := TemplateData{
		ModuleName:  "github.com/test/my-new-cli",
		BinaryName:  "my-new-cli",
		ProjectName: "my-new-cli",
	}

	if err := Generate(outputDir, data); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	goModPath := filepath.Join(outputDir, "go.mod")
	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("read generated go.mod: %v", err)
	}
	if !strings.Contains(string(goModContent), "module github.com/test/my-new-cli") {
		t.Fatalf("go.mod does not contain expected module path: %s", goModContent)
	}
	if !strings.Contains(string(goModContent), "\ngo "+defaultGoVersion()+"\n") {
		t.Fatalf("go.mod does not contain expected go version directive: %s", goModContent)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "main.go")); err != nil {
		t.Fatalf("generated main.go missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "changeLog.md")); err != nil {
		t.Fatalf("generated changeLog.md missing: %v", err)
	}
	requiredUtils := []string{
		filepath.Join("utils", "retry", "retry.go"),
		filepath.Join("utils", "httputil", "client.go"),
		filepath.Join("utils", "fileutil", "fileutil.go"),
	}
	for _, relPath := range requiredUtils {
		if _, err := os.Stat(filepath.Join(outputDir, relPath)); err != nil {
			t.Fatalf("generated %s missing: %v", relPath, err)
		}
	}
	licenseContent, err := os.ReadFile(filepath.Join(outputDir, "LICENSE"))
	if err != nil {
		t.Fatalf("generated LICENSE missing: %v", err)
	}
	if !strings.Contains(string(licenseContent), "MIT License") {
		t.Fatalf("generated LICENSE does not contain MIT header: %s", licenseContent)
	}
}

func TestGenerateRejectsNonEmptyOutputDir(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "existing")

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "placeholder.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write placeholder file: %v", err)
	}

	err := Generate(outputDir, TemplateData{
		ModuleName:  "github.com/test/sample",
		BinaryName:  "sample",
		ProjectName: "sample",
	})
	if err == nil {
		t.Fatal("Generate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("Generate() unexpected error: %v", err)
	}
}

func TestGenerateAllowsHiddenEntriesOnly(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "existing")

	if err := os.MkdirAll(filepath.Join(outputDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git dir: %v", err)
	}

	err := Generate(outputDir, TemplateData{
		ModuleName:  "github.com/test/sample",
		BinaryName:  "sample",
		ProjectName: "sample",
	})
	if err != nil {
		t.Fatalf("Generate() expected success with hidden-only entries, got error: %v", err)
	}
}

func TestGenerateRejectsInvalidTemplateData(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "out")
	err := Generate(outputDir, TemplateData{
		ModuleName:  "bad module",
		BinaryName:  "my-cli",
		ProjectName: "my-cli",
	})
	if err == nil {
		t.Fatal("Generate() expected error for invalid module, got nil")
	}
	if !strings.Contains(err.Error(), "invalid template data") {
		t.Fatalf("Generate() unexpected error: %v", err)
	}
}

func TestTemplateDataValidateRejectsInvalidNames(t *testing.T) {
	tests := []struct {
		name      string
		data      TemplateData
		errSubstr string
	}{
		{
			name: "module missing domain",
			data: TemplateData{
				ModuleName:  "my-cli",
				BinaryName:  "my-cli",
				ProjectName: "my-cli",
			},
			errSubstr: "invalid module path",
		},
		{
			name: "module has quote",
			data: TemplateData{
				ModuleName:  `github.com/test/"my-cli"`,
				BinaryName:  "my-cli",
				ProjectName: "my-cli",
			},
			errSubstr: "module cannot contain quotes",
		},
		{
			name: "binary has path separator",
			data: TemplateData{
				ModuleName:  "github.com/test/my-cli",
				BinaryName:  "cmd/my-cli",
				ProjectName: "my-cli",
			},
			errSubstr: "binary cannot contain path separators",
		},
		{
			name: "binary has quote",
			data: TemplateData{
				ModuleName:  "github.com/test/my-cli",
				BinaryName:  `my"cli`,
				ProjectName: "my-cli",
			},
			errSubstr: "binary cannot contain quotes",
		},
		{
			name: "binary has invalid punctuation",
			data: TemplateData{
				ModuleName:  "github.com/test/my-cli",
				BinaryName:  "my*cli",
				ProjectName: "my-cli",
			},
			errSubstr: "binary contains invalid characters",
		},
		{
			name: "project name has quote",
			data: TemplateData{
				ModuleName:  "github.com/test/my-cli",
				BinaryName:  "my-cli",
				ProjectName: `my"cli`,
			},
			errSubstr: "project name cannot contain quotes",
		},
		{
			name: "project name has invalid punctuation",
			data: TemplateData{
				ModuleName:  "github.com/test/my-cli",
				BinaryName:  "my-cli",
				ProjectName: "my*cli",
			},
			errSubstr: "project name contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if err == nil {
				t.Fatal("Validate() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.errSubstr)
			}
		})
	}
}

func TestGenerateE2EProjectBuildAndMCPHelp(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "demo-e2e")
	data := TemplateData{
		ModuleName:  "github.com/test/demo-e2e",
		BinaryName:  "demo-e2e",
		ProjectName: "demo-e2e",
	}

	if err := Generate(outputDir, data); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if stdout, stderr, err := runGoCommand(outputDir, "mod", "tidy"); err != nil {
		t.Fatalf(
			"go mod tidy failed: %v\nstdout:\n%s\nstderr:\n%s",
			err,
			stdout,
			stderr,
		)
	}

	if stdout, stderr, err := runGoCommand(outputDir, "build", "./..."); err != nil {
		t.Fatalf("go build failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	stdout, stderr, err := runGoCommand(outputDir, "run", ".", "mcp", "--help")
	if err != nil {
		t.Fatalf("go run . mcp --help failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	helpOutput := stdout + stderr
	if !strings.Contains(helpOutput, "Start an MCP stdio server") {
		t.Fatalf("mcp --help output missing expected text:\n%s", helpOutput)
	}
}

func TestGenerateMinimalMode(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "demo-minimal")
	data := TemplateData{
		ModuleName:  "github.com/test/demo-minimal",
		BinaryName:  "demo-minimal",
		ProjectName: "demo-minimal",
		Minimal:     true,
	}

	if err := Generate(outputDir, data); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "cmd", "root.go")); err != nil {
		t.Fatalf("expected cmd/root.go to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "cmd", "version.go")); err != nil {
		t.Fatalf("expected cmd/version.go to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "cmd", "greet.go")); !os.IsNotExist(err) {
		t.Fatalf("expected cmd/greet.go to be absent, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "cmd", "mcp.go")); !os.IsNotExist(err) {
		t.Fatalf("expected cmd/mcp.go to be absent, got err=%v", err)
	}
	for _, relPath := range []string{
		filepath.Join("utils", "retry", "retry.go"),
		filepath.Join("utils", "httputil", "client.go"),
		filepath.Join("utils", "fileutil", "fileutil.go"),
	} {
		if _, err := os.Stat(filepath.Join(outputDir, relPath)); err != nil {
			t.Fatalf("expected %s to exist in minimal mode: %v", relPath, err)
		}
	}
	licenseContent, err := os.ReadFile(filepath.Join(outputDir, "LICENSE"))
	if err != nil {
		t.Fatalf("generated LICENSE missing: %v", err)
	}
	if !strings.Contains(string(licenseContent), "MIT License") {
		t.Fatalf("generated LICENSE does not contain MIT header: %s", licenseContent)
	}

	goModContent, err := os.ReadFile(filepath.Join(outputDir, "go.mod"))
	if err != nil {
		t.Fatalf("read generated go.mod: %v", err)
	}
	if strings.Contains(string(goModContent), "github.com/mark3labs/mcp-go") {
		t.Fatalf("minimal mode go.mod should not contain mcp-go dependency:\n%s", goModContent)
	}

	if stdout, stderr, err := runGoCommand(outputDir, "mod", "tidy"); err != nil {
		t.Fatalf(
			"go mod tidy failed: %v\nstdout:\n%s\nstderr:\n%s",
			err,
			stdout,
			stderr,
		)
	}
	if stdout, stderr, err := runGoCommand(outputDir, "build", "./..."); err != nil {
		t.Fatalf("go build failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	stdout, stderr, err := runGoCommand(outputDir, "run", ".", "mcp", "--help")
	if err == nil {
		t.Fatalf(
			"minimal mode should not expose mcp command\nstdout:\n%s\nstderr:\n%s",
			stdout,
			stderr,
		)
	}
	if !strings.Contains(stdout+stderr, "unknown command") {
		t.Fatalf("expected unknown command error, got:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGenerateMCPStdioSmoke(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "demo-mcp-smoke")
	data := TemplateData{
		ModuleName:  "github.com/test/demo-mcp-smoke",
		BinaryName:  "demo-mcp-smoke",
		ProjectName: "demo-mcp-smoke",
	}

	if err := Generate(outputDir, data); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if stdout, stderr, err := runGoCommand(outputDir, "mod", "tidy"); err != nil {
		t.Fatalf(
			"go mod tidy failed: %v\nstdout:\n%s\nstderr:\n%s",
			err,
			stdout,
			stderr,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", ".", "mcp")
	cmd.Dir = outputDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	var stderrBuf bytes.Buffer
	stderrDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&stderrBuf, stderrPipe)
		close(stderrDone)
	}()

	responseCh := make(chan string, 1)
	responseErrCh := make(chan error, 1)
	go func() {
		body, readErr := readJSONRPCFrame(stdoutPipe)
		if readErr != nil {
			responseErrCh <- readErr
			return
		}
		responseCh <- body
	}()

	if err := cmd.Start(); err != nil {
		t.Fatalf("start mcp command: %v", err)
	}

	requestBody := `{"jsonrpc":"2.0","id":1,"method":"unknown.method","params":{}}`
	if _, err := fmt.Fprintf(
		stdin,
		"Content-Length: %d\r\n\r\n%s",
		len(requestBody),
		requestBody,
	); err != nil {
		t.Fatalf("write json-rpc request: %v", err)
	}
	_ = stdin.Close()

	var responseBody string
	select {
	case responseBody = <-responseCh:
	case err := <-responseErrCh:
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		<-stderrDone
		t.Fatalf("read json-rpc response: %v", err)
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		<-stderrDone
		t.Fatal("timed out waiting for json-rpc response")
	}

	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	<-stderrDone

	if ctx.Err() != nil {
		t.Fatalf("mcp command context error: %v", ctx.Err())
	}

	var response map[string]any
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		t.Fatalf("json-rpc response is not valid JSON: %v\nbody: %s", err, responseBody)
	}
	if response["jsonrpc"] != "2.0" {
		t.Fatalf("json-rpc version mismatch: %#v", response["jsonrpc"])
	}
	if _, hasID := response["id"]; !hasID {
		if _, hasMethod := response["method"]; !hasMethod {
			t.Fatalf("json-rpc payload missing both id and method: %#v", response)
		}
	}

	stderrOutput := stderrBuf.String()
	if !strings.Contains(stderrOutput, "starting mcp stdio server") {
		t.Fatalf("stderr missing mcp startup log:\n%s", stderrOutput)
	}
	if strings.Contains(responseBody, "starting mcp stdio server") {
		t.Fatalf("stdout protocol payload unexpectedly contains log text: %s", responseBody)
	}
}

func runGoCommand(dir string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() != nil {
		return stdout.String(), stderr.String(), fmt.Errorf(
			"go %s: %w",
			strings.Join(args, " "),
			ctx.Err(),
		)
	}

	return stdout.String(), stderr.String(), err
}

func readJSONRPCFrame(r io.Reader) (string, error) {
	reader := bufio.NewReader(r)
	firstLine, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read protocol output: %w", err)
	}

	firstLine = strings.TrimSpace(firstLine)
	if firstLine == "" {
		return "", fmt.Errorf("empty stdout response")
	}
	if strings.HasPrefix(firstLine, "{") {
		return firstLine, nil
	}

	headers := []string{firstLine}
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return "", fmt.Errorf("read frame header: %w", readErr)
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		headers = append(headers, line)
	}

	contentLength := -1
	for _, line := range headers {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("unexpected non-protocol output on stdout: %q", line)
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "content-length":
			n, convErr := strconv.Atoi(value)
			if convErr != nil {
				return "", fmt.Errorf("parse content-length %q: %w", value, convErr)
			}
			contentLength = n
		case "content-type":
			// optional header, intentionally ignored
		default:
			return "", fmt.Errorf("unexpected stdout header %q", parts[0])
		}
	}

	if contentLength <= 0 {
		return "", fmt.Errorf("missing or invalid content-length header")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, body); err != nil {
		return "", fmt.Errorf("read frame body: %w", err)
	}

	return string(body), nil
}

package scaf_fold

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
	if _, err := os.Stat(filepath.Join(outputDir, "README_CN.md")); err != nil {
		t.Fatalf("generated README_CN.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "changeLog.md")); err != nil {
		t.Fatalf("generated changeLog.md missing: %v", err)
	}
	requiredUtils := []string{
		filepath.Join("utils", "retry", "retry.go"),
		filepath.Join("utils", "http_util", "client.go"),
		filepath.Join("utils", "file_util", "fileutil.go"),
	}
	for _, relPath := range requiredUtils {
		if _, err := os.Stat(filepath.Join(outputDir, relPath)); err != nil {
			t.Fatalf("generated %s missing: %v", relPath, err)
		}
	}
	requiredMCPFiles := []string{
		filepath.Join("internal", "mcp", "server.go"),
		filepath.Join("internal", "mcp", "handler", "hello.go"),
		filepath.Join("internal", "mcp", "service", "greeting.go"),
		filepath.Join("docs", "examples", "mcp.json"),
	}
	for _, relPath := range requiredMCPFiles {
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
	if !strings.Contains(helpOutput, "Start MCP (Model Context Protocol) server.") {
		t.Fatalf("mcp --help output missing expected text:\n%s", helpOutput)
	}
	if !strings.Contains(helpOutput, "all mode uses --port for SSE, and --port+1 for HTTP") {
		t.Fatalf("mcp --help output missing all-mode port mapping:\n%s", helpOutput)
	}
}

func TestGenerateMCPConfigAllRequiresEnabledTransport(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "demo-config-all-validation")
	data := TemplateData{
		ModuleName:  "github.com/test/demo-config-all-validation",
		BinaryName:  "demo-config-all-validation",
		ProjectName: "demo-config-all-validation",
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

	invalidConfig := `{
  "name": "demo-config-all-validation",
  "version": "v0.1.0",
  "transport": "all",
  "host": "0.0.0.0",
  "port": 8080,
  "enable_stdio": false,
  "enable_sse": false,
  "enable_http": false
}`
	configPath := filepath.Join(outputDir, "invalid_all_config.json")
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0o644); err != nil {
		t.Fatalf("write invalid config file: %v", err)
	}

	stdout, stderr, err := runGoCommand(outputDir, "run", ".", "mcp", "--config", configPath)
	if err == nil {
		t.Fatalf(
			"expected config validation failure, got success\nstdout:\n%s\nstderr:\n%s",
			stdout,
			stderr,
		)
	}
	output := stdout + stderr
	if !strings.Contains(output, "transport=all requires at least one enabled transport") {
		t.Fatalf("unexpected error output:\n%s", output)
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
	if _, err := os.Stat(filepath.Join(outputDir, "README_CN.md")); err != nil {
		t.Fatalf("expected README_CN.md to exist in minimal mode: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "cmd", "greet.go")); !os.IsNotExist(err) {
		t.Fatalf("expected cmd/greet.go to be absent, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "cmd", "mcp.go")); !os.IsNotExist(err) {
		t.Fatalf("expected cmd/mcp.go to be absent, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "internal")); !os.IsNotExist(err) {
		t.Fatalf("expected internal directory to be absent, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "docs")); !os.IsNotExist(err) {
		t.Fatalf("expected docs directory to be absent, got err=%v", err)
	}
	for _, relPath := range []string{
		filepath.Join("internal", "mcp", "server.go"),
		filepath.Join("internal", "mcp", "handler", "hello.go"),
		filepath.Join("internal", "mcp", "service", "greeting.go"),
	} {
		if _, err := os.Stat(filepath.Join(outputDir, relPath)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be absent, got err=%v", relPath, err)
		}
	}
	for _, relPath := range []string{
		filepath.Join("utils", "retry", "retry.go"),
		filepath.Join("utils", "http_util", "client.go"),
		filepath.Join("utils", "file_util", "fileutil.go"),
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

	if err := cmd.Start(); err != nil {
		t.Fatalf("start mcp command: %v", err)
	}
	protocolReader := bufio.NewReader(stdoutPipe)

	readFrameWithTimeout := func(timeout time.Duration) (string, error) {
		t.Helper()

		responseCh := make(chan string, 1)
		responseErrCh := make(chan error, 1)
		go func() {
			body, readErr := readJSONRPCFrame(protocolReader)
			if readErr != nil {
				responseErrCh <- readErr
				return
			}
			responseCh <- body
		}()

		select {
		case body := <-responseCh:
			return body, nil
		case err := <-responseErrCh:
			return "", err
		case <-time.After(timeout):
			return "", fmt.Errorf("timed out waiting for json-rpc response")
		}
	}

	sendRequest := func(request map[string]any) {
		t.Helper()

		requestBody, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("marshal json-rpc request: %v", err)
		}
		if _, err := fmt.Fprintf(stdin, "%s\n", requestBody); err != nil {
			t.Fatalf("write json-rpc request: %v", err)
		}
	}

	readResponseForID := func(expectedID int) (map[string]any, string) {
		t.Helper()

		deadline := time.Now().Add(10 * time.Second)
		for {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				t.Fatalf("timed out waiting for json-rpc response id=%d", expectedID)
			}

			body, err := readFrameWithTimeout(remaining)
			if err != nil {
				t.Fatalf("read json-rpc response for id=%d: %v", expectedID, err)
			}

			var response map[string]any
			if err := json.Unmarshal([]byte(body), &response); err != nil {
				t.Fatalf("json-rpc response is not valid JSON: %v\nbody: %s", err, body)
			}
			if response["jsonrpc"] != "2.0" {
				t.Fatalf("json-rpc version mismatch: %#v", response["jsonrpc"])
			}

			idRaw, hasID := response["id"]
			if !hasID || idRaw == nil {
				// Ignore notifications and continue waiting for target response.
				continue
			}
			idNumber, ok := idRaw.(float64)
			if !ok {
				t.Fatalf("json-rpc id has unexpected type: %#v", idRaw)
			}
			if int(idNumber) != expectedID {
				// Ignore out-of-order responses and continue waiting for target response.
				continue
			}

			if responseErr, exists := response["error"]; exists && responseErr != nil {
				t.Fatalf("json-rpc returned error for id=%d: %#v", expectedID, responseErr)
			}

			return response, body
		}
	}

	// Step 1: initialize session.
	sendRequest(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "scaf_fold-smoke-test",
				"version": "1.0.0",
			},
		},
	})
	initializeResp, initializeBody := readResponseForID(1)
	if _, ok := initializeResp["result"].(map[string]any); !ok {
		t.Fatalf("initialize response missing result payload: %#v", initializeResp)
	}
	sendRequest(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
		"params":  map[string]any{},
	})

	// Step 2: list tools and assert hello exists.
	sendRequest(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	listResp, listBody := readResponseForID(2)
	listResult, ok := listResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/list response missing result payload: %#v", listResp)
	}
	tools, ok := listResult["tools"].([]any)
	if !ok {
		t.Fatalf("tools/list response missing tools array: %#v", listResp)
	}
	hasHelloTool := false
	for _, item := range tools {
		tool, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := tool["name"].(string); ok && name == "hello" {
			hasHelloTool = true
			break
		}
	}
	if !hasHelloTool {
		t.Fatalf("tools/list response does not include hello tool: %#v", listResp)
	}

	// Step 3: call hello tool and assert greeting payload.
	sendRequest(map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "hello",
			"arguments": map[string]any{
				"name": "Tester",
			},
		},
	})
	callResp, callBody := readResponseForID(3)
	callResult, ok := callResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/call response missing result payload: %#v", callResp)
	}
	content, ok := callResult["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("tools/call response missing content payload: %#v", callResp)
	}
	firstContent, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("tools/call first content item has invalid shape: %#v", content[0])
	}
	text, ok := firstContent["text"].(string)
	if !ok {
		t.Fatalf("tools/call first content item missing text: %#v", firstContent)
	}
	if text != "Hello, Tester!" {
		t.Fatalf("unexpected hello tool result text: %q", text)
	}

	_ = stdin.Close()

	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	<-stderrDone

	if ctx.Err() != nil {
		t.Fatalf("mcp command context error: %v", ctx.Err())
	}

	stderrOutput := stderrBuf.String()
	if !strings.Contains(stderrOutput, "starting mcp stdio server") {
		t.Fatalf("stderr missing mcp startup log:\n%s", stderrOutput)
	}
	for _, body := range []string{initializeBody, listBody, callBody} {
		if strings.Contains(body, "starting mcp stdio server") {
			t.Fatalf("stdout protocol payload unexpectedly contains log text: %s", body)
		}
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

func readJSONRPCFrame(reader *bufio.Reader) (string, error) {
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

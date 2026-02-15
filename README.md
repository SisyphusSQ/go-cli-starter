# go-cli-starter

English | [简体中文](./README_CN.md)

[![Go Version](https://img.shields.io/github/go-mod/go-version/SisyphusSQ/go-cli-starter?style=flat-square)](https://github.com/SisyphusSQ/go-cli-starter/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/SisyphusSQ/go-cli-starter?style=flat-square)](https://github.com/SisyphusSQ/go-cli-starter/releases)
[![License](https://img.shields.io/github/license/SisyphusSQ/go-cli-starter?style=flat-square)](./LICENSE)
[![Issues](https://img.shields.io/github/issues/SisyphusSQ/go-cli-starter?style=flat-square)](https://github.com/SisyphusSQ/go-cli-starter/issues)

`go-cli-starter` is a Go CLI scaffolding tool that generates a production-ready command-line project template in seconds.

## Why use it

When starting a new CLI project, you usually repeat the same setup work:

- Build a Cobra command skeleton
- Set up logging and version metadata
- Add Makefile-based engineering workflow
- Create a minimal MCP command example

`go-cli-starter` templates all of this so you can focus on business logic from day one.

## What the template includes

| Capability | Description |
|---|---|
| CLI framework | Based on `github.com/spf13/cobra`, with `root/version/greet/mcp` commands |
| Logging | Zap logger with `stdio/stderr/file` modes (`stdio` by default) |
| MCP example | Built on `github.com/mark3labs/mcp-go`, with a built-in `hello` tool |
| Engineering defaults | Enhanced `Makefile` (`fmt/lint/test/coverage/tidy/cross-build`) |
| Version injection | `vars` package + `version` command with ldflags support |
| Changelog | Built-in `changeLog.md` sample (`feature/optimization/bugFix`) |
| License | Built-in MIT `LICENSE` template for generated projects |

## Installation

### 1) Recommended: install latest

```bash
go install github.com/SisyphusSQ/go-cli-starter@latest
```

Then use it directly:

```bash
go-cli-starter new ./my-new-cli -m github.com/you/my-new-cli -b my-new-cli
# or generate minimal template (root/version only)
go-cli-starter new ./my-min-cli --minimal -m github.com/you/my-min-cli -b my-min-cli
```

### 2) Run without installing

```bash
go run github.com/SisyphusSQ/go-cli-starter@latest new ./my-new-cli -m github.com/you/my-new-cli -b my-new-cli
```

### 3) Build from source

```bash
git clone https://github.com/SisyphusSQ/go-cli-starter.git
cd go-cli-starter
make tidy
make build
./bin/go-cli-starter new ./my-new-cli -m github.com/you/my-new-cli -b my-new-cli
```

## Create a new project

```bash
go-cli-starter new <output-dir> [flags]
```

| Parameter | Meaning | Default |
|---|---|---|
| `<output-dir>` | Output directory (required) | None |
| `-m, --module` | Go module path for the generated project | `example.com/<directory-name>` |
| `-b, --binary` | Binary name for the generated project | Directory name |
| `--minimal` | Generate minimal template (only `root/version`) | `false` |

Example:

```bash
go-cli-starter new ./my-new-cli -m github.com/you/my-new-cli -b my-new-cli
cd ./my-new-cli
go mod tidy
make build
./bin/my-new-cli version
```

## Initialize in current directory

Use this command in an **existing empty directory**:

```bash
go-cli-starter init [flags]
```

Parameters:

- `-m, --module`: Go module path (default `example.com/<current-directory-name>`)
- `-b, --binary`: Binary name (default current directory name)
- `--minimal`: initialize minimal template with only `root/version`

Example:

```bash
mkdir my-init-cli && cd my-init-cli
go-cli-starter init -m github.com/you/my-init-cli -b my-init-cli
```

## Generated project structure (example)

```text
my-new-cli/
├── main.go
├── go.mod
├── Makefile
├── changeLog.md
├── LICENSE
├── README.md
├── cmd/
│   ├── root.go
│   ├── version.go
│   ├── greet.go
│   └── mcp.go
├── internal/
│   └── mcp/
│       ├── server.go
│       ├── handler/
│       │   └── hello.go
│       └── service/
│           └── greeting.go
├── pkg/log/logger.go
├── utils/timeutil/timeutil.go
└── vars/vars.go
```

MCP layering in generated template:

- `cmd/mcp.go`: thin CLI entry, delegates to internal package
- `internal/mcp/server.go`: builds and starts MCP stdio server
- `internal/mcp/handler/*`: tool registration + request parsing
- `internal/mcp/service/*`: business logic called by handlers

## Default commands in generated project

| Command | Purpose |
|---|---|
| `version` | Print version and build metadata |
| `greet --name <name>` | Sample command demonstrating Cobra development |
| `mcp` | Start MCP stdio server (with built-in `hello` tool) |

In `--minimal` mode, generated project keeps only the `version` command.
`cmd/greet.go`, `cmd/mcp.go`, and `internal/mcp/**` are not generated.

## Logging modes

The generated project supports:

```bash
--log-output stdio|stderr|file
--log-file ./logs/app.log
```

| Mode | Behavior |
|---|---|
| `stdio` | Logs to stdout (default) |
| `stderr` | Logs to stderr |
| `file` | Logs to file (with rotation) |

Note: in `mcp` command, logging automatically switches to `stderr` to avoid polluting JSON-RPC frames on stdout.

## FAQ

### 1. Why does `mcp` avoid logging to stdout by default?

Because MCP stdio uses stdout for JSON-RPC messages. If logs are mixed into stdout, clients may fail to parse protocol frames.

### 2. Why does `--module` default to `example.com/<directory-name>`?

The default favors quick bootstrap while staying Go-module valid. For team or open-source projects, always pass `-m` explicitly, for example:

```bash
-m github.com/you/my-new-cli
```

### 3. Can I generate into an existing directory?

Only if the target directory is empty. Non-empty directories are rejected to prevent accidental overwrite.

### 4. What's the difference between `new` and `init`?

- `new`: generate into a target path (`new ./my-cli`)
- `init`: initialize in current empty directory (`init`)

### 5. What should I run first after generation?

```bash
go mod tidy
make build
```

Then verify with `version`, `greet`, and `mcp --help`.
For minimal template, verify with `version` only.

### 6. What is the minimum Go version?

The repository currently targets Go 1.26.0 (including generated templates).

## Roadmap

- [x] Support `new` command to generate a complete CLI template
- [x] Bundle Cobra + Zap + Makefile + changeLog in template
- [x] Bundle MCP command and sample tool in template
- [x] Add `new --minimal` (root/version only)
- [x] Add `init` command for bootstrapping existing empty directories
- [ ] Add `new --template <name>` for multiple presets (api/tool/agent)
- [ ] Add optional GitHub Actions CI template
- [ ] Add test template generator (`testify` / table-driven)
- [ ] Provide Homebrew distribution

## Contributing

Issues and PRs are welcome:

- Bug report: <https://github.com/SisyphusSQ/go-cli-starter/issues>
- Feature request: please include use case, expected behavior, and compatibility notes

## License

MIT. See [LICENSE](./LICENSE).

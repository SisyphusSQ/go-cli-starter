# go-cli-starter

[English](./README.md) | 简体中文

[![Go Version](https://img.shields.io/github/go-mod/go-version/SisyphusSQ/go-cli-starter?style=flat-square)](https://github.com/SisyphusSQ/go-cli-starter/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/SisyphusSQ/go-cli-starter?style=flat-square)](https://github.com/SisyphusSQ/go-cli-starter/releases)
[![License](https://img.shields.io/github/license/SisyphusSQ/go-cli-starter?style=flat-square)](./LICENSE)
[![Issues](https://img.shields.io/github/issues/SisyphusSQ/go-cli-starter?style=flat-square)](https://github.com/SisyphusSQ/go-cli-starter/issues)

`go-cli-starter` 是一个 Go CLI 脚手架工具，用来快速生成「可直接开发、可发布」的命令行项目模板。

## 为什么要用它

每次新建 CLI 项目，常见重复工作包括：

- 搭建 Cobra 命令骨架
- 接入日志和版本信息
- 写 Makefile / 基础工程化
- 给 MCP 命令做一份最小可运行示例

`go-cli-starter` 把这部分一次性模板化，生成后即可专注在业务命令本身。

## 模板能力一览

| 能力 | 说明 |
|---|---|
| CLI 框架 | 基于 `github.com/spf13/cobra`，内置 `root/version/greet/mcp` |
| 日志 | Zap 日志，支持 `stdio/stderr/file`，默认 `stdio` |
| MCP 示例 | 基于 `github.com/mark3labs/mcp-go`，内置 `hello` tool |
| 工程化 | 内置增强 `Makefile`（fmt/lint/test/coverage/tidy/交叉编译） |
| 版本注入 | 内置 `vars` + `version` 命令（支持 ldflags 注入） |
| 变更日志 | 内置 `changeLog.md` 样例（feature/optimization/bugFix） |
| License | 内置 MIT `LICENSE` 模板，生成项目默认携带 |

## 安装方式

### 1) 推荐：直接安装最新版本

```bash
go install github.com/SisyphusSQ/go-cli-starter@latest
```

安装后可直接使用：

```bash
go-cli-starter new ./my-new-cli -m github.com/you/my-new-cli -b my-new-cli
# 或生成极简模板（仅 root/version）
go-cli-starter new ./my-min-cli --minimal -m github.com/you/my-min-cli -b my-min-cli
```

### 2) 不安装，直接运行

```bash
go run github.com/SisyphusSQ/go-cli-starter@latest new ./my-new-cli -m github.com/you/my-new-cli -b my-new-cli
```

### 3) 从源码构建

```bash
git clone https://github.com/SisyphusSQ/go-cli-starter.git
cd go-cli-starter
make tidy
make build
./bin/go-cli-starter new ./my-new-cli -m github.com/you/my-new-cli -b my-new-cli
```

## 如何 new 一个项目

```bash
go-cli-starter new <output-dir> [flags]
```

| 参数 | 含义 | 默认值 |
|---|---|---|
| `<output-dir>` | 输出目录（必须） | 无 |
| `-m, --module` | 生成项目的 Go Module | `example.com/<目录名>` |
| `-b, --binary` | 生成二进制名称 | 目录名 |
| `--minimal` | 生成极简模板（仅 `root/version`） | `false` |

示例：

```bash
go-cli-starter new ./my-new-cli -m github.com/you/my-new-cli -b my-new-cli
cd ./my-new-cli
go mod tidy
make build
./bin/my-new-cli version
```

## 在当前目录初始化项目

在**已存在的空目录**中使用：

```bash
go-cli-starter init [flags]
```

参数：

- `-m, --module`：Go Module 路径（默认 `example.com/<当前目录名>`）
- `-b, --binary`：二进制名称（默认当前目录名）
- `--minimal`：初始化极简模板（仅 `root/version`）

示例：

```bash
mkdir my-init-cli && cd my-init-cli
go-cli-starter init -m github.com/you/my-init-cli -b my-init-cli
```

## 生成后的项目结构（示例）

```text
my-new-cli/
├── main.go
├── go.mod
├── Makefile
├── changeLog.md
├── LICENSE
├── README.md
├── README_CN.md
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
├── utils/time_util/timeutil.go
└── vars/vars.go
```

模板中的 MCP 分层：

- `cmd/mcp.go`：薄 CLI 入口，仅做命令委托
- `internal/mcp/server.go`：创建并启动 MCP stdio server
- `internal/mcp/handler/*`：tool 注册与请求解析
- `internal/mcp/service/*`：handler 调用的业务逻辑

## 生成项目默认命令

| 命令 | 用途 |
|---|---|
| `version` | 打印版本/构建信息 |
| `greet --name <name>` | 样例命令，演示 Cobra 开发模式 |
| `mcp` | 启动 MCP stdio Server（内置 `hello` tool） |

若使用 `--minimal`，生成项目仅保留 `version` 命令。
`cmd/greet.go`、`cmd/mcp.go` 与 `internal/mcp/**` 不会生成。

## 日志模式说明

生成项目支持：

```bash
--log-output stdio|stderr|file
--log-file ./logs/app.log
```

| 模式 | 行为 |
|---|---|
| `stdio` | 日志输出到 stdout（默认） |
| `stderr` | 日志输出到 stderr |
| `file` | 日志写入文件（支持滚动） |

说明：`mcp` 命令会自动切换到 `stderr`，避免 JSON-RPC 协议流被日志污染。

## FAQ

### 1. 为什么 `mcp` 命令默认不把日志打到 stdout？

因为 MCP stdio 协议使用 stdout 传输 JSON-RPC 消息。日志如果混入 stdout，会导致客户端解析失败。  
所以模板默认在 `mcp` 场景将日志重定向到 stderr。

### 2. 为什么 `--module` 默认是 `example.com/<目录名>`？

脚手架默认值既保证快速初始化，也保证 module 路径语法合法。建议在团队/开源场景始终显式传 `-m`，例如：

```bash
-m github.com/you/my-new-cli
```

### 3. 已存在目录可以直接生成吗？

不行。若目标目录存在且非空，会拒绝覆盖，避免误伤现有文件。

### 4. `new` 和 `init` 有什么区别？

- `new`：在指定路径生成（如 `new ./my-cli`）
- `init`：在当前空目录初始化（如 `init`）

### 5. 生成后第一步该做什么？

进入项目后先执行：

```bash
go mod tidy
make build
```

然后跑 `version`、`greet`、`mcp --help` 进行基础验收。
若是 `--minimal` 模板，只需验证 `version`。

### 6. 最低 Go 版本要求是什么？

当前仓库按 `go.mod` 要求使用 Go 1.26.0（含模板项目）。

## Roadmap

- [x] 支持 `new` 命令生成完整 CLI 模板
- [x] 模板内置 Cobra + Zap + Makefile + changeLog
- [x] 模板内置 MCP 命令与示例 tool
- [x] 增加 `new --minimal`（仅保留 root/version）
- [x] 增加 `init` 命令（在现有空目录初始化）
- [ ] 增加 `new --template <name>` 多模板选择（api/tool/agent）
- [ ] 增加 GitHub Actions CI 模板（可选开关）
- [ ] 增加单元测试模板生成器（`testify`/表驱动）
- [ ] 发布 Homebrew 安装方式

## 贡献

欢迎提交 Issue / PR：

- Bug 反馈：<https://github.com/SisyphusSQ/go-cli-starter/issues>
- 功能建议：请描述使用场景、预期行为、兼容性考虑

## License

MIT，见 [LICENSE](./LICENSE)。

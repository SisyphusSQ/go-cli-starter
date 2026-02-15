### v1.1.0(20260215)
#### feature:
1. 新增 MCP 模板分层结构：将 `cmd/mcp.go` 中的 server 与 tool 逻辑下沉到 `internal/mcp/server.go`、`internal/mcp/handler/hello.go`、`internal/mcp/service/greeting.go`，提升 CLI 层职责清晰度并便于后续扩展。
2. 新增模板测试覆盖：`internal/scaffold/scaffold_test.go` 增加对 `internal/mcp` 分层文件存在性与 minimal 模式缺失性的断言，确保模板输出与分层约定一致。

#### optimization:
1. 优化模板生成跳过逻辑：`Generate` 在 minimal 模式下遇到被跳过目录时返回 `fs.SkipDir`，避免创建无效空目录。
2. 优化中英文文档与模板 README 的分层说明，统一展示 `cmd -> internal/mcp/server -> handler -> service` 的 MCP 代码组织方式，便于使用者理解模板架构。
3. 升级脚手架版本号：`AppVersion` 更新为 `v1.1.0`，与本次发布内容保持一致。

## v1.0.0(20260215)
#### feature:
1. 新增 `go-cli-starter` 脚手架工具，支持通过 `go run github.com/SisyphusSQ/go-cli-starter@latest new <dir>` 快速生成 Go CLI 项目基础结构。
2. 新增模板命令体系：`version`、`greet`、`mcp`，并支持 `new --minimal` 生成极简模板（仅保留 `root/version`，自动跳过 `greet/mcp` 与 `mcp-go` 依赖）。
3. 新增 `init` 命令，支持在当前目录初始化项目并可与 `--module`、`--binary`、`--minimal` 组合使用；当前目录仅含隐藏文件（如 `.git`）时也可正常初始化。
4. 新增模板通用工具模块：`utils/retry`（指数退避+抖动重试）、`utils/httputil`（内置重试与 JSON 辅助的 HTTP 客户端）、`utils/fileutil`（文件存在判断/复制/JSON 读写），并在 `minimal` 模式下同样保留。
5. 新增模板日志组件，支持 `stdio/stderr/file` 三种输出模式；通过注解机制统一控制 MCP 场景日志强制输出到 `stderr`，避免污染 stdout 协议流。
6. 新增模板工程基础文件：`LICENSE`、`README.md`、`.gitignore`、`.golangci.yml`、`changeLog.md` 与 `Makefile`，完善开箱即用体验。
7. 新增模板与根项目构建增强：`Makefile` 默认 `test -race`，支持注入 `AppVersion/GoVersion/BuildTime/GitCommit/GitRemote` 等构建信息。
8. 新增输入与模板渲染约束：加强 `module/binary/projectName` 校验，避免生成不可编译项目；模板 `go.mod` 支持按运行时自动匹配 Go 版本。
9. 新增测试与质量保障：补充 `cmd/new` 与 `init` 集成测试、`internal/scaffold` 校验用例，以及模板生成后的 `go mod tidy`、`go build`、`mcp` 探活 E2E 测试。
10. 新增中英文文档能力：完善 `README.md`/`README_CN.md`，补充 `new --minimal`、`init`、FAQ 与模板内置 utilities 说明。

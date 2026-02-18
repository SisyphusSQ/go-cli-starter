### v1.2.0(20260218)
#### feature:
1. 新增 MCP 多协议传输模板封装：`internal/mcp/server.go` 模板支持 `stdio`、`sse`、`http` 与 `all` 并发启动，补充 `Transport`/`Config` 结构、`LoadConfig` 配置加载与信号优雅退出能力，便于生成项目在不同 MCP 客户端与部署形态下直接使用。
2. 新增 MCP 命令多协议参数：`cmd/mcp.go` 模板新增 `--transport`、`--host`、`--port`、`--config`，并支持按 JSON 配置文件启动，降低生成项目接入 SSE/HTTP 场景的改造成本。
3. 新增 MCP 配置示例模板：增加 `docs/examples/mcp.json`，生成项目可直接参考并复用标准配置。

#### optimization:
1. 升级模板 `mcp-go` 依赖至 `v0.44.0`，与多协议传输能力保持一致。
2. 优化 minimal 模式模板跳过逻辑：新增 `docs/` 目录跳过规则，避免生成无效文档目录。
3. 优化模板文档与测试：更新中英文 README 的 MCP 使用说明与多协议示例，并补充 `docs/examples/mcp.json` 存在性和 minimal 模式目录缺失断言。
4. 根据代码评审修复多协议边界：`transport=all` 在配置场景增加“至少启用一个协议”校验，避免无服务阻塞；同步统一 all 模式端口语义（SSE=`--port`、HTTP=`--port+1`），并补充对应回归测试与 JSON-RPC 读帧稳定性优化（复用同一 `bufio.Reader`）。

### v1.1.1(20260215)
#### feature:
1. 统一包名与目录蛇形命名
2. 新增模板中文文档：在模板中补充 `README_CN.md.tmpl`，生成项目默认同时包含 `README.md` 与 `README_CN.md`。

#### optimization:
1. 统一模板 utilities 的目录与包名为蛇形。
2. 更新根项目中英文 README 的目录结构示例，使文档与最新模板输出保持一致。
3. 升级脚手架版本号：`AppVersion` 更新为 `v1.1.1`，与本次发布内容保持一致。

### v1.1.0(20260215)
#### feature:
1. 新增 MCP 模板分层结构：将 `cmd/mcp.go` 中的 server 与 tool 逻辑下沉到 `internal/mcp/server.go`、`internal/mcp/handler/hello.go`、`internal/mcp/service/greeting.go`，提升 CLI 层职责清晰度并便于后续扩展。
2. 新增模板测试覆盖：`internal/scaf_fold/scaffold_test.go` 增加对 `internal/mcp` 分层文件存在性与 minimal 模式缺失性的断言，确保模板输出与分层约定一致。

#### optimization:
1. 优化模板生成跳过逻辑：`Generate` 在 minimal 模式下遇到被跳过目录时返回 `fs.SkipDir`，避免创建无效空目录。
2. 优化中英文文档与模板 README 的分层说明，统一展示 `cmd -> internal/mcp/server -> handler -> service` 的 MCP 代码组织方式，便于使用者理解模板架构。
3. 升级脚手架版本号：`AppVersion` 更新为 `v1.1.0`，与本次发布内容保持一致。

## v1.0.0(20260215)
#### feature:
1. 新增 `go-cli-starter` 脚手架工具，支持通过 `go run github.com/SisyphusSQ/go-cli-starter@latest new <dir>` 快速生成 Go CLI 项目基础结构。
2. 新增模板命令体系：`version`、`greet`、`mcp`，并支持 `new --minimal` 生成极简模板（仅保留 `root/version`，自动跳过 `greet/mcp` 与 `mcp-go` 依赖）。
3. 新增 `init` 命令，支持在当前目录初始化项目并可与 `--module`、`--binary`、`--minimal` 组合使用；当前目录仅含隐藏文件（如 `.git`）时也可正常初始化。
4. 新增模板通用工具模块：`utils/retry`（指数退避+抖动重试）、`utils/http_util`（内置重试与 JSON 辅助的 HTTP 客户端）、`utils/file_util`（文件存在判断/复制/JSON 读写），并在 `minimal` 模式下同样保留。
5. 新增模板日志组件，支持 `stdio/stderr/file` 三种输出模式；通过注解机制统一控制 MCP 场景日志强制输出到 `stderr`，避免污染 stdout 协议流。
6. 新增模板工程基础文件：`LICENSE`、`README.md`、`.gitignore`、`.golangci.yml`、`changeLog.md` 与 `Makefile`，完善开箱即用体验。
7. 新增模板与根项目构建增强：`Makefile` 默认 `test -race`，支持注入 `AppVersion/GoVersion/BuildTime/GitCommit/GitRemote` 等构建信息。
8. 新增输入与模板渲染约束：加强 `module/binary/projectName` 校验，避免生成不可编译项目；模板 `go.mod` 支持按运行时自动匹配 Go 版本。
9. 新增测试与质量保障：补充 `cmd/new` 与 `init` 集成测试、`internal/scaf_fold` 校验用例，以及模板生成后的 `go mod tidy`、`go build`、`mcp` 探活 E2E 测试。
10. 新增中英文文档能力：完善 `README.md`/`README_CN.md`，补充 `new --minimal`、`init`、FAQ 与模板内置 utilities 说明。

---
date: 2026-02-18
type: feature
module: mcp
tags: [mcp, scaffold, multi-transport, review-fix]
---

# MCP 多协议封装反哺脚手架

## 原始需求
用户希望参考 `summary-sys` 的 MCP 封装能力，将 `io(stdio)`、`sse`、`http`
多协议支持反哺到 `go-cli-starter` 模板中，并保持模板可用、可测、可维护。

后续又要求：
- 使用 `code-reviewer-codex` 做新增代码评审
- 基于评审结果补齐修复方案并落地

## 决策记录
- 方案 A：只补 `ServeSSE/ServeHTTP`，保持 CLI 单一入口不变
  - 优点：改动小
  - 缺点：配置驱动、all 模式和文档语义不完整
- 方案 B：对齐 `summary-sys` 的完整模式（transport flags + config + all）
  - 优点：能力完整、迁移成本低、模板可复用
  - 选择：B

- 评审后关键取舍：
  - 对 `transport=all` 增加强校验（至少启用一个协议），优先消除“无服务阻塞”
  - 明确 all 模式端口语义：`SSE=port`、`HTTP=port+1`
  - 补最小回归测试而非过度扩展测试矩阵，先覆盖高风险路径

## 执行摘要
- 改动文件：
  - `internal/scaf_fold/_template/internal/mcp/server.go.tmpl`
  - `internal/scaf_fold/_template/cmd/mcp.go.tmpl`
  - `internal/scaf_fold/_template/go.mod.tmpl`
  - `internal/scaf_fold/_template/docs/examples/mcp.json.tmpl`
  - `internal/scaf_fold/scaffold.go`
  - `internal/scaf_fold/_template/README.md.tmpl`
  - `internal/scaf_fold/_template/README_CN.md.tmpl`
  - `internal/scaf_fold/scaffold_test.go`
  - `changeLog.md`

- 核心变更：
  - 模板支持 `stdio/sse/http/all`，并支持 `--config` 启动
  - 增加配置结构与加载逻辑，升级 `mcp-go` 到 `v0.44.0`
  - minimal 模式跳过 `docs/` 目录
  - 新增 `docs/examples/mcp.json` 模板
  - 修复评审发现的高风险问题：`all + 全 false` 直接报错
  - 统一 help/readme/log 的 all 模式端口语义
  - 补充回归测试并修复 JSON-RPC 读帧潜在 flaky（复用 `bufio.Reader`）

## 经验教训
### 有效做法
- 先做“功能反哺”，再用代码评审闭环补边界条件，质量提升明显。
- 模板类改动必须同步文档、示例配置、测试三件套，否则容易语义漂移。
- 对协议型 CLI（尤其 stdout/stderr 有约束）应优先做 smoke + 错误路径测试。

### 踩坑记录
- `transport=all` 在配置字段缺失时默认 bool=false，容易触发“空启动阻塞”。
  - 解决：在 CLI 配置分支和 server 入口双层校验。
- all 模式下多端口语义如果不写清楚，用户会误认为单端口统一生效。
  - 解决：help、README、日志统一写明映射关系。
- JSON-RPC 逐帧读取若每次新建 reader，可能丢失预读缓冲导致偶发超时。
  - 解决：复用同一个 `bufio.Reader`。

## 相关
- 关联记忆：暂无
- 参考资料：
  - `summary-sys/internal/mcp/server.go`
  - `summary-sys/cmd/mcp.go`

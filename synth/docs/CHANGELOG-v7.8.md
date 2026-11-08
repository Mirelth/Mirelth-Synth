# Synth v7.8 发布说明（当前版本 v7.8.0）

> 发布日期：2026-10-29
> 版本：v7.8.0
> 代号：累计发布（v7.6 第一印象 / v7.7 A2A 楔子 / v7.8 8.0 彩排）

---

## 说明：为什么是 7.8.0

v7.6、v7.7、v7.8 三批工作均已完成并合入 `main`，但提交在时间线上交错（v7.6 的收尾修复晚于 v7.7/v7.8 的功能提交），无法回溯成三个内容干净的独立 semver tag。因此本项目以 **v7.8.0 做一次累计发布**：

- v7.6 / v7.7 不单独发布，保留 `batch/v7.6`、`batch/v7.7` 两个非发布 tag 作为历史定位；
- v7.8.0 包含三批全部内容与 2026-10-26 安全加固批次；
- 详见 [`docs/版本状态-v7.6-v7.8.md`](../../docs/版本状态-v7.6-v7.8.md)。

## 核心亮点

- **英文前门**：README / getting-started 英译，clean-room quickstart CI 门，`synth init` 免 API key 首跑。
- **API 表面冻结**：`pkg/api_surface_test.go` + 1367 个导出符号黄金快照，新增导出即红。
- **A2A 跨进程互操作 demo**：发现→委托→轮询→取回→取消，零凭据约 5 秒，协议符合性 8/8。
- **WASM 工具供应链 demo**：ed25519 签名、公钥钉扎、验签门、wazero 沙箱真实执行，四类拒绝分支全部实证。
- **8.0 彩排**：387 个导出符号 `v8.0-cut` 标记、`synth migrate`（/v7→/v8 重写）、8.0 范围决定稿草案（待签署）。
- **安全加固**：Shell workdir/选项值/符号链接逃逸拦截，MCP/A2A 默认 loopback，流式背压与并发缺陷修复。

## 行为变更（升级需注意）

1. `builtin.NewShell()` 默认把路径禁锢到当前工作目录；需要更多目录请显式 `WithAllowedWorkdirs`，完全关闭用 `WithAllowedWorkdirs(nil)`（不推荐）。
2. A2A JSON-RPC/gRPC Server 默认认证改为 loopback-only；跨机访问必须 `WithAuth` / `WithGRPCAuth`。同机反向代理会显示为 loopback，代理部署必须显式认证。
3. MCP HTTP Server 未配置 APIKey 时拒绝非 loopback 监听/请求；跨机部署必须 `WithAPIKey`。
4. TS SDK 升级 zod 4 / vitest 5；覆盖率阈值按新运行环境重校准。

完整技术记录见 [`docs/CHANGELOG.md`](../../docs/CHANGELOG.md) 的 `[7.8.0]` 节。

## 安装与升级

```bash
go install github.com/Mirelth/Mirelth-Synth/v7/synth/cmd/synth@v7.8.0
synth version   # 应显示 7.8.0
```

## 验证

- Go：主模块 104 个测试包 + gateway/pgvector/wasm/operator 全绿；核心包 `-race` 通过；Windows/darwin 交叉编译通过
- 门禁：版本一致性、deprecation 243/243、文档死链、双线矩阵、覆盖率（agent 85.7% / llm 75.9% / pkg 72.5%）
- 安全：`govulncheck` 0 可达漏洞
- TS：typecheck / 2749 tests / eslint 全绿
- CNB：`go-ci`、`ts-ci` 均通过

## 已知边界

- 8.0 范围决定稿尚待维护者逐项签署；签署前不执行任何删除。
- v7.6 的人类退出判据（非维护者 15 分钟首胜）仍待外部记录。
- 流式消费者完全停滞时，error 帧可能无法送达，调用方只能拿到已收内容。

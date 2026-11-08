# Synth v7.9 发布说明（当前版本 v7.9.0）

> 发布日期：2026-11-08
> 版本：v7.9.0
> 代号：改名发布（AgentPrimordia → Synth）
> 迁移指南：[`docs/迁移到-Synth.md`](../../docs/迁移到-Synth.md)

---

## 一句话

这是一次**不兼容的品牌与契约切换**：项目从 `AgentPrimordia`（组织 `Gleamseekers`）改名为 **Synth**（组织 `Mirelth`，仓库 `Mirelth/Mirelth-Synth`），模块路径、Go 公共包名、CLI、配置契约、环境变量、指标前缀、CRD group 与全部分发身份一起更换。框架**不提供兼容别名层**，下游需按迁移指南一次性完成。

选择 7.9.0 而不是 8.0.0：`/v7` 导入版本后缀保留，语义化主版本不变；改名本身是破坏性的，但破坏面止于「路径与命名」，框架能力与行为契约未变。

## 核心变更

- **Go 模块路径**：`github.com/Gleamseekers/AgentPrimordia/v7` → `github.com/Mirelth/Mirelth-Synth/v7`（子模块 `gateway` / `pgvector` / `wasm` / `synth/operator` 同步）。
- **Go 公共 API 包名**：`package ap` → `package synth`，即 `ap.NewAgent(...)` → `synth.NewAgent(...)`。
- **CLI 与配置契约**：`ap` → `synth`（22 个子命令）；`.ap.yaml` / `.ap.json` → `.synth.yaml` / `.synth.json`；`AP_*` → `SYNTH_*`（97 个环境变量）。
- **指标与协议标识**：`ap_*` / `agentprimordia_*` → `synth_*`；CRD group `agent.primordia.dev` → `synth.dev`。
- **分发身份**：npm `@mirelth/synth`、PyPI/crates `mirelth-synth`（Python import 名与 Rust crate 名仍为 `synth`）、VS Code publisher `mirelth`、镜像 `ghcr.io/mirelth/mirelth-synth{,-studio,-operator}`。
- **迁移支持**：`docs/迁移到-Synth.md`（对照表 + 手工要点 + 验证清单 + 回滚）、`scripts/migrate-to-synth.py`（默认 dry-run，只改工作区、不建提交）。

## 新增门禁

- **残留名门禁** `scripts/check-brand-residuals.sh`：扫描旧品牌名、旧组织名、旧 CRD group、`AP_` / `ap_` 前缀、`.ap.yaml` / `.ap.json`、`ap-` 产物名、`ap/` 路径段；命中即失败，接入 CNB 与 GitHub CI。白名单只放四类按设计必须引用旧名的文件：迁移指南、迁移脚本、品牌资产说明、冻结评测题面。
- **protobuf 描述符一致性**：`TestProtoGoPackageMatchesImportPath` 用反射取真实 import 路径反查 `a2a.pb.go` 里的 go_package，防止改路径时「编译过、init panic」。

## 有意保留

- **git 历史与旧 tag**：历史 commit message、16 个旧 tag、`origin` 均保持原样。旧 tag 未推送到新仓库——它们指向的提交里 `go.mod` 仍是旧模块路径，`go get …@v7.8.0` 会报模块路径不匹配。**v7.9.0 是新基线的第一个 tag。**
- **冻结评测题面**：`docs/evals/embedding-corpus-v1.json`、`gap-tools-v1.json`、`synth/docs/evals/v72/p4-intelligence.json` 保持原样（sha256 冻结，改动即门禁报警），内部因此仍含旧品牌字样；这是设计例外，不是残留。
- **AELA 产品截图**：属于另一条产品线，未改名。

## 已知问题（未在本次修复）

- `synth/tests/integration` 的 `TestApStartFlow` 依赖 `pgvector` 子模块发布成 tag，当前 `unknown revision pgvector/v0.0.0`。
- `synth/bench/soak` 的吞吐退化门在共享负载下 flaky（错误率 0、延迟 0%，仅吞吐波动触发）。
- `gt-014`（`gap-tools-v1.json`）的成功断言 `M4QTEA====` 从来不是合法 base32，该题恒不可通过，已在 `docs/evals/README.md` 登记，修正走 `gap-tools-v2.json`。
- Docker 三个镜像未在本仓库 CI 之外真实构建；`@mirelth/synth` 与 `mirelth-synth` 尚未首次发布。

## 安装与升级

```bash
go install github.com/Mirelth/Mirelth-Synth/v7/synth/cmd/synth@v7.9.0
synth version   # 应显示 7.9.0
```

> 模块依赖：`pgvector` 是独立子模块（`github.com/Mirelth/Mirelth-Synth/pgvector`），
> 主模块以 `v0.1.0` 依赖它。该子模块已有 tag `pgvector/v0.1.0`，因此干净环境下
> `go get` / `go install` 无需任何本地 replace；仓库内 `go.work` 的 replace 仅用于
> 本地开发。

下游项目从旧品牌迁移：

```bash
python3 scripts/migrate-to-synth.py --target /path/to/your-agent          # 预览
python3 scripts/migrate-to-synth.py --target /path/to/your-agent --apply  # 执行
```

完整技术记录见 [`docs/CHANGELOG.md`](../../docs/CHANGELOG.md) 的 `[7.9.0]` 节。

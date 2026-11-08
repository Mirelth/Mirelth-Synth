# 迁移到 Synth

> 适用对象：基于 `github.com/Gleamseekers/AgentPrimordia/v7`（v7.8.0 及更早）构建的**下游项目**，以及按旧契约生成的脚手架项目。
> 本文档是「当前事实」类文档，已在 [`文档权威索引.md`](文档权威索引.md) §二登记。

## 1. 为什么要迁移

Synth 改名同时更换了模块路径、Go 公共 API 包名、CLI、配置契约与协议标识。这些属于**不兼容变更**：旧 tag 仍可从模块代理拉取，但不会自动重定向到新路径；继续留在旧 tag 可获得稳定性，但拿不到后续修复。

**框架不提供兼容别名层**（例如不会同时保留 `package ap` 与 `package synth`），以避免双包维护与符号歧义。因此迁移需要一次性完成。

## 2. 变更对照

| 类别 | 旧 | 新 |
|---|---|---|
| Go 模块 | `github.com/Gleamseekers/AgentPrimordia/v7` | `github.com/Mirelth/Mirelth-Synth/v7` |
| Go API 包名 | `ap.NewAgent(...)`、`package ap` | `synth.NewAgent(...)`、`package synth` |
| CLI | `ap init` / `ap run` / `ap a2a` | `synth init` / `synth run` / `synth a2a` |
| 配置文件 | `.ap.yaml` / `.ap.json` | `.synth.yaml` / `.synth.json` |
| 环境变量 | `AP_LLM_API_KEY`、`AP_ROOT`、`AP_*` | `SYNTH_LLM_API_KEY`、`SYNTH_ROOT`、`SYNTH_*` |
| 指标前缀 | `ap_*`、`agentprimordia_*` | `synth_*` |
| npm / 扩展 | `@agentprimordia/sdk`、publisher `agentprimordia`、`agentprimordia.*` 命令 ID | `@mirelth/synth`、publisher `mirelth`、`synth.*` 命令 ID |
| Python / Rust | 包 `agentprimordia` | 分发名 `mirelth-synth`（Python import 名与 Rust crate 名仍为 `synth`） |
| K8s | CRD group `agent.primordia.dev` | `synth.dev` |
| 镜像 | `ghcr.io/gleamseekers/...` | `ghcr.io/mirelth/...` |

## 3. 自动迁移（推荐）

```bash
# 1) 预览（默认 dry-run，不写任何文件）
python3 scripts/migrate-to-synth.py --target /path/to/your-agent

# 2) 确认预览无误后执行
python3 scripts/migrate-to-synth.py --target /path/to/your-agent --apply
```

脚本会：

- 替换源码与配置中的品牌名、模块路径、包名别名、环境变量前缀、指标前缀、CRD group；
- 重命名 `.ap.yaml` / `.ap.json` 等含旧名的文件；
- 跳过 `.git`、`node_modules`、`dist`、`vendor`、二进制文件；
- 保持原文件编码（UTF-8，失败时回退 GBK）；
- 检测目标路径冲突，冲突即中止，不覆盖已有文件。

脚本的替换**只落在无歧义的语境**上：完整模块路径、`AgentPrimordia` 各大小写变体、`Gleamseekers`、`AP_` / `ap_` 前缀、`.ap.yaml` / `.ap.json`、`ap-` 前缀产物名、`cmd/ap/` 路径段、`package ap`、`import ap "…"`、`ap.NewAgent` 这类限定符、`ap init` 这类 CLI 调用。

**裸词 `AP` / `ap` 不会被自动改写**：它们同时是 CAP 定理里的 Availability（`CP 优先于 AP`）、下游局部变量（`for _, ap := range …`）和业务缩写，盲改会破坏语义。脚本会把这类命中列在 `manual review` 清单里（文件 + 命中数），需人工逐处确认。

脚本**不会**做以下事情（需人工处理，见 §5）：

- 不改写 git 历史 / commit message / tag；
- 不迁移存量 K8s CRD 资源、Grafana 面板与告警规则；
- 不重建 npm / PyPI / crates.io / VS Code Marketplace 的发布身份；
- 不重置 `.env`、CI secrets 或部署平台上的环境变量名。

## 4. 手工迁移要点

1. `go.mod`：把 `require` 中的旧模块路径换成新路径，版本号保持不变；保留 `/v7` 语义化导入版本后缀。
2. Go 源码：import 路径与包限定符一起改。若旧代码用隐式包名 `ap.NewAgent`，新包声明为 `package synth`，必须同步改成 `synth.NewAgent`，否则编译器报未定义标识符。
3. 运行 `go mod tidy && go build ./... && go test ./...`。
4. 配置文件改名后，检查 CI / 启动脚本里读取 `.ap.yaml` 的硬编码路径。
5. 环境变量前缀替换后，逐个核对部署清单、`.env`、CI secrets。
6. K8s：CRD group 改变意味着**新建资源**而非原地升级；先导出旧资源、按新 group 重建，再切换 HPA 与仪表盘查询。
7. 镜像仓库：所有镜像引用统一为 `ghcr.io/mirelth/mirelth-synth`（Studio / Operator 为 `ghcr.io/mirelth/mirelth-synth-studio` / `ghcr.io/mirelth/mirelth-synth-operator`），与 `release.yml` 推送的 `ghcr.io/${{ github.repository }}` 保持一致；Helm `values.yaml`、Terraform 变量、K8s 清单里残留的旧 namespace 会导致 `ImagePullBackOff`。
8. kv 插件数据：`ecosystem/plugins/kv` 的默认库名由旧前缀改为 `synth_kv.db`、表名改为 `synth_kv_store`。插件在 `Init` 时会自动兼容：未指定 `db_path` 时若新库不存在而旧库存在，继续使用旧库；发现旧表 `ap_kv_store` 且新表不存在时，执行 `ALTER TABLE … RENAME TO` 原地迁移，存量 KV 数据不会丢。若自行维护了旧库副本，需手工执行同样的重命名。

## 5. 验证清单

- [ ] `go build ./...` 与 `go test ./...` 通过
- [ ] `synth version` 可运行，配置文件被正确加载
- [ ] `printenv | grep '^AP_'` 无输出，`SYNTH_*` 已注入
- [ ] 新 Prometheus 查询使用 `synth_*` 前缀，告警规则已更新
- [ ] 集群中 `kubectl get crd` 显示新 group，旧资源已按计划下线
- [ ] 发布渠道凭证、镜像仓库 namespace 已切换到新身份
- [ ] `bash scripts/check-brand-residuals.sh` 在本仓库通过（防止改名回潮）

## 6. 回滚

自动迁移前请先确保目标项目处于干净的 git 工作区。迁移脚本只改工作区文件，不创建提交；回滚用目标项目自己的版本控制（例如 `git checkout -- .` 并清理新文件名）。

## 7. 参考标准

- Go Modules 语义化导入版本规则：`/v7` 后缀是模块路径的一部分，改名不重置主版本。
- 版本真值与兼容性承诺：[`synth/docs/版本规范.md`](../synth/docs/版本规范.md)。
- 文档权威入口：[`docs/文档权威索引.md`](文档权威索引.md)。
- 迁移脚本的错误输出风格对齐 `scripts/version-check.sh`（CI annotation `::error::`）。

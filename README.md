# Synth

**Build AI agents in Go — compiled, concurrent-native, single binary, zero CGO**

[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26+-00ADD8.svg)](https://golang.org)
<a href="https://github.com/Mirelth/Mirelth-Synth/releases"><img src="https://img.shields.io/github/v/release/Mirelth/Mirelth-Synth?style=flat-square" alt="Latest release"></a>
<a href="https://github.com/Mirelth/Mirelth-Synth/actions/workflows/ci.yml"><img src="https://github.com/Mirelth/Mirelth-Synth/actions/workflows/ci.yml/badge.svg" alt="CI"></a>

<p align="center">
  <img src="synth/docs/synth-architecture.svg" alt="Synth Architecture" width="90%">
  <br>
  <sub>Interactive architecture diagram (zoom / theme switch / guided views):
  <a href="synth/docs/synth-architecture-v7.3.html">synth-architecture-v7.3.html</a></sub>
</p>

```bash
# Install the CLI straight from the module proxy — no clone, no replace directive
go install github.com/Mirelth/Mirelth-Synth/v7/synth/cmd/synth@latest

# Scaffold and run your first agent — no API key required
synth init my-agent --template quickstart
cd my-agent
go mod tidy
go run .
```

**About a minute. Zero config. No API key.** The agent is already running: with
`SYNTH_LLM_API_KEY` unset the scaffolded project runs in Demo mode, so the full ReAct loop is
real and only the model is simulated.

中文说明见 [README.zh-CN.md](README.zh-CN.md)。

---

## Why Synth

| | Synth | Other frameworks |
|---|---|---|
| **Getting started** | `go install` + `synth init`, then `go run .` — no key needed for the first run | install deps, configure keys, write boilerplate |
| **Language** | Go — compiled, concurrent-native, single binary | Python — runtime dependency churn |
| **Deployment** | Single binary + SQLite, zero infrastructure | Docker + Redis + Postgres + … |
| **Interop** | A2A protocol native, with a self-check (`synth a2a interop-check`) | bespoke integrations |
| **Evidence** | Claims are measured and published; gates are pre-registered | marketing tables |

## Core capabilities

- **ReAct engine** — reasoning + acting loop with per-turn checkpoints, crash resume, streaming, context trimming, and 30+ lifecycle hooks
- **Learning loop you can inspect** — tool success rates and task outcomes are recorded and surfaced (`synth profile`, `synth profile history`); skills learned are persisted and injected back into the loop. Capability *gains* are measured per capability, never promised wholesale
- **Orchestration** — Pipeline / Handoff / DAG / GroupChat / Debate / MapReduce, plus a Pool scheduler with bounded concurrency, session isolation and autoscaling
- **Tool system** — filesystem / shell / web / api / database / code-execution built-ins with scope validation, MCP integration, signed plugin marketplace
- **Three-tier memory** — SQLite (FTS5) + vector store (HNSW) + RAG hybrid retrieval (RRF fusion)
- **10+ LLM providers** — OpenAI / Anthropic / Gemini / Ollama (local) / Azure / Qwen / GLM / DeepSeek / Mistral / Cohere, behind one interface with retry, fallback chain and circuit breaker
- **A2A protocol** — agent discovery, task delegation and result return as a standard protocol (`synth a2a interop-check`)
- **Security** — ACL / sandbox / guardrails (injection, PII, topic) / path-traversal and symlink-escape protection; agent-generated code executes only inside a signed WASM sandbox
- **Observability** — Prometheus metrics, OpenTelemetry OTLP export, 35 structured error codes, Grafana-ready dashboards
- **Kubernetes Operator** — `AgentDeployment` CRD with declarative rollout and HPA autoscaling

## Quick start

### Install

```bash
# From the module proxy (recommended — the framework is a normal Go module)
go install github.com/Mirelth/Mirelth-Synth/v7/synth/cmd/synth@latest

# Or build from a source checkout, at the repository root
go build -o synth ./synth/cmd/synth/
```

### First run in about a minute (no API key)

```bash
synth init my-agent --template quickstart
cd my-agent
go mod tidy
go run .          # or: synth run  (build + hot reload)
```

That is the whole path. `synth init` writes `main.go`, `.synth.yaml`, `go.mod` and `.gitignore`;
`go run .` builds and executes. Without `SYNTH_LLM_API_KEY` you get Demo mode — a
keyword-matching provider that exercises the real loop before you spend a token.

`synth start my-agent` is a one-command variant that also resolves dependencies, but it
currently requires a framework source checkout to be reachable (in-repo, or via `SYNTH_ROOT`);
from a clean machine use the `synth init` path above.

### Configure a real LLM

```bash
synth config set api-key sk-xxx
synth config set provider openai
synth config set model gpt-4o
go run .           # this time with a real model
```

Environment variables win over `.synth.yaml`: `SYNTH_LLM_API_KEY`, `SYNTH_LLM_BASE_URL`,
`SYNTH_LLM_MODEL`. Any OpenAI-compatible endpoint works.

### Inspect what the agent learned

```bash
synth profile                 # capability profile: success rates, strengths, weak areas
synth profile history         # growth events over time
```

```
Agent Growth Profile
====================

Overall: 78% success (14/18 tasks), avg 4.2 turns

Top Capabilities:
  1. code-review     92% (23/25)  improving ↑
  2. data-analysis   85% (17/20)  stable →

Weak Areas:
  1. sql-migration   45% (5/11)   declining ↓
```

*(sample output — numbers are illustrative)*

### Learning dashboard

Studio is a separate binary in the framework module; it binds to `:8090` by default and
serves the learning dashboard at `/dashboard/learning`:

```bash
# from the repository root
go run ./synth/cmd/studio          # real engines by default
go run ./synth/cmd/studio -demo    # built-in demo data
# open http://localhost:8090/dashboard/learning
```

Capability radar, growth curves and per-domain detail — what the agent recorded, visible.

## Code examples

### Minimal agent

```go
provider, err := synth.NewOpenAIProvider(synth.Config{APIKey: os.Getenv("SYNTH_LLM_API_KEY"), Model: "gpt-4o-mini"})
if err != nil {
    log.Fatal(err)
}

agent, err := synth.NewAgent("HelloAgent", "You are a helpful assistant.", provider, synth.WithMaxTurns(10))
if err != nil {
    log.Fatal(err)
}

resp, err := agent.Run(ctx, synth.UserMessage("hello"))
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Content)
```

### Agent with tools + memory

```go
registry, err := synth.DefaultToolkit(synth.ToolkitConfig{RootDir: ".", EnableFS: true, EnableShell: true})
if err != nil {
    log.Fatal(err)
}
memory, err := synth.WithInMemory()
if err != nil {
    log.Fatal(err)
}
defer memory.Close()

agent, err := synth.NewAgent("CodingAssistant", "", provider,
    synth.WithMaxTurns(20),
    synth.WithToolkit(registry),
    synth.WithMemory(memory),
)
if err != nil {
    log.Fatal(err)
}

resp, err := agent.Run(ctx, synth.UserMessage("Review main.go for bugs"))
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Content)
```

### Concurrent multi-agent scheduling

```go
pool := synth.NewPool(synth.PoolConfig{MaxConcurrency: 5})
defer pool.Close()
pool.SetModel(provider)

results, err := pool.Dispatch(ctx, []synth.TaskConfig{
    {ID: "task-1", Title: "Code analysis", Prompt: "Analyze main.go"},
    {ID: "task-2", Title: "Run tests", Prompt: "Run go test"},
})
if err != nil {
    log.Fatal(err)
}
for _, r := range results {
    fmt.Println(r.TaskID, r.Status, r.Duration)
}
```

### DAG orchestration

```go
dag, err := synth.NewDAGBuilder("pipeline").
    Node("collect", func(ctx context.Context, in string) (string, error) { return "raw: " + in, nil }).
    Node("analyze", func(ctx context.Context, in string) (string, error) { return "analyzed: " + in, nil }).
    Node("report", func(ctx context.Context, in string) (string, error) { return "report: " + in, nil }).
    Edge("collect", "analyze").
    Edge("analyze", "report").
    Build()
if err != nil {
    log.Fatal(err)
}

result, err := dag.Run(ctx, "sales data")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Succeeded, "/", result.TotalNodes, "nodes")
```

### A2A interop

```go
// Delegate a task to a peer agent over the A2A protocol
client := synth.NewOpenInteropClient("https://peer.example.com")

task, err := client.SendTask(ctx, synth.OpenMessage{
    Role:  "user",
    Parts: []synth.OpenPart{{Type: "text", Text: "Summarize this repository"}},
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(task.ID)
```

To *serve* delegated tasks, wrap an `synth.TaskExecutor` (`Execute` / `Cancel`) and register it:
`synth.NewOpenInteropServer(card, cfg).WithExecutor(myExecutor)`.

Every snippet above was compiled **and executed** against the published module (v7.5.1),
in Demo mode where no API key is set.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Your Application                    │
├─────────────────────────────────────────────────────┤
│              Synth Framework                │
│                                                      │
│  ┌──────────┐ ┌──────────┐ ┌────────┐ ┌─────────┐  │
│  │ ReActLoop│ │   Pool   │ │  DAG   │ │Pipeline │  │
│  └────┬─────┘ └────┬─────┘ └───┬────┘ └────┬────┘  │
│       └────────────┼───────────┼───────────┘        │
│              ┌─────┴─────┐                           │
│              │Tool System│                           │
│              │Built-in + MCP + Plugin                │
│              └───────────┘                           │
│                                                      │
│  ┌──────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐  │
│  │ Memory   │ │ Learning │ │ Metrics│ │Guardrails│  │
│  │SQLite+FTS│ │SelfModel │ │OTel/   │ │ACL/Sandbox│ │
│  │+Vector   │ │+GrowthLog│ │Prom    │ │+PII      │  │
│  │+RAG      │ │+Studio   │ │        │ │          │  │
│  └──────────┘ └──────────┘ └────────┘ └──────────┘  │
└─────────────────────────────────────────────────────┘
```

## Project structure

```
Synth/
├── synth/          # core framework module
│   ├── cmd/synth/              # the CLI
│   ├── internal/            # 30+ internal packages
│   │   ├── agent/           # ReAct engine + microkernel (30+ sub-packages)
│   │   ├── pool/            # multi-agent scheduling
│   │   ├── tools/           # tool system
│   │   ├── memory/          # memory store + SelfModel + GrowthLog
│   │   ├── llm/             # 10+ LLM providers
│   │   ├── studio/          # Studio engine + learning dashboard
│   │   └── ...              # full list: internal/AGENTS.md
│   ├── pkg/                 # public API
│   ├── operator/            # Kubernetes Operator (separate Go module)
│   ├── bench/               # benchmarks and eval suites
│   └── ecosystem/           # examples + plugins + templates
├── pgvector/                # pgvector vector-store extension
├── gateway/                 # gateway
├── wasm/                    # WASM sandbox (wazero)
└── sdk/typescript/          # TypeScript SDK (feature parity with Go)
```

## CLI commands

```bash
synth start my-agent              # create + install deps + run (needs a framework checkout)
synth init my-agent               # scaffold a project (manual mode)
synth run                         # build and run the current project (--watch hot reload)
synth profile                     # agent growth report
synth profile history             # growth event history
synth config set api-key sk-xxx   # configure the API key
synth debug                       # debug server
synth loop trace                  # execution trace
synth test                        # eval test suite
synth mcp list / add / test       # MCP server management
synth plugin install <url>        # install a plugin (signature verified)
synth marketplace install <url>   # install an agent template (signature verified)
synth cluster init / join / status / leave / scale
synth autonomy run / list / resume / status
synth skill list / add / remove / verify
synth a2a interop-check           # A2A protocol self-check
synth doctor                      # health check
synth completion                  # shell completion scripts
```

## Examples

| Example | What it shows | Run |
|---|---|---|
| code-review-agent | code review agent with a learning loop | `go run ./synth/ecosystem/examples/code-review-agent/` |
| data-analysis-agent | data analysis agent | `go run ./synth/ecosystem/examples/data-analysis-agent/` |
| a2a-handoff | **cross-process A2A handoff: two processes, zero credentials** | `./synth/ecosystem/examples/a2a-handoff/run.sh` |
| a2a-connect | A2A protocol interop | `go run ./synth/ecosystem/examples/a2a-connect/` |
| github-issue-triage | automatic GitHub issue triage | `go run ./synth/ecosystem/examples/github-issue-triage/` |
| multi-agent | Pool concurrent scheduling | `go run ./synth/ecosystem/examples/multi-agent/` |
| chain-api | minimal chain API | `go run ./synth/ecosystem/examples/chain-api/` |

Run examples from a source checkout at the repository root.

## Design philosophy

1. **Born in production** — the core patterns were extracted from a production agent system, not invented for a demo
2. **Interfaces first** — LLM / tools / memory are all behind interfaces, freely replaceable
3. **Zero-config start** — `synth init` + `go run .` needs no configuration to run
4. **Measured, not promised** — capability claims carry evidence anchors; the ones that failed their gates are documented as failures
5. **Minimal dependencies** — the core is CGO-free (SQLite + YAML only); gRPC / Redis / etcd / wazero are opt-in
6. **TDD enforced** — Red → Green → Refactor, with a CI gate on coverage tiers, deprecation residue and API contract drift

## Tech stack

- **Go 1.26+** — compiled, concurrent-native, single-binary deployment
- **SQLite** — embedded memory store (`modernc.org/sqlite`, pure Go, zero CGO)
- **gRPC** — A2A protocol transport
- **wazero** — WASM sandbox execution (pure Go, zero CGO)
- **OpenTelemetry** — observability standard

## What we deliberately do not claim

Honesty is part of the API surface. Three capabilities are intentionally kept off the
headline table:

- **The world model is opt-in.** `synth.WithWorldModel()` stays off by default: the
  pre-registered experiment ran 112 pairs, A 89/112 vs B 95/112 (+5.4pp, McNemar p=0.1796),
  and the gate (+15pp and p<0.05) was not met — flipping the default was formally rejected.
  Registry: [docs/降级豁免登记-V7弧线.md](docs/降级豁免登记-V7弧线.md) (E-1).
- **`EnhancedPlanner` is not a selling point.** Its recursive decomposition measured
  **−47.8pp (p=0.001)** on short-horizon ReAct tasks
  ([docs/v7.2-验收实验报告.md](docs/v7.2-验收实验报告.md)); it ships as `Experimental` behind
  planning-budget guardrails.
- **Agent-authored tools are research-blocked.** The sandbox half is real — agent-generated
  artifacts execute only inside the signature-verified WASM sandbox and the host never
  compiles or loads agent code — but the generator has no production implementation.

What we do claim: one provider interface with resilient calling, a ReAct loop with
per-turn checkpoints and resume, three-tier memory, pool scheduling, sandbox/ACL, A2A
protocol interop, and a reproducible evaluation culture (pre-registered gates, quarterly
bootstrap comparisons, published reports).

## Documentation

- **[Doc authority index](docs/文档权威索引.md)** — the single entry point for "which of 260+ documents do I trust" (version truth / current facts / plans / history)
- **[Experimental capability registry](docs/实验性能力清单.md)** — which capabilities are really wired, which are demo-only or shells
- **[English front door](synth/README.md)** · [English getting started](synth/docs/getting-started/five-minutes.md) · [English handbook](synth/docs/Synth-开发手册-EN.md)
- [v7.5 release notes](synth/docs/CHANGELOG-v7.8.md) · [Three-track user guide](synth/docs/三战线用户指南.md) · [Demo video script](synth/docs/demo-video-script.md)
- [Architecture](docs/架构图.md) · [API reference](docs/API参考.md) · [Version policy](synth/docs/版本规范.md)
- [TypeScript SDK](sdk/typescript/README.md) · [Getting started](synth/ecosystem/docs/getting-started.md) · [CLI guide](synth/ecosystem/docs/synth-guide.md) · [Best practices](synth/ecosystem/docs/best-practices.md) · [Cookbook](synth/ecosystem/docs/cookbook/)
- [Internal module map](synth/internal/AGENTS.md) · [Full changelog](docs/CHANGELOG.md)
- **[Migration to Synth](docs/迁移到-Synth.md)** — module path / API / CLI / env contract changes and the automated migration script
- **[Brand assets & history boundary](docs/品牌资产与历史说明.md)** — binary assets, git history and the residual-name CI gate

## Version history

| Version | Theme | Highlights |
|---|---|---|
| **v7.9.0** | Rebrand release | Module path `github.com/Mirelth/Mirelth-Synth/v7`, `package synth`, CLI `synth`, `.synth.yaml` / `SYNTH_*`, CRD group `synth.dev`, `mirelth-synth` / `@mirelth/synth` distribution names; migration guide + script, residual-name CI gate, regenerated brand assets. Breaking: import paths, package name and config contract all change — see [Migration to Synth](docs/迁移到-Synth.md) |
| **v7.8.0** | Cumulative release: 7.6–7.8 | English front door, clean-room CI, API surface freeze, A2A interop demo, signed WASM supply-chain demo, v8.0 deprecation markers + `synth migrate`; 7.6/7.7 not separately released |
| **v7.5.1** | Release pipeline fix | 6-platform binaries + checksums + cosign signatures + SBOM; `go install …@latest` works end to end |
| v7.5 | Security & quality campaign | shell allowlist hardened against 7 verified attack chains, `jsonutil` concurrency corruption fixed, 167 disabled tests restored; module path moved to `/v7` so the framework is distributable via `GOPROXY` |
| v7.4 | Wiring batch | OTel end-to-end export, Studio on real engines, signed template registry, tool-learner auto-assembly, skill library into the loop, autonomy CLI real execution |
| v7.3 | Three tracks | zero-barrier entry, visible learning loop, A2A ecosystem interop |
| v7.2 | A/B acceptance | full acceptance experiments (P2 EnhancedPlanner came out significantly negative — see above) |
| v7.0–7.1 | Five components | distillation, gap tools, resident runtime, federation; world model shipped opt-in |
| v6.0 | Consolidation | 39 modules frozen, cognitive engine + self-evolution loop |
| v5.0 | Balance | real wiring + multimodal + distributed + platform |
| v4.0 | Stability | contract lock + performance + compatibility tightening |
| v1–3 | Incubation | core engine + microkernel + dual language + ecosystem |

Full history: [CHANGELOG](docs/CHANGELOG.md).

## License

Apache-2.0 © Synth Contributors

---

<!--
  Machine-checked anchors for scripts/doc-version-consistency.mjs.
  These lines are deliberately not rendered: the gate's regexes are Chinese-phrased and
  numeric, so the visible English text above stays qualitative and non-rotting while the
  gate still forces an update whenever the measured numbers move.
  Update together with the code, then re-run: node scripts/doc-version-consistency.mjs

  [![Version](https://img.shields.io/badge/version-7.9.0-2ea44f.svg)](synth/docs/CHANGELOG-v7.9.md)
  [![Tests](https://img.shields.io/badge/tests-686%20files-green.svg)](synth/internal/)
  Red → Green → Refactor，686 个测试文件
  核心框架主模块（163 个包，1397 个 Go 文件）
  微内核（34 个子包）
-->

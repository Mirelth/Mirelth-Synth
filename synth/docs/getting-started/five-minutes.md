# Five-Minute Quickstart

Create and run your first Synth agent in five minutes — **without an API key**.

> 中文版：[五分钟入门.md](五分钟入门.md)

## Prerequisites

- Go 1.26 or newer (`go version`)
- A terminal. That is it.

## Step 1 — Install the CLI

```bash
go install github.com/Mirelth/Mirelth-Synth/v7/synth/cmd/synth@latest

synth version
```

## Step 2 — Scaffold a project

Run this from **any directory** — it does not need to live inside the framework repository:

```bash
synth init my-agent --template quickstart
cd my-agent
```

`synth init` writes four files:

| File | Contents |
|------|----------|
| `main.go` | Minimal public-API sample: read config → build a provider → run the ReAct loop once → print stats |
| `.synth.yaml` | Project config: `llm` (provider/model/api_key), `memory` (backend, path), `agent` (max_turns, system_prompt) |
| `go.mod` | Dependency declaration pointing at the current released framework version |
| `.gitignore` | Ignore rules |

Resolve dependencies once:

```bash
go mod tidy
```

## Step 3 — Run it in Demo mode

```bash
go run .
```

With `SYNTH_LLM_API_KEY` unset the generated project prints:

```
Synth 快速入门
=========================

提示: 未设置 SYNTH_LLM_API_KEY，使用 Demo 模式
   Demo 模式基于关键词匹配产生响应，用于体验基本功能
   …
```

Demo mode swaps in `synth.NewDemoProvider()`, a keyword-matching provider, so the **whole
pipeline is real** — agent construction, the ReAct loop, turn accounting, tool dispatch
machinery and stats — only the model weights are simulated. Use it to learn the shape of the
API before spending a cent.

What you should see at the end of the run:

```
运行统计:
   状态: completed
   当前轮数: 1
   消息总数: 2
   工具调用: map[]
```

## Step 4 — Switch to a real model (optional)

```bash
export SYNTH_LLM_API_KEY=sk-...
export SYNTH_LLM_BASE_URL=https://api.openai.com/v1   # any OpenAI-compatible endpoint
export SYNTH_LLM_MODEL=gpt-4o-mini

go run .
```

Environment variables win over `.synth.yaml`. Alternatively keep the key out of the shell:
`synth config set api-key`, or set `llm.api_key` in `.synth.yaml` (do not commit it).

For a build-and-hot-reload loop while you edit, use `synth run` instead of `go run .`:

```bash
synth run        # watches sources, rebuilds and restarts on change
```

## What just happened

`main.go` is deliberately short — read it end to end, it is the fastest way to learn the
framework:

```go
apiKey := os.Getenv("SYNTH_LLM_API_KEY")

var provider synth.Provider
if apiKey == "" {
    provider = synth.NewDemoProvider()          // no key → Demo mode
} else {
    cfg := synth.ConfigFromEnv("")              // reads SYNTH_LLM_* env vars
    provider, _ = synth.NewOpenAIProvider(cfg)  // real LLM
}

agent, _ := synth.NewAgent("QuickStartAgent", "You are a helpful assistant.", provider,
    synth.WithMaxTurns(10))

resp, _ := agent.Run(ctx, synth.UserMessage("你好！请用一句话介绍自己"))
```

The loop behind `agent.Run` is the ReAct engine: the model reasons, optionally calls tools,
observes the results, and repeats until it answers or hits `max_turns`. Every turn is
checkpointed, so a crashed run resumes instead of restarting.

## Other templates

```bash
synth init <name> --template basic              # minimal skeleton
synth init <name> --template with-tools         # filesystem / shell / web tools wired in
synth init <name> --template agent-with-rag     # knowledge-base QA (RAG)
synth init <name> --template agent-with-cache   # provider caching
synth init <name> --template agent-with-metrics # Prometheus metrics
synth init <name> --template multi-agent        # multi-agent collaboration
```

## FAQ

**Q: Do I need an API key to finish this guide?**
No. Demo mode covers the whole first run. Keys are only needed for real model output.

**Q: Can I use a local model?**
Yes — point `SYNTH_LLM_BASE_URL` at Ollama (`http://localhost:11434`) and set `SYNTH_LLM_MODEL`
(for example `llama3.1`). No key required.

**Q: How do I change the agent's behaviour?**
Edit `agent.system_prompt` in `.synth.yaml`, or pass a different prompt to `synth.NewAgent`. That
single string is your first custom agent.

**Q: How do I see what the agent is doing?**
`synth loop trace` replays the ReAct execution trace, `synth loop inspect` shows a run in detail,
`synth debug` starts the debug server on <http://localhost:6060>, and `synth profile` shows the
agent's growth profile.

**Q: Where is memory stored?**
`./data/memory.db` (SQLite with FTS5 full-text and vector columns) by default; set
`memory.backend: memory` for a volatile in-process store.

## Next steps

- 📖 [Build your first agent from scratch](first-agent.md)
- 🧠 [Memory system](../concepts/记忆系统.md) · 🛠️ [Tool system](../concepts/工具系统.md)
- 👥 [Multi-agent orchestration](../concepts/多Agent编排.md)

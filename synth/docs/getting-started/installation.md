# Installation Guide

This guide installs the Synth CLI (`synth`) and prepares your machine for agent
development. Every command below was verified against a released version of the framework.

> 中文版：[安装指南.md](安装指南.md)

## System requirements

| Requirement | Notes |
|---|---|
| Go ≥ 1.26 | The `toolchain` directive in `go.mod` downloads the pinned toolchain automatically |
| OS | Windows 10+, macOS 10.15+, Linux (Ubuntu 18.04+, CentOS 7+) |
| Memory | 512 MB minimum, 2 GB+ recommended |
| Disk | 100 MB minimum (the module cache grows with your dependencies) |
| LLM API key | Optional for the first run — Demo mode needs no key. Pick any of: OpenAI / Anthropic / Gemini / Ollama (local, no key) / Azure / DeepSeek / GLM / Cohere / Mistral / Qwen |

## Method 1: `go install` (recommended)

The framework is published as a regular Go module, so the CLI installs straight from the
module proxy — **no clone, no `replace` directive, no local build**:

```bash
go install github.com/Mirelth/Synth/v7/synth/cmd/synth@latest
```

Verify:

```bash
synth version
```

> If you want a specific release, pin the tag: `…/cmd/synth@v7.8.0`. Release binaries with
> SHA256 checksums and cosign signatures are attached to every
> [GitHub Release](https://github.com/Mirelth/Synth/releases).

### Installing Go first

**Windows (PowerShell):**

```powershell
choco install go
# or download the installer from https://go.dev/dl/
```

**macOS:**

```bash
brew install go
# or download the installer from https://go.dev/dl/
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt update && sudo apt install golang-go
```

**Linux (CentOS/RHEL):**

```bash
sudo yum install golang
```

Verify the toolchain:

```bash
go version
# go version go1.26.x linux/amd64 (or similar)
```

## Method 2: build from source

Only needed when you hack on the framework itself. Clone, then build the CLI:

```bash
git clone https://github.com/Mirelth/Synth.git
cd Synth

go build -o synth ./synth/cmd/synth/
./synth version
```

Windows produces `synth.exe` from the same command. Add the binary to your `PATH`, or install it
into the Go bin directory:

```bash
go install ./synth/cmd/synth
```

## Method 3: Docker

Every release publishes a signed image to GitHub Container Registry:

```bash
docker pull ghcr.io/mirelth/synth:<release-tag>

docker run -it --rm \
  -e SYNTH_LLM_API_KEY="$SYNTH_LLM_API_KEY" \
  ghcr.io/mirelth/synth:<release-tag> --help
```

Replace `<release-tag>` with a tag from the
[Releases page](https://github.com/Mirelth/Synth/releases) (for example `v7.8.0`).

## Method 4: TypeScript SDK (Node.js)

For TypeScript projects there is no CLI to install — the SDK is an npm package:

```bash
npm install @mirelth/synth
# optional, for SQLite persistence
npm install better-sqlite3
```

## Verify the installation

```bash
synth version                                   # CLI version
synth doctor                                    # four-point checkup: Go, project config, API key, dependencies
synth init smoke-test --template quickstart     # scaffold a throwaway project
cd smoke-test && go mod tidy && go run .     # first agent, Demo mode, no API key
```

`synth doctor` reports `⚠ no LLM API key detected` until you configure credentials — that is
expected on a fresh machine and does not block Demo mode.

## Configure LLM credentials

Credentials are read from the environment; `.synth.yaml` is the fallback (the environment
variable wins).

| Provider | Environment variables |
|---|---|
| OpenAI | `SYNTH_LLM_API_KEY` (or `OPENAI_API_KEY`), `SYNTH_LLM_BASE_URL`, `SYNTH_LLM_MODEL` |
| Anthropic | `SYNTH_LLM_API_KEY`, `SYNTH_LLM_BASE_URL`, `SYNTH_LLM_MODEL` |
| Gemini | `SYNTH_LLM_API_KEY`, `SYNTH_LLM_MODEL` |
| Ollama (local) | `SYNTH_LLM_BASE_URL=http://localhost:11434`, `SYNTH_LLM_MODEL=llama3.1` — no key needed |
| Azure OpenAI | `SYNTH_LLM_API_KEY`, `SYNTH_LLM_BASE_URL`, `SYNTH_LLM_MODEL` |
| DeepSeek / GLM / Qwen | `SYNTH_LLM_API_KEY`, `SYNTH_LLM_BASE_URL`, `SYNTH_LLM_MODEL` |

Any OpenAI-compatible endpoint works — point `SYNTH_LLM_BASE_URL` at it and set the model name.

```bash
# Linux / macOS
export SYNTH_LLM_API_KEY=sk-...
export SYNTH_LLM_BASE_URL=https://api.openai.com/v1
export SYNTH_LLM_MODEL=gpt-4o-mini

# Windows PowerShell
$env:SYNTH_LLM_API_KEY="sk-..."
```

Persist them in your shell profile (`~/.bashrc`, `~/.zshrc`) or use
`synth config set api-key` inside a project. Never commit keys into `.synth.yaml`.

## Troubleshooting

### `synth: command not found`

The Go bin directory is not on your `PATH`:

```bash
go env GOPATH                 # e.g. /Users/you/go
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Permission denied running the binary

```bash
chmod +x "$(go env GOPATH)/bin/synth"
```

### Module download fails behind a proxy

```bash
export HTTPS_PROXY=http://proxy.example.com:8080
# or a regional module mirror
export GOPROXY=https://goproxy.cn,direct
```

### Go version too old

`synth` and the generated projects require Go 1.26+. Upgrade with `brew upgrade go`, your
package manager, or the installer from <https://go.dev/dl/>.

### Scaffolded project cannot resolve the framework

`synth init` writes a `go.mod` that requires the current released framework version, and
`go mod tidy` resolves it from the module proxy. If you work **inside** a framework
checkout, the scaffold wires a local `replace` instead — build with `GOWORK=off` or add the
project to the workspace `go.work` (`go work use ./my-agent`). Details:
[Synth-开发手册-EN.md](../Synth-开发手册-EN.md) FAQ.

## Next steps

- 🚀 [Five-minute quickstart](five-minutes.md)
- 📖 [Build your first agent](first-agent.md)
- 📚 [Core concepts](../concepts/Agent架构.md)

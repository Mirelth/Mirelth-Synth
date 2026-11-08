# 创建 Agent

本指南详细介绍如何创建和配置 Agent。

## 基础创建

=== "Go"

    ```go
    import (
        "os"
        synth "github.com/Mirelth/Synth/v7/synth/pkg"
    )

    provider, err := synth.NewOpenAIProvider(synth.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
        Model:  "gpt-4o",
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, err := synth.NewAgent("my-agent", "你是一个智能助手", provider,
        synth.WithMaxTurns(10),
    )
    if err != nil {
        log.Fatal(err)
    }

    resp, err := agent.Run(ctx, synth.UserMessage("你好"))
    ```

=== "TypeScript"

    ```typescript
    import { ReActAgent, OpenAIProvider, ToolRegistry } from '@mirelth/synth';

    const agent = new ReActAgent({
      name: 'my-agent',
      model: new OpenAIProvider({ apiKey: process.env.OPENAI_API_KEY!, model: 'gpt-4o' }),
      toolkit: new ToolRegistry(),
      maxTurns: 10,
      systemPrompt: '你是一个智能助手',
    });

    const resp = await agent.run('你好');
    ```

## 配置 LLM

### OpenAI

=== "Go"

    ```go
    provider := synth.NewOpenAIProvider(synth.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
        Model:  "gpt-4o",
    })
    ```

=== "TypeScript"

    ```typescript
    const provider = new OpenAIProvider({
      apiKey: process.env.OPENAI_API_KEY!,
      model: 'gpt-4o',
    });
    ```

### Anthropic Claude

=== "Go"

    ```go
    provider := synth.NewAnthropicProvider(synth.Config{
        APIKey: os.Getenv("ANTHROPIC_API_KEY"),
        Model:  "claude-3-5-sonnet-20241022",
    })
    ```

=== "TypeScript"

    ```typescript
    import { AnthropicProvider } from '@mirelth/synth';

    const provider = new AnthropicProvider({
      apiKey: process.env.ANTHROPIC_API_KEY!,
      model: 'claude-3-5-sonnet-20241022',
    });
    ```

### Ollama（本地模型）

=== "Go"

    ```go
    provider := synth.NewOllamaProvider(synth.Config{
        BaseURL: "http://localhost:11434",
        Model:   "llama3",
    })
    ```

=== "TypeScript"

    ```typescript
    import { OllamaProvider } from '@mirelth/synth';

    const provider = new OllamaProvider({
      baseURL: 'http://localhost:11434',
      model: 'llama3',
    });
    ```

### ResilientProvider（推荐生产使用）

内置重试、熔断和降级：

=== "Go"

    ```go
    primary, _ := synth.NewOpenAIProvider(synth.Config{APIKey: key, Model: "gpt-4o"})
    fallback, _ := synth.NewGeminiProvider(synth.Config{APIKey: geminiKey, Model: "gemini-1.5-pro"})

    resilient, _ := synth.NewResilientProvider(primary, synth.DefaultResilientConfig())
    resilient.AddFallback(fallback)

    agent, err := synth.NewAgent("assistant", "你是助手", resilient,
        synth.WithMaxTurns(10),
    )
    if err != nil {
        log.Fatal(err)
    }
    ```

=== "TypeScript"

    ```typescript
    import { ResilientProvider } from '@mirelth/synth';

    const resilient = new ResilientProvider(primaryProvider);
    resilient.addFallback(fallbackProvider);

    const agent = new ReActAgent({
      name: 'assistant',
      model: resilient,
      toolkit: new ToolRegistry(),
      maxTurns: 10,
    });
    ```

## 配置工具

### 注册内置工具

=== "Go"

    ```go
    registry, _ := synth.DefaultToolkit(synth.ToolkitConfig{
        RootDir:     ".",
        EnableFS:    true,
        EnableShell: true,
        EnableWeb:   true,
    })

    agent := synth.NewAgent("assistant", "你是助手", provider,
        synth.WithMaxTurns(10),
        synth.WithToolkit(registry),
    )
    ```

=== "TypeScript"

    ```typescript
    import { ToolRegistry, FileSystemTool, ShellTool, WebTool } from '@mirelth/synth';

    const registry = new ToolRegistry();
    registry.register(new FileSystemTool({ rootDir: '.' }));
    registry.register(new ShellTool({ allowedCommands: ['ls', 'cat'] }));
    registry.register(new WebTool());

    const agent = new ReActAgent({
      name: 'assistant',
      model: provider,
      toolkit: registry,
      maxTurns: 10,
    });
    ```

### 注册自定义工具

=== "Go"

    ```go
    // 自定义工具只需实现 Tool 接口的 4 个方法
    type SearchTool struct{}
    func (t *SearchTool) Name() string        { return "search" }
    func (t *SearchTool) Description() string { return "搜索信息" }
    func (t *SearchTool) Parameters() json.RawMessage {
        return json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`)
    }
    func (t *SearchTool) Execute(ctx context.Context, args json.RawMessage) (*synth.ToolResult, error) {
        var params struct { Query string `json:"query"` }
        json.Unmarshal(args, &params)
        return &synth.ToolResult{Content: searchResults(params.Query)}, nil
    }

    registry := synth.NewToolRegistry()
    registry.Register(&SearchTool{})

    agent, err := synth.NewAgent("assistant", "你是助手", provider,
        synth.WithMaxTurns(10),
        synth.WithToolkit(registry),
    )
    if err != nil {
        log.Fatal(err)
    }
    ```

=== "TypeScript"

    ```typescript
    import { Tool, ToolRegistry } from '@mirelth/synth';

    class WeatherTool implements Tool {
      name = 'get_weather';
      description = 'Get current weather for a city';
      parameters = {
        type: 'object' as const,
        properties: {
          city: { type: 'string', description: 'City name' },
        },
        required: ['city'],
      };

      async execute(args: { city: string }): Promise<string> {
        return `Weather in ${args.city}: 22°C, sunny`;
      }
    }

    const registry = new ToolRegistry();
    registry.register(new WeatherTool());
    ```

## 配置记忆

=== "Go"

    ```go
    memory, _ := synth.WithInMemory()
    defer memory.Close()

    agent := synth.NewAgent("assistant", "你是助手", provider,
        synth.WithMaxTurns(10),
        synth.WithMemory(memory),
    )
    ```

=== "TypeScript"

    ```typescript
    import { InMemoryStore } from '@mirelth/synth';

    const memory = new InMemoryStore();

    const agent = new ReActAgent({
      name: 'assistant',
      model: provider,
      toolkit: new ToolRegistry(),
      maxTurns: 10,
      memory,
    });
    ```

## 生命周期钩子

=== "Go"

    通过 `WithHooks()` 注入：

    ```go
    hooks := synth.NewHookManager()
    hooks.Register(synth.HookBeforeLLM, func(ctx context.Context, hctx *synth.HookContext) error {
        log.Printf("开始处理")
        return nil
    })
    hooks.Register(synth.HookAfterTool, func(ctx context.Context, hctx *synth.HookContext) error {
        if hctx.ToolCall != nil {
            log.Printf("工具 %s 执行完成", hctx.ToolCall.Name)
        }
        return nil
    })

    agent, err := synth.NewAgent("assistant", "你是助手", provider,
        synth.WithMaxTurns(10),
        synth.WithHooks(hooks),
    )
    if err != nil {
        log.Fatal(err)
    }
    ```

=== "TypeScript"

    通过 `HookManager` 注入：

    ```typescript
    import { HookManager } from '@mirelth/synth';

    const hooks = new HookManager();
    hooks.register('before_llm', (ctx) => {
      console.log(`Turn ${ctx.turn}: 开始处理`);
    });
    hooks.register('after_tool', (ctx) => {
      console.log(`工具 ${ctx.toolCall?.name} 执行完成`);
    });

    const agent = new ReActAgent({
      name: 'assistant',
      model: provider,
      toolkit: new ToolRegistry(),
      maxTurns: 10,
      hooks,
    });
    ```

## 链式 API

=== "Go"

    ```go
    agent, err := synth.NewAgent("assistant", "你是助手", provider, synth.WithMaxTurns(10)).
        WithToolkit(toolkit).
        WithMemory(memory).
        WithRAG(ragProvider)
    if err != nil {
        log.Fatal(err)
    }
    ```

=== "TypeScript"

    ```typescript
    const agent = new ReActAgent({
      name: 'assistant',
      model: provider,
      toolkit: registry,
      maxTurns: 10,
      systemPrompt: '你是助手',
      memory,
      hooks,
    });
    ```

## 完整示例

=== "Go"

    ```go
    package main

    import (
        "context"
        "fmt"
        "os"

        synth "github.com/Mirelth/Synth/v7/synth/pkg"
    )

    func main() {
        // 1. LLM Provider
        provider, err := synth.NewOpenAIProvider(synth.Config{
            APIKey: os.Getenv("OPENAI_API_KEY"),
            Model:  "gpt-4o",
        })
        if err != nil {
            log.Fatal(err)
        }

        // 2. 工具
        toolkit, _ := synth.DefaultToolkit(synth.ToolkitConfig{
            RootDir:     ".",
            EnableFS:    true,
            EnableShell: true,
        })

        // 3. 记忆
        memory, _ := synth.WithInMemory()
        defer memory.Close()

        // 4. Agent（Functional Options）
        agent, err := synth.NewAgent("assistant", "你是一个专业的助手", provider,
            synth.WithMaxTurns(15),
            synth.WithToolkit(toolkit),
            synth.WithMemory(memory),
        )
        if err != nil {
            log.Fatal(err)
        }

        // 5. 运行
        resp, err := agent.Run(context.Background(), synth.UserMessage("帮我查询今天的天气"))
        if err != nil {
            fmt.Printf("错误: %v\n", err)
            return
        }

        fmt.Printf("结果: %s\n", resp.Content)
    }
    ```

=== "TypeScript"

    ```typescript
    import {
      ReActAgent, OpenAIProvider, ToolRegistry,
      FileSystemTool, ShellTool, InMemoryStore,
    } from '@mirelth/synth';

    async function main() {
      // 1. LLM Provider
      const provider = new OpenAIProvider({
        apiKey: process.env.OPENAI_API_KEY!,
        model: 'gpt-4o',
      });

      // 2. 工具
      const toolkit = new ToolRegistry();
      toolkit.register(new FileSystemTool({ rootDir: '.' }));
      toolkit.register(new ShellTool({ allowedCommands: ['ls', 'cat'] }));

      // 3. 记忆
      const memory = new InMemoryStore();

      // 4. Agent
      const agent = new ReActAgent({
        name: 'assistant',
        model: provider,
        toolkit,
        maxTurns: 15,
        systemPrompt: '你是一个专业的助手',
        memory,
      });

      // 5. 运行
      const resp = await agent.run('帮我查询今天的天气');
      console.log('结果:', resp.content);
    }

    main();
    ```

## 下一步

- 学习 [添加工具](添加工具.md) 的详细指南
- 查看 [配置记忆](配置记忆.md) 的完整选项
- 阅读 [多 Agent 协作](多Agent协作.md) 了解编排能力

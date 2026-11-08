package main

import (
	"context"
	"fmt"
	"log"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
	"github.com/Mirelth/Mirelth-Synth/v7/synth/testutil"
)

func main() {
	fmt.Println("=== Synth: 工具 + 记忆示例 ===")
	fmt.Println()

	// 使用 DefaultToolkit 快速配置工具集
	registry, err := synth.DefaultToolkit(synth.ToolkitConfig{
		RootDir:     ".",
		EnableFS:    true,
		EnableShell: true,
		EnableWeb:   true,
	})
	if err != nil {
		log.Fatalf("创建 DefaultToolkit 失败: %v", err)
	}

	memory, err := synth.WithInMemory()
	if err != nil {
		log.Fatalf("创建 Memory 失败: %v", err)
	}
	defer memory.Close()

	hooks := synth.NewHookManager()
	hooks.Register(synth.HookBeforeRun, func(ctx context.Context, hctx *synth.HookContext) error {
		fmt.Printf("[Hook] Agent %s 开始运行\n", hctx.AgentID)
		return nil
	})
	hooks.Register(synth.HookAfterRun, func(ctx context.Context, hctx *synth.HookContext) error {
		fmt.Printf("[Hook] Agent %s 运行完成\n", hctx.AgentID)
		return nil
	})

	mock := testutil.NewMockProvider("让我帮你读取文件内容。")

	agent, err := synth.NewAgent("TooledAgent", "你是一个可以读写文件、执行命令和访问网页的助手", mock, synth.WithMaxTurns(5))
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}
	agent = agent.WithToolkit(registry).
		WithMemory(memory).
		WithHooks(hooks)

	resp, err := agent.Run(context.Background(), synth.UserMessage("读取当前目录的文件"))
	if err != nil {
		log.Fatalf("Agent 运行失败: %v", err)
	}

	fmt.Printf("回复: %s\n", resp.Content)
	fmt.Printf("工具调用次数: %d\n", resp.Metrics.TotalTools)

	episodes, _ := memory.Search(context.Background(), "文件", nil)
	fmt.Printf("记忆条目数: %d\n", len(episodes))
}

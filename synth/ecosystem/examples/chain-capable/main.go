package main

import (
	"context"
	"fmt"
	"log"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

// MockLLM 是示例用的模拟 LLM
type MockLLM struct{}

func (m *MockLLM) Complete(ctx context.Context, req *synth.CompletionRequest) (*synth.CompletionResponse, error) {
	return &synth.CompletionResponse{
		ID: "mock-1", Content: "我已经读取了文件内容并保存到记忆中。",
		Role: "assistant", Usage: synth.Usage{PromptTokens: 10, CompletionTokens: 20},
	}, nil
}
func (m *MockLLM) Stream(ctx context.Context, req *synth.CompletionRequest) (<-chan synth.Chunk, error) {
	ch := make(chan synth.Chunk, 1)
	go func() { defer close(ch); ch <- synth.Chunk{Content: "处理中...", Done: true} }()
	return ch, nil
}
func (m *MockLLM) CallTools(ctx context.Context, req *synth.ToolCallRequest) (*synth.ToolCallResponse, error) {
	return &synth.ToolCallResponse{Usage: synth.Usage{}}, nil
}
func (m *MockLLM) Info() synth.ModelInfo {
	return synth.ModelInfo{Name: "mock", Provider: "mock", MaxContext: 4096, SupportsTools: true}
}

func main() {
	fmt.Println("=== 链式 API：渐进式添加能力 ===")
	fmt.Println()

	// 第 1 步：按需准备能力 —— 工具 + 记忆 + Hook
	registry, err := synth.DefaultToolkit(synth.ToolkitConfig{
		RootDir:     ".",
		EnableFS:    true,
		EnableShell: true,
		EnableWeb:   true,
	})
	if err != nil {
		log.Fatalf("创建工具集失败: %v", err)
	}

	mem, err := synth.WithInMemory()
	if err != nil {
		log.Fatalf("创建记忆存储失败: %v", err)
	}
	defer mem.Close()

	hooks := synth.NewHookManager()
	hooks.Register(synth.HookBeforeRun, func(ctx context.Context, hctx *synth.HookContext) error {
		fmt.Printf("[Hook] Agent 开始运行\n")
		return nil
	})
	hooks.Register(synth.HookAfterRun, func(ctx context.Context, hctx *synth.HookContext) error {
		fmt.Printf("[Hook] Agent 运行完成\n")
		return nil
	})

	// 第 2 步：创建 Agent 并一次性注入所有能力
	capableAgent, err := synth.NewAgent("capable-agent", "", &MockLLM{},
		synth.WithMaxTurns(5),
		synth.WithToolkit(registry),
		synth.WithMemory(mem),
		synth.WithHooks(hooks),
	)
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	// 运行
	resp, err := capableAgent.Run(context.Background(), synth.UserMessage("读取当前目录的文件"))
	if err != nil {
		log.Fatalf("运行失败: %v", err)
	}

	fmt.Printf("回复: %s\n", resp.Content)
	fmt.Printf("工具调用次数: %d\n", resp.Metrics.TotalTools)

	// 验证记忆已保存
	episodes, _ := mem.Search(context.Background(), "文件", nil)
	fmt.Printf("记忆条目数: %d\n", len(episodes))
}

package suite

// Phase 5.5: QPS 基准测试
//
// 专门用于测量各核心路径的 QPS（Queries Per Second），
// 从 ns/op 换算：QPS = 10^9 / ns/op。
//
// 运行方式：
//
//	go test -bench=BenchmarkQPS -benchmem -benchtime=500ms -run=^$ ./bench/suite/

import (
	"context"
	"fmt"
	"testing"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

// BenchmarkQPS_AgentRun 单 Agent 完整运行 QPS
func BenchmarkQPS_AgentRun(b *testing.B) {
	agent, err := synth.NewAgent("QPSAgent", "你是助手", &benchMockLLM{}, synth.WithMaxTurns(3))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = agent.Run(context.Background(), synth.UserMessage("hello"))
	}
}

// BenchmarkQPS_PoolDispatch AgentPool 调度 QPS（10 并发）
func BenchmarkQPS_PoolDispatch(b *testing.B) {
	pool := synth.NewPool(synth.PoolConfig{
		MaxConcurrency: 10,
		DefaultAgent: synth.ReActAgentConfig{
			SystemPrompt: "你是助手",
			MaxTurns:     1,
		},
	})
	defer pool.Close()
	pool.SetModel(&benchMockLLM{})

	tasks := make([]synth.TaskConfig, 10)
	for j := range tasks {
		tasks[j] = synth.TaskConfig{
			ID:     fmt.Sprintf("qps-%d", j),
			Title:  "QPS 基准",
			Prompt: "hello",
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = pool.Dispatch(context.Background(), tasks)
	}
}

// BenchmarkQPS_MemoryAdd 记忆写入 QPS
func BenchmarkQPS_MemoryAdd(b *testing.B) {
	memory, _ := synth.WithInMemory()
	defer memory.Close()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = memory.Add(ctx, &synth.Episode{
			ID:      fmt.Sprintf("qps-mem-%d", i),
			Content: "benchmark episode",
			Role:    "user",
		})
	}
}

// BenchmarkQPS_MemorySearch 记忆搜索 QPS
func BenchmarkQPS_MemorySearch(b *testing.B) {
	memory, _ := synth.WithInMemory()
	defer memory.Close()
	ctx := context.Background()

	// 预填充 1000 条
	for i := 0; i < 1000; i++ {
		_ = memory.Add(ctx, &synth.Episode{
			ID:      fmt.Sprintf("pre-%d", i),
			Content: fmt.Sprintf("episode content %d for search", i),
			Role:    "user",
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = memory.Search(ctx, "benchmark", nil)
	}
}

// BenchmarkQPS_ToolCall 工具调用端到端 QPS
func BenchmarkQPS_ToolCall(b *testing.B) {
	registry := synth.NewToolRegistry()
	fs, _ := synth.NewFileSystem(".")
	_ = registry.Register(fs)

	agent, err := synth.NewAgent("QPSToolAgent", "你是助手", &benchMockLLM{}, synth.WithMaxTurns(5), synth.WithToolkit(registry))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = agent.Run(context.Background(), synth.UserMessage("读取文件"))
	}
}

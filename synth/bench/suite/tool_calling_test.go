package suite

import (
	"context"
	"fmt"
	"testing"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

// BenchmarkToolCalling 基准：工具调用准确率
// 使用 MockLLM 模拟 LLM 返回工具调用请求，验证工具选择和执行正确性
func BenchmarkToolCalling(b *testing.B) {
	registry := synth.NewToolRegistry()
	fs, _ := synth.NewFileSystem(".")
	_ = registry.Register(fs)

	agent, err := synth.NewAgent("BenchAgent", "你是一个助手，使用工具完成任务。", &benchMockLLM{}, synth.WithMaxTurns(5), synth.WithToolkit(registry))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = agent.Run(context.Background(), synth.UserMessage("读取文件"))
	}
}

// BenchmarkAgentRun 基准：单次 Agent 运行吞吐量
func BenchmarkAgentRun(b *testing.B) {
	agent, err := synth.NewAgent("ThroughputAgent", "你是助手", &benchMockLLM{}, synth.WithMaxTurns(3))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = agent.Run(context.Background(), synth.UserMessage("hello"))
	}
}

// BenchmarkMemoryStore 基准：记忆存储写入和搜索
func BenchmarkMemoryStore(b *testing.B) {
	memory, _ := synth.WithInMemory()
	defer memory.Close()
	ctx := context.Background()

	b.Run("Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = memory.Add(ctx, &synth.Episode{
				ID:      fmt.Sprintf("bench-%d", i),
				Content: "benchmark test episode",
				Role:    "user",
			})
		}
	})

	b.Run("Search", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = memory.Search(ctx, "benchmark", nil)
		}
	})
}

// benchMockLLM 基准测试用 Mock LLM
type benchMockLLM struct{}

func (m *benchMockLLM) Complete(ctx context.Context, req *synth.CompletionRequest) (*synth.CompletionResponse, error) {
	return &synth.CompletionResponse{
		ID:      "bench-mock",
		Content: "模拟回复",
		Role:    "assistant",
		Usage:   synth.Usage{PromptTokens: 10, CompletionTokens: 10},
	}, nil
}

func (m *benchMockLLM) Stream(ctx context.Context, req *synth.CompletionRequest) (<-chan synth.Chunk, error) {
	ch := make(chan synth.Chunk, 1)
	go func() {
		defer close(ch)
		ch <- synth.Chunk{Content: "模拟回复", Done: true}
	}()
	return ch, nil
}

func (m *benchMockLLM) CallTools(ctx context.Context, req *synth.ToolCallRequest) (*synth.ToolCallResponse, error) {
	return &synth.ToolCallResponse{Usage: synth.Usage{}}, nil
}

func (m *benchMockLLM) Embeddings(ctx context.Context, texts []string) ([][]float32, error) {
	return make([][]float32, len(texts)), nil
}

func (m *benchMockLLM) Info() synth.ModelInfo {
	return synth.ModelInfo{Name: "bench-mock", Provider: "mock", MaxContext: 4096, SupportsTools: true}
}

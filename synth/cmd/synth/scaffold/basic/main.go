package main

import (
	"context"
	"fmt"
	"log"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

func main() {
	// 从环境变量读取 LLM 配置（SYNTH_LLM_API_KEY, SYNTH_LLM_MODEL 等）
	cfg := synth.ConfigFromEnv("")
	if cfg.APIKey == "" {
		log.Fatal("set SYNTH_LLM_API_KEY env var, e.g.: set SYNTH_LLM_API_KEY=sk-xxx")
	}

	provider, err := synth.NewOpenAIProvider(cfg)
	if err != nil {
		log.Fatalf("create provider failed: %v", err)
	}

	agent, err := synth.NewAgent("{{.ProjectName}}", "you are a helpful assistant.", provider, synth.WithMaxTurns(10))
	if err != nil {
		log.Fatalf("create agent failed: %v", err)
	}

	prompt := "Hello!"
	resp, err := agent.Run(context.Background(), synth.UserMessage(prompt))
	if err != nil {
		log.Fatalf("agent run failed: %v", err)
	}

	fmt.Printf("Reply: %s\n", resp.Content)
	fmt.Printf("Turns: %d\n", resp.Metrics.TotalTurns)
}

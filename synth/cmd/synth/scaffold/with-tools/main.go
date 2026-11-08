package main

import (
	"context"
	"fmt"
	"log"
	"os"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

func main() {
	// 从环境变量读取 LLM 配置
	cfg := synth.ConfigFromEnv("")
	if cfg.APIKey == "" {
		log.Fatal("set SYNTH_LLM_API_KEY env var, e.g.: set SYNTH_LLM_API_KEY=sk-xxx")
	}

	provider, err := synth.NewOpenAIProvider(cfg)
	if err != nil {
		log.Fatalf("create provider failed: %v", err)
	}

	// 配置工具集
	registry, err := synth.DefaultToolkit(synth.ToolkitConfig{
		RootDir:     ".",
		EnableFS:    true,
		EnableShell: true,
		EnableWeb:   true,
	})
	if err != nil {
		log.Fatalf("create toolkit failed: %v", err)
	}

	// 配置记忆存储
	memory, err := synth.WithInMemory()
	if err != nil {
		log.Fatalf("create memory store failed: %v", err)
	}
	defer memory.Close()

	agent, err := synth.NewAgent("{{.ProjectName}}", "you are an assistant that can read/write files, execute commands, and browse the web.",
		provider,
		synth.WithMaxTurns(20),
		synth.WithToolkit(registry),
		synth.WithMemory(memory),
	)
	if err != nil {
		log.Fatalf("create agent failed: %v", err)
	}

	prompt := "list files in the current directory"
	if envPrompt := os.Getenv("SYNTH_PROMPT"); envPrompt != "" {
		prompt = envPrompt
	}

	resp, err := agent.Run(context.Background(), synth.UserMessage(prompt))
	if err != nil {
		log.Fatalf("agent run failed: %v", err)
	}

	fmt.Printf("Reply: %s\n", resp.Content)
	fmt.Printf("Tool calls: %d\n", resp.Metrics.TotalTools)
}

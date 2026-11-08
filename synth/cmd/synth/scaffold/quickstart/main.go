package main

import (
	"context"
	"fmt"
	"log"
	"os"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

// Synth 快速入门示例
// 这是一个最小化的Agent示例，帮助你在5分钟内体验核心功能

func main() {
	fmt.Println("Synth 快速入门")
	fmt.Println("=========================")
	fmt.Println()

	// 步骤1: 配置LLM提供者
	apiKey := os.Getenv("SYNTH_LLM_API_KEY")

	var provider synth.Provider
	var err error

	if apiKey == "" {
		// 无 API key，使用 Demo 模式
		fmt.Println("提示: 未设置 SYNTH_LLM_API_KEY，使用 Demo 模式")
		fmt.Println("   Demo 模式基于关键词匹配产生响应，用于体验基本功能")
		fmt.Println("   设置 SYNTH_LLM_API_KEY 或运行 synth config set api-key 启用真实 LLM")
		fmt.Println()
		provider = synth.NewDemoProvider()
	} else {
		// 有 API key，使用真实 LLM（从环境变量读取配置）
		cfg := synth.ConfigFromEnv("")
		if cfg.Model == "" {
			cfg.Model = "gpt-4o-mini"
		}
		if cfg.BaseURL == "" {
			cfg.BaseURL = "https://api.openai.com/v1"
		}
		provider, err = synth.NewOpenAIProvider(cfg)
		if err != nil {
			log.Fatalf("创建LLM失败: %v", err)
		}
	}

	// 步骤2: 创建Agent（可选接入 OpenTelemetry 遥测）
	agentOpts := []synth.AgentOption{synth.WithMaxTurns(10)}

	// 设置 OTEL_EXPORTER_OTLP_ENDPOINT 或 SYNTH_OTEL_ENABLED=1 即启用：
	// ReAct 循环产生的 span 与指标将按 SYNTH_OTEL_EXPORT_INTERVAL（默认 30s）导出到 OTLP 端点。
	if tp, ok, terr := synth.TelemetryFromEnv(); terr == nil && ok {
		defer func() { _ = tp.Shutdown() }()
		agentOpts = append(agentOpts, synth.WithTelemetry(tp))
		fmt.Println("已启用 OpenTelemetry 遥测导出")
		fmt.Println()
	}

	myAgent, err := synth.NewAgent("QuickStartAgent", "你是一个友好的AI助手，用简洁的中文回答问题。", provider,
		agentOpts...,
	)
	if err != nil {
		log.Fatalf("创建Agent失败: %v", err)
	}

	fmt.Println("Agent已创建")
	fmt.Printf("   名称: %s\n", myAgent.Name())
	fmt.Println()

	// 步骤3: 运行Agent
	fmt.Println("开始对话...")
	fmt.Println()

	ctx := context.Background()

	userMessage := synth.UserMessage("你好！请用一句话介绍自己")
	response, err := myAgent.Run(ctx, userMessage)
	if err != nil {
		log.Fatalf("Agent运行失败: %v", err)
	}

	fmt.Printf("用户: %s\n", userMessage.Content)
	fmt.Printf("助手: %s\n", response.Content)
	fmt.Println()

	// 步骤4: 查看统计信息
	stats := myAgent.Stats()
	fmt.Println("运行统计:")
	fmt.Printf("   状态: %s\n", stats.Status)
	fmt.Printf("   当前轮数: %d\n", stats.CurrentTurn)
	fmt.Printf("   消息总数: %d\n", stats.TotalMessages)
	fmt.Printf("   工具调用: %v\n", stats.ToolsCalled)
	fmt.Println()

	fmt.Println("恭喜！你已成功运行第一个Agent")
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Println("   1. 尝试修改SystemPrompt，改变Agent的行为")
	fmt.Println("   2. 添加多轮对话，体验上下文记忆")
	fmt.Println("   3. 使用 --template with-tools 创建带工具的Agent")
	fmt.Println("   4. 运行 synth doctor 检查项目状态")
}

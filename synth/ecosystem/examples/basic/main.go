package main

import (
	"context"
	"fmt"
	"log"

	synth "github.com/Mirelth/Synth/v7/synth/pkg"
	"github.com/Mirelth/Synth/v7/synth/testutil"
)

func main() {
	fmt.Println("=== Synth: 基础示例 ===")
	fmt.Println()

	// 使用 PromptTemplate 构建系统提示词
	tmpl := synth.NewPromptTemplate("你是一个{{.Role}}助手，专注于{{.Domain}}领域。")
	tmpl.WithVar("Role", "Go开发").WithVar("Domain", "Agent框架")
	systemPrompt, err := tmpl.Render()
	if err != nil {
		log.Fatalf("渲染模板失败: %v", err)
	}
	fmt.Printf("系统提示词: %s\n\n", systemPrompt)

	mockLLM := testutil.NewMockProvider("你好！我是 Synth 助手，有什么可以帮你的？")

	agent, err := synth.NewAgent("BasicAgent", systemPrompt, mockLLM, synth.WithMaxTurns(3))
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	ctx := context.Background()

	// 检查点恢复：当被 `synth loop resume` 启动时，SYNTH_RESUME=1 会被注入
	if synth.ShouldResume() {
		agentID := synth.ResumeAgentID()
		if agentID == "" || agentID == "BasicAgent" {
			fmt.Printf("[info] 从检查点恢复 Agent %q...\n", agentID)
			if _, err := agent.ResumeFromCheckpoint(ctx); err != nil {
				log.Fatalf("恢复失败: %v", err)
			}
			return
		}
	}

	// 正常启动：使用 --prompt 参数或默认问候
	prompt := synth.ResumePrompt()
	if prompt == "" {
		prompt = "你好！"
	}

	resp, err := agent.Run(ctx, synth.UserMessage(prompt))
	if err != nil {
		log.Fatalf("Agent 运行失败: %v", err)
	}

	fmt.Printf("回复: %s\n", resp.Content)
	fmt.Printf("轮数: %d\n", resp.Metrics.TotalTurns)
	fmt.Printf("版本: %s\n", synth.Version)
}

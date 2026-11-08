package main

import (
	"context"
	"fmt"
	"log"

	synth "github.com/Mirelth/Synth/v7/synth/pkg"
	"github.com/Mirelth/Synth/v7/synth/testutil"
)

func main() {
	fmt.Println("=== 链式 API：最简 Agent ===")
	fmt.Println()

	mock := testutil.NewMockProvider("你好！我是链式 API 创建的 Agent，有什么可以帮你的？")

	// 只需 3 个必填参数即可创建 Agent
	agent, err := synth.NewAgent("hello-agent", "", mock, synth.WithMaxTurns(3))
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	resp, err := agent.Run(context.Background(), synth.UserMessage("你好！"))
	if err != nil {
		log.Fatalf("运行失败: %v", err)
	}

	fmt.Printf("回复: %s\n", resp.Content)
	fmt.Printf("轮数: %d\n", resp.Metrics.TotalTurns)
}

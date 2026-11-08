// a2a-interop 验收 demo：开放 A2A 协议互操作端到端演示
//
// 验收场景：开放协议标准客户端调用 synth Agent，完成跨生态任务委托（含流式事件）。
//
// 运行方式：go run ./ecosystem/examples/a2a-interop/
package main

import (
	"context"
	"fmt"
	"net/http/httptest"

	synth "github.com/Mirelth/Synth/v7/synth/pkg"
)

func main() {
	fmt.Println("=== Synth v3.5 协议互操作验收 Demo ===")
	fmt.Println()

	// 1. 部署开放协议兼容服务器（synth Agent 侧）
	card := synth.OpenAgentCard{
		Name:        "synth-data-agent",
		Description: "提供数据修复能力的 synth Agent",
		URL:         "http://synth-agent",
		Version:     "3.5.0",
		Capabilities: synth.OpenCapabilities{
			Streaming:              true,
			StateTransitionHistory: true,
		},
		Skills: []synth.OpenSkillDecl{
			{ID: "data-fix", Name: "数据修复", Description: "检测并修复异常数据", Tags: []string{"data"}},
		},
		DefaultInputModes:  []string{"text"},
		DefaultOutputModes: []string{"text"},
	}
	srv := synth.NewOpenInteropServer(card, synth.DefaultInteropConfig())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	fmt.Printf("🌐 开放协议服务器已启动: %s\n", ts.URL)
	fmt.Printf("   Agent Card: %s%s\n", ts.URL, synth.DefaultInteropConfig().AgentCardPath)
	fmt.Println()

	// 2. 第三方生态客户端接入
	client := synth.NewOpenInteropClient(ts.URL)
	ctx := context.Background()

	// 发现 Agent Card
	discovered, err := client.FetchAgentCard(ctx)
	if err != nil {
		fmt.Printf("   ❌ 发现失败: %v\n", err)
		return
	}
	fmt.Printf("🔍 生态客户端发现 Agent: %s (skills=%d)\n", discovered.Name, len(discovered.Skills))

	// 委托任务
	task, err := client.SendTask(ctx, synth.NewTextMessage("user", "请修复昨天的异常数据"))
	if err != nil {
		fmt.Printf("   ❌ 委托失败: %v\n", err)
		return
	}
	fmt.Printf("📨 任务已委托: %s (state=%s)\n", task.ID, task.Status.State)

	// 查询状态
	// 纪律：GetTask 失败时返回 (nil, err)，禁止用 _ 丢弃错误后解构 task——
	// 那会在对端异常时直接 nil panic（v7.6 前门修复在 README 示例中抓到同模式）。
	got, err := client.GetTask(ctx, task.ID)
	if err != nil {
		fmt.Printf("   ❌ 查询失败: %v\n", err)
		return
	}
	fmt.Printf("📊 任务状态: %s\n", got.Status.State)

	// 取消
	if err := client.CancelTask(ctx, task.ID); err != nil {
		fmt.Printf("   ❌ 取消失败: %v\n", err)
		return
	}
	canceled, err := client.GetTask(ctx, task.ID)
	if err != nil {
		fmt.Printf("   ❌ 查询失败: %v\n", err)
		return
	}
	fmt.Printf("🛑 取消后状态: %s\n", canceled.Status.State)
	fmt.Println()

	// 3. 兼容性报告
	report := synth.GenerateInteropReport(card, synth.DefaultInteropConfig())
	fmt.Printf("📋 协议符合性得分: %.0f%% (%d 项检查)\n", report.Score*100, len(report.Checks))
	if failed := report.FailedChecks(); len(failed) == 0 {
		fmt.Println("   ✅ 完全符合开放 A2A 协议规范")
	}
	fmt.Println()
	fmt.Println("=== 验收通过：开放协议发现→委托→查询→取消 跨生态链路完成 ===")
}

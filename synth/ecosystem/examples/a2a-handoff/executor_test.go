package main

import (
	"context"
	"strings"
	"testing"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

// TestDemoExecutorExecute Demo 模式执行器：委托任务后应完成并经 Status.Message
// 回传 DemoProvider 的能力清单响应（确定性关键词分支，含 A2A 主题内容）。
func TestDemoExecutorExecute(t *testing.T) {
	t.Setenv("SYNTH_A2A_HANDOFF_EXEC_DELAY_MS", "0") // 单测关闭模拟耗时
	exec, err := newDemoExecutor()
	if err != nil {
		t.Fatalf("newDemoExecutor: %v", err)
	}

	task := &synth.OpenTask{ID: "task-unit-1"}
	task.Messages = []synth.OpenMessage{synth.NewTextMessage("user", demoPrompt)}

	out, err := exec.Execute(context.Background(), task)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.ID != task.ID {
		t.Errorf("任务 ID 应保持 %q，实际 %q", task.ID, out.ID)
	}
	if out.Status.State != synth.OpenTaskCompleted {
		t.Fatalf("终态应为 %s，实际 %s", synth.OpenTaskCompleted, out.Status.State)
	}
	if out.Status.Message == nil {
		t.Fatal("结果应经 Status.Message 回传（pkg 未导出 OpenArtifact 构造面）")
	}
	result := out.Status.Message.TextContent()
	if !strings.Contains(result, "A2A 协议通信") {
		t.Errorf("Demo 响应应包含 A2A 主题内容，实际: %q", result)
	}
	if out.Metadata["llm_mode"] != "demo-keyword-matching" {
		t.Errorf("metadata llm_mode 应为 demo-keyword-matching，实际 %v", out.Metadata["llm_mode"])
	}

	// 原任务不得被执行器改写（锁内拷回由 server 负责）
	if task.Status.State == synth.OpenTaskCompleted {
		t.Error("Execute 不应就地改写入参 task 的状态")
	}
}

// TestDemoExecutorCancel 取消回调应可达且不报错。
func TestDemoExecutorCancel(t *testing.T) {
	exec, err := newDemoExecutor()
	if err != nil {
		t.Fatalf("newDemoExecutor: %v", err)
	}
	if err := exec.Cancel(context.Background(), "task-unit-2"); err != nil {
		t.Errorf("Cancel: %v", err)
	}
}

// TestUserTextOf 提取 user 消息文本；无 user 消息时返回空串。
func TestUserTextOf(t *testing.T) {
	task := &synth.OpenTask{ID: "t"}
	task.Messages = []synth.OpenMessage{
		synth.NewTextMessage("agent", "忽略我"),
		synth.NewTextMessage("user", "委托内容"),
	}
	if got := userTextOf(task); got != "委托内容" {
		t.Errorf("userTextOf = %q，期望 %q", got, "委托内容")
	}

	empty := &synth.OpenTask{ID: "t2"}
	empty.Messages = []synth.OpenMessage{synth.NewTextMessage("agent", "无用户消息")}
	if got := userTextOf(empty); got != "" {
		t.Errorf("无 user 消息应返回空串，实际 %q", got)
	}
}

// TestBuildCardPassesInteropChecks demo card 应 100% 通过协议符合性检查
// （与 run.sh / CI 门断言的 "8/8" 同一口径）。
func TestBuildCardPassesInteropChecks(t *testing.T) {
	card := buildCard("127.0.0.1:9797")
	if card.Version != synth.Version {
		t.Errorf("card 版本应为框架真值 %q，实际 %q", synth.Version, card.Version)
	}
	report := synth.GenerateInteropReport(card, synth.DefaultInteropConfig())
	if len(report.Checks) != 8 {
		t.Fatalf("检查项应为 8 个，实际 %d", len(report.Checks))
	}
	if failed := report.FailedChecks(); len(failed) != 0 {
		t.Errorf("应符合性 8/8 全过，未过项: %v", failed)
	}
	if report.Score != 1.0 {
		t.Errorf("得分应为 1.0，实际 %v", report.Score)
	}
}

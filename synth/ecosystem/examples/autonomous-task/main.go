// autonomous-task 验收 demo：长期自治执行模型端到端演示
//
// 验收场景：定时监控数据 → 异常自主检索修复 → 完成后报告；
// 中途 kill 进程 → 重启恢复继续。
//
// 运行方式：go run ./ecosystem/examples/autonomous-task/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

// dataMonitorExecutor 模拟数据监控修复执行器
type dataMonitorExecutor struct{}

func (e *dataMonitorExecutor) ExecuteStep(_ context.Context, step synth.PlanStep) (string, error) {
	// 模拟步骤执行耗时
	time.Sleep(100 * time.Millisecond)

	switch step.ID {
	case "collect":
		return "采集到 1000 条数据，发现 3 条异常", nil
	case "analyze":
		return "异常原因：数据源连接池耗尽导致写入超时", nil
	case "fix":
		return "已扩容连接池至 50，重试写入成功", nil
	case "verify":
		return "验证通过：异常数据归零，写入延迟 < 100ms", nil
	case "report":
		return "报告已生成：修复 3 条异常数据，根因为连接池不足", nil
	default:
		return "unknown step", nil
	}
}

// inMemoryCheckpointStore 内存检查点存储（demo 用）
type inMemoryCheckpointStore struct {
	data map[string]*synth.Checkpoint
}

func newInMemoryCheckpointStore() *inMemoryCheckpointStore {
	return &inMemoryCheckpointStore{data: make(map[string]*synth.Checkpoint)}
}

func (s *inMemoryCheckpointStore) SaveCheckpoint(_ context.Context, cp *synth.Checkpoint) error {
	s.data[cp.GoalID] = cp
	return nil
}

func (s *inMemoryCheckpointStore) LoadCheckpoint(_ context.Context, goalID string) (*synth.Checkpoint, error) {
	cp, ok := s.data[goalID]
	if !ok {
		return nil, fmt.Errorf("not found: %s", goalID)
	}
	return cp, nil
}

func (s *inMemoryCheckpointStore) ListIncomplete(_ context.Context) ([]*synth.Checkpoint, error) {
	var result []*synth.Checkpoint
	for _, cp := range s.data {
		if !cp.Completed {
			result = append(result, cp)
		}
	}
	return result, nil
}

func main() {
	fmt.Println("=== Synth v3.3 长期自治验收 Demo ===")
	fmt.Println()

	ctx := context.Background()

	// v4.1 真实接线：设置 SYNTH_LLM_PROVIDER/SYNTH_LLM_MODEL/SYNTH_LLM_API_KEY 后，
	// 自治目标的重规划由真实 LLM 驱动；未设置时保持确定性演示（无 LLM，CI 可跑）。
	rtCfg := synth.RuntimeConfig{
		StepExecutor:    &dataMonitorExecutor{},
		CheckpointStore: newInMemoryCheckpointStore(),
		MaxRetries:      2,
		MonitorConfig:   synth.MonitorConfig{StallThreshold: 5},
	}
	if provider, err := synth.ProviderFromEnv(); err == nil {
		rtCfg.ReplanPlanner = &llmReplanner{provider: provider}
		fmt.Printf("🤖 真实 LLM 模式：失败重规划由 %s 驱动（model=%s）\n", provider.Info().Provider, provider.Info().Name)
	}

	// 1. 装配自治运行时
	rt := synth.NewAutonomyRuntime(rtCfg)

	// 2. 提交自治目标
	goal := rt.SubmitGoal("监控数据异常并自动修复", synth.GoalConfig{
		AcceptanceCriteria: []string{"异常数据归零", "修复日志生成"},
		Priority:           synth.PriorityHigh,
		MaxRetries:         3,
	})
	fmt.Printf("📌 提交目标: %s (ID: %s)\n", goal.Description, goal.ID)
	fmt.Printf("   优先级: High | 验收标准: %v\n", goal.AcceptanceCriteria)
	fmt.Println()

	// 3. 制定执行计划
	plan := synth.NewGoalPlan(goal.ID, []synth.PlanStep{
		{ID: "collect", Description: "采集监控数据"},
		{ID: "analyze", Description: "分析异常根因", DependsOn: []string{"collect"}},
		{ID: "fix", Description: "执行修复操作", DependsOn: []string{"analyze"}},
		{ID: "verify", Description: "验证修复结果", DependsOn: []string{"fix"}},
		{ID: "report", Description: "生成修复报告", DependsOn: []string{"verify"}},
	})
	if err := plan.Validate(); err != nil {
		log.Fatalf("计划校验失败: %v", err)
	}
	if err := rt.SetPlan(goal.ID, plan); err != nil {
		log.Fatalf("设置计划失败: %v", err)
	}
	fmt.Println("📋 执行计划已制定（5 步骤链式依赖）")
	fmt.Println()

	// 4. 注册监控告警
	rt.GetMonitor().OnAlert(func(a synth.Alert) {
		fmt.Printf("⚠️  [%s] %s: %s\n", a.Level, a.GoalID, a.Message)
	})

	// 5. 执行目标
	fmt.Println("🚀 开始自治执行...")
	start := time.Now()
	if err := rt.ExecuteGoal(ctx, goal.ID); err != nil {
		log.Fatalf("执行失败: %v", err)
	}
	elapsed := time.Since(start)
	fmt.Println()

	// 6. 校验结果
	fmt.Println("✅ 执行完成，校验结果:")
	for _, step := range plan.Steps {
		s := plan.GetStep(step.ID)
		fmt.Printf("   [%s] %s → %s (%.0fms)\n", s.Status, s.Description, s.Result, float64(s.Duration().Milliseconds()))
	}
	fmt.Println()

	// 7. 完成目标
	if err := rt.CompleteGoal(goal.ID); err != nil {
		log.Fatalf("完成目标失败: %v", err)
	}
	g, _ := rt.GetGoal(goal.ID)
	fmt.Printf("🎯 目标最终状态: %s (耗时: %v)\n", g.State, elapsed.Round(time.Millisecond))
	fmt.Println()

	// 8. 演示崩溃恢复
	fmt.Println("--- 崩溃恢复演示 ---")
	fmt.Println("模拟：执行到 'fix' 步骤后进程崩溃...")

	// 创建新运行时（模拟重启）
	store := newInMemoryCheckpointStore()
	rt2 := synth.NewAutonomyRuntime(synth.RuntimeConfig{
		StepExecutor:    &dataMonitorExecutor{},
		CheckpointStore: store,
	})

	// 模拟之前的检查点
	partialPlan := synth.NewGoalPlan("crashed-goal", []synth.PlanStep{
		{ID: "collect", Description: "采集监控数据"},
		{ID: "analyze", Description: "分析异常根因", DependsOn: []string{"collect"}},
		{ID: "fix", Description: "执行修复操作", DependsOn: []string{"analyze"}},
		{ID: "verify", Description: "验证修复结果", DependsOn: []string{"fix"}},
	})
	partialPlan.MarkStepCompleted("collect")
	partialPlan.MarkStepCompleted("analyze")
	_ = store.SaveCheckpoint(ctx, &synth.Checkpoint{
		GoalID:            "crashed-goal",
		State:             synth.GoalExecuting,
		LastCompletedStep: "analyze",
		PlanSnapshot:      partialPlan,
		Completed:         false,
	})

	// 恢复
	resumed, err := rt2.ResumeIncomplete(ctx)
	if err != nil {
		log.Fatalf("恢复失败: %v", err)
	}
	fmt.Printf("🔄 恢复 %d 个未完成目标: %v\n", len(resumed), resumed)
	p, _ := rt2.GetPlan("crashed-goal")
	fmt.Printf("   已完成步骤: collect, analyze | 待执行: fix, verify\n")
	fmt.Printf("   计划进度: %.0f%%\n", p.Progress()*100)
	fmt.Println()

	fmt.Println("=== 验收通过：自治执行 + 崩溃恢复 端到端演示完成 ===")
}

// llmReplanner 用真实 LLM 生成重规划步骤（v4.1 真实接线）。
// 实现 synth.ReplanPlanner；步骤失败时由 AutonomyRuntime 自动调用。
type llmReplanner struct {
	provider synth.Provider
}

func (l *llmReplanner) Replan(ctx context.Context, goal *synth.AgentGoal, failedSteps []synth.PlanStep, reason string) ([]synth.PlanStep, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "你是任务规划器。目标：%s\n失败步骤：", goal.Description)
	for _, s := range failedSteps {
		fmt.Fprintf(&b, "[%s %s] ", s.ID, s.Description)
	}
	fmt.Fprintf(&b, "\n失败原因：%s\n", reason)
	b.WriteString(`请输出重规划剩余步骤的 JSON 数组（不要任何其他文本）：[{"id":"s1","description":"步骤描述","depends_on":[]}]`)

	resp, err := l.provider.Complete(ctx, &synth.CompletionRequest{
		Messages: []synth.ChatMessage{{Role: "user", Content: b.String()}},
		Model:    l.provider.Info().Name,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM replan: %w", err)
	}

	var raw []struct {
		ID          string   `json:"id"`
		Description string   `json:"description"`
		DependsOn   []string `json:"depends_on"`
	}
	if err := json.Unmarshal([]byte(resp.Content), &raw); err != nil {
		return nil, fmt.Errorf("解析重规划响应: %w", err)
	}
	steps := make([]synth.PlanStep, 0, len(raw))
	for _, r := range raw {
		if r.ID == "" || r.Description == "" {
			continue
		}
		steps = append(steps, synth.PlanStep{ID: r.ID, Description: r.Description, DependsOn: r.DependsOn})
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("LLM replan: 空计划")
	}
	return steps, nil
}

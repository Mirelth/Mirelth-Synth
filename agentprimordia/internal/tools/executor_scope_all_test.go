package tools

import (
	"context"
	"strings"
	"testing"
)

// 2026-10-26 P1 回归：extractPathsFromArgs 旧实现最多返回两个路径，
// 第三个及后续路径不会经过 ScopePolicy。这里用三个路径字段，
// 第三个落在允许范围外，必须在工具执行前拒绝。
func TestExecutor_ScopePolicyChecksAllPaths(t *testing.T) {
	reg := NewRegistry()
	tool := &mockTool{name: "multi_path", response: "executed"}
	if err := reg.Register(tool); err != nil {
		t.Fatalf("register: %v", err)
	}

	policy := NewFileScopePolicy()
	policy.SetScope("agent-1", []string{"/allowed"})

	executor := NewExecutor(reg).WithScopePolicy(policy, "agent-1")
	result, err := executor.Execute(context.Background(), &FunctionCall{
		ID:   "call_1",
		Name: "multi_path",
		Args: `{"src":"/allowed/a","dst":"/allowed/b","output_path":"/outside/c"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "denied access") {
		t.Fatalf("expected scope denial for third path, got %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatalf("expected error result, got %+v", result)
	}
}

func TestExtractAllPathsFromArgs_DeterministicOrder(t *testing.T) {
	args := `{"path":"/z","file":"/a","nested":{"output_path":"/n"}}`
	first := extractAllPathsFromArgs(args)
	second := extractAllPathsFromArgs(args)
	if len(first) != 3 {
		t.Fatalf("expected 3 paths, got %v", first)
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("non-deterministic order: %v vs %v", first, second)
		}
	}
}

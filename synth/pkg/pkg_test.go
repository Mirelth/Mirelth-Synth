package synth_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Mirelth/Synth/v7/synth/internal/agent"
	synth "github.com/Mirelth/Synth/v7/synth/pkg"
)

func TestVersion(t *testing.T) {
	t.Parallel()
	if synth.Version == "" {
		t.Fatal("Version should not be empty")
	}
}

func TestErrorSentinels(t *testing.T) {
	t.Parallel()
	sentinelTests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrAgentStopped", synth.ErrAgentStopped, "agent is stopped"},
		{"ErrMaxTurnsExceeded", synth.ErrMaxTurnsExceeded, "max turns exceeded"},
		{"ErrInvalidConfig", synth.ErrInvalidConfig, "invalid configuration"},
		{"ErrTimeout", synth.ErrTimeout, "operation timed out"},
		{"ErrTaskNotFound", synth.ErrTaskNotFound, "task not found"},
		{"ErrAgentRunning", synth.ErrAgentRunning, "agent is already running"},
		{"ErrToolNotFound", synth.ErrToolNotFound, "tool not found"},
		{"ErrToolExecution", synth.ErrToolExecution, "tool execution failed"},
		{"ErrLLMCallFailed", synth.ErrLLMCallFailed, "LLM call failed"},
		{"ErrContextCanceled", synth.ErrContextCanceled, "context canceled"},
		{"ErrPoolFull", synth.ErrPoolFull, "pool is at max capacity"},
	}

	for _, tt := range sentinelTests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, errors.New(tt.msg)) {
				if tt.err.Error() != tt.msg {
					t.Errorf("expected error message %q, got %q", tt.msg, tt.err.Error())
				}
			}
		})
	}
}

func TestAgentTypes(t *testing.T) {
	t.Parallel()
	msg := synth.UserMessage("hello")
	if msg.Role != synth.RoleUser {
		t.Errorf("expected RoleUser, got %s", msg.Role)
	}

	sysMsg := synth.SystemMessage("you are a helper")
	if sysMsg.Role != synth.RoleSystem {
		t.Errorf("expected RoleSystem, got %s", sysMsg.Role)
	}
}

func TestHookManager(t *testing.T) {
	t.Parallel()
	hm := synth.NewHookManager()
	if hm == nil {
		t.Fatal("NewHookManager returned nil")
	}
	hm.Register(synth.HookBeforeRun, func(ctx context.Context, hctx *synth.HookContext) error {
		return nil
	})
	if hm.Count(synth.HookBeforeRun) != 1 {
		t.Errorf("expected 1 hook, got %d", hm.Count(synth.HookBeforeRun))
	}
}

func TestLifecycle(t *testing.T) {
	t.Parallel()
	lc := synth.NewLifecycle()
	if lc == nil {
		t.Fatal("NewLifecycle returned nil")
	}
	if lc.Status() != synth.StatusIdle {
		t.Errorf("expected StatusIdle, got %s", lc.Status())
	}
}

func TestToolRegistry(t *testing.T) {
	t.Parallel()
	reg := synth.NewToolRegistry()
	if reg == nil {
		t.Fatal("NewToolRegistry returned nil")
	}
}

func TestMemoryStore(t *testing.T) {
	t.Parallel()
	store, err := synth.WithInMemory()
	if err != nil {
		t.Fatalf("WithInMemory failed: %v", err)
	}
	if store == nil {
		t.Fatal("WithInMemory returned nil")
	}
	store.Close()
}

func TestVectorStore(t *testing.T) {
	t.Parallel()
	vs := synth.NewVectorStore(3)
	if vs == nil {
		t.Fatal("NewVectorStore returned nil")
	}
	if vs.Dimensions() != 3 {
		t.Errorf("expected 3 dimensions, got %d", vs.Dimensions())
	}
}

func TestEventBus(t *testing.T) {
	t.Parallel()
	bus := synth.NewBus(16)
	if bus == nil {
		t.Fatal("NewBus returned nil")
	}
	bus.Close()
}

func TestMetrics(t *testing.T) {
	t.Parallel()
	m := synth.NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}
}

func TestSecurity(t *testing.T) {
	t.Parallel()
	acl := synth.NewACL()
	if acl == nil {
		t.Fatal("NewACL returned nil")
	}
	sb := synth.NewSandbox(acl)
	if sb == nil {
		t.Fatal("NewSandbox returned nil")
	}
}

func TestCheckpointStore(t *testing.T) {
	t.Parallel()
	store, err := synth.InMemoryCheckpointStore()
	if err != nil {
		t.Fatalf("InMemoryCheckpointStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("InMemoryCheckpointStore returned nil")
	}
	store.Close()
}

func TestFileLockManager(t *testing.T) {
	t.Parallel()
	flm := synth.NewFileLockManager()
	if flm == nil {
		t.Fatal("NewFileLockManager returned nil")
	}
}

func TestContextWindowStrategy(t *testing.T) {
	t.Parallel()
	strategy := synth.NewDefaultStrategy(5)
	if strategy == nil {
		t.Fatal("NewDefaultStrategy returned nil")
	}
}

func TestResilientConfig(t *testing.T) {
	t.Parallel()
	cfg := synth.DefaultResilientConfig()
	if cfg.MaxRetries <= 0 {
		t.Error("DefaultResilientConfig MaxRetries should be > 0")
	}
}

func TestFileScopePolicy(t *testing.T) {
	t.Parallel()
	policy := synth.NewFileScopePolicy()
	if policy == nil {
		t.Fatal("NewFileScopePolicy returned nil")
	}
}

func TestBuiltinTools(t *testing.T) {
	t.Parallel()
	fs, err := synth.NewFileSystem(".")
	if err != nil {
		t.Fatalf("NewFileSystem error: %v", err)
	}
	if fs == nil {
		t.Fatal("NewFileSystem returned nil")
	}
	sh := synth.NewShell()
	if sh == nil {
		t.Fatal("NewShell returned nil")
	}
	web := synth.NewWeb()
	if web == nil {
		t.Fatal("NewWeb returned nil")
	}
}

func TestErrorCodeMapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err  error
		code string
	}{
		{synth.ErrAgentStopped, "AGENT_001"},
		{synth.ErrAgentRunning, "AGENT_002"},
		{synth.ErrMaxTurnsExceeded, "AGENT_003"},
		{synth.ErrNoToolkit, "AGENT_004"},
		{synth.ErrToolNotFound, "TOOL_001"},
		{synth.ErrToolExecution, "TOOL_002"},
		{synth.ErrInvalidConfig, "TOOL_003"},
		{synth.ErrConfirmDenied, "TOOL_004"},
		{synth.ErrLLMCallFailed, "LLM_001"},
		{synth.ErrNotSupported, "LLM_002"},
		{synth.ErrCircuitOpen, "LLM_003"},
		{synth.ErrAPIKeyRequired, "LLM_004"},
		{synth.ErrEmptyResponse, "LLM_005"},
		{synth.ErrResponseParseFailed, "LLM_006"},
		{synth.ErrRetriesExhausted, "LLM_007"},
		{synth.ErrFallbackFailed, "LLM_008"},
		{synth.ErrPoolFull, "POOL_001"},
		{synth.ErrTaskNotFound, "POOL_002"},
		{synth.ErrTimeout, "POOL_003"},
		{synth.ErrContextCanceled, "CTX_001"},
		{synth.ErrEpisodeNotFound, "MEM_001"},
		{synth.ErrInvalidImportance, "MEM_002"},
		{synth.ErrEmptyEpisodeID, "MEM_003"},
		{synth.ErrEmptySessionID, "MEM_004"},
		{synth.ErrEmptyRole, "MEM_005"},
		{synth.ErrEmptyContent, "MEM_006"},
		{synth.ErrDimensionMismatch, "MEM_007"},
		{synth.ErrVectorNotFound, "MEM_008"},
		{synth.ErrCommandBlocked, "SEC_001"},
		{synth.ErrCommandNotAllowed, "SEC_002"},
		{synth.ErrAccessDenied, "SEC_003"},
		{synth.ErrPathTraversal, "SEC_004"},
		{synth.ErrBusClosed, "EVT_001"},
		{synth.ErrCheckpointNotFound, "PST_001"},
		{synth.ErrGlobalWriteConflict, "CON_001"},
		{synth.ErrScopeOverlap, "CON_002"},
	}

	for _, tt := range tests {
		code := synth.GetErrorCode(tt.err)
		if code != tt.code {
			t.Errorf("GetErrorCode(%v) = %q, want %q", tt.err, code, tt.code)
		}
	}
}

func TestErrorCodeUnknown(t *testing.T) {
	t.Parallel()
	code := synth.GetErrorCode(errors.New("some random error"))
	if code != "UNKNOWN" {
		t.Errorf("GetErrorCode(unknown) = %q, want UNKNOWN", code)
	}
}

func TestCodeError(t *testing.T) {
	t.Parallel()
	ce := synth.WithCode("CUSTOM_001", "custom error")
	if ce.Code != "CUSTOM_001" {
		t.Errorf("Code = %q, want CUSTOM_001", ce.Code)
	}
	if ce.Error() != "custom error" {
		t.Errorf("Error() = %q, want 'custom error'", ce.Error())
	}
	if ce.GetCode() != "CUSTOM_001" {
		t.Errorf("GetCode() = %q, want CUSTOM_001", ce.GetCode())
	}
	// 修复评估报告 §四.1-①：GetErrorCode 的 errors.As 路径必须命中 CodeError
	//（修复前 WithCode 构造的错误码返回 "UNKNOWN"）。
	if code := synth.GetErrorCode(ce); code != "CUSTOM_001" {
		t.Errorf("GetErrorCode(CodeError) = %q, want CUSTOM_001", code)
	}
	// %w 包装后同样可命中
	wrapped := fmt.Errorf("outer: %w", ce)
	if code := synth.GetErrorCode(wrapped); code != "CUSTOM_001" {
		t.Errorf("GetErrorCode(wrapped CodeError) = %q, want CUSTOM_001", code)
	}
}

// TestErrOutputBlockedIdentity 验证 pkg.ErrOutputBlocked 与 agent 护栏拦截
// 抛出的 sentinel 是同一错误值（修复评估报告 §四.1-③）。
func TestErrOutputBlockedIdentity(t *testing.T) {
	t.Parallel()
	if synth.ErrOutputBlocked != agent.ErrOutputBlocked {
		t.Fatal("synth.ErrOutputBlocked 与 agent.ErrOutputBlocked 不是同一 sentinel")
	}
	if !errors.Is(agent.ErrOutputBlocked, synth.ErrOutputBlocked) {
		t.Fatal("errors.Is(agent.ErrOutputBlocked, synth.ErrOutputBlocked) 应为 true")
	}
	if code := synth.GetErrorCode(agent.ErrOutputBlocked); code != "GRD_002" {
		t.Fatalf("GetErrorCode(ErrOutputBlocked) = %q, want GRD_002", code)
	}
	if !synth.IsAccessDenied(agent.ErrOutputBlocked) {
		t.Fatal("IsAccessDenied(ErrOutputBlocked) 应为 true")
	}
}

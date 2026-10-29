package persist

import (
	"context"
	"testing"
	"time"
)

// 2026-10-26 P1 回归：重复 startHeartbeat 不能倍增心跳协程。
func TestStartHeartbeat_DoesNotDuplicate(t *testing.T) {
	le := NewLeaderElector(NewMemoryCoordinator("n1", time.Second), "n1", LeaderConfig{
		ElectionKey:       "leader:test",
		HeartbeatInterval: 10 * time.Millisecond,
		TTL:               time.Second,
		RetryInterval:     10 * time.Millisecond,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	le.startHeartbeat(ctx)
	le.startHeartbeat(ctx) // 旧实现会再起一个 ticker
	time.Sleep(55 * time.Millisecond)

	// 单协程约 5 次；双协程约 10 次。留出调度抖动余量。
	if got := le.heartbeatsSent.Load(); got > 8 {
		t.Fatalf("heartbeat goroutine duplicated: %d beats in 55ms", got)
	}
	le.Stop()
}

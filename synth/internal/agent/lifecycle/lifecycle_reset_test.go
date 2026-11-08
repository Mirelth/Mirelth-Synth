package lifecycle

import (
	"context"
	"testing"
	"time"
)

// 2026-10-26 P1 回归：Reset 后 runningDone/gracefulCh 必须是新通道，
// 否则第二次 Run 的 GracefulShutdown 会立即返回（假停机）。
func TestReset_RecreatesRunLifecycleChannels(t *testing.T) {
	lc := New()
	_ = lc.SetStatus(StatusRunning)
	_ = lc.SetStatus(StatusCompleted)
	if err := lc.Reset(); err != nil {
		t.Fatal(err)
	}
	_ = lc.SetStatus(StatusRunning)

	done := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		done <- lc.GracefulShutdown(ctx)
	}()

	select {
	case err := <-done:
		t.Fatalf("GracefulShutdown returned before the run completed: %v", err)
	case <-time.After(50 * time.Millisecond):
		// expected: it must wait for runningDone
	}

	_ = lc.SetStatus(StatusCompleted)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("GracefulShutdown after completion: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("GracefulShutdown did not return after completion")
	}
}

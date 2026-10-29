package agent

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

// 2026-10-26 P0 回归：消费者不排空 channel 时 emitStream 不能永久阻塞。
func TestEmitStream_DoesNotBlockWhenConsumerStops(t *testing.T) {
	a := &ReActAgent{logger: slog.Default()}
	ch := make(chan StreamEvent) // 无缓冲且无人读取
	cfg := loopConfig{stream: true, streamCh: ch, streamCtx: context.Background(), requestID: "req-1"}

	done := make(chan struct{})
	go func() {
		a.emitStream(cfg, StreamEvent{Type: StreamEventToken, Content: "hello"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("emitStream blocked on a non-draining consumer")
	}
}

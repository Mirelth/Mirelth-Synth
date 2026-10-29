package llm

import (
	"context"
	"time"
)

// streamSendTimeout 是单个流式 chunk 等待消费者读取的上限。
// 旧实现只 select ch<-chunk / ctx.Done()，消费者停止读取且未 cancel 时
// 缓冲满后 goroutine 与 resp.Body 会永久挂住；这里超时后直接结束流。
const streamSendTimeout = 5 * time.Second

// sendChunk 把 chunk 发送到流通道，返回 false 表示流应终止。
//
// 超时意味着消费者在 streamSendTimeout 内没有读取。此时仅关闭通道会让
// 上层把已收到的半截内容误当成完整结果，所以先尽力投递一个带 Done 的
// 错误帧；若缓冲仍满（消费者完全停滞）则投递失败，但至少不会阻塞。
func sendChunk(ctx context.Context, ch chan<- Chunk, chunk Chunk) bool {
	timer := time.NewTimer(streamSendTimeout)
	defer timer.Stop()
	select {
	case ch <- chunk:
		return true
	case <-ctx.Done():
		return false
	case <-timer.C:
		select {
		case ch <- Chunk{Error: "stream send timed out: consumer stalled for " + streamSendTimeout.String(), Done: true}:
		default:
		}
		return false
	}
}

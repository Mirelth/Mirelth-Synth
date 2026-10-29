package llm

import (
	"strings"
	"testing"
)

func TestStreamCollector_PropagatesChunkError(t *testing.T) {
	ch := make(chan Chunk, 1)
	ch <- Chunk{Error: "429 rate limited", Done: true}
	close(ch)

	_, err := NewStreamCollector().Collect(ch)
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected stream error, got %v", err)
	}
}

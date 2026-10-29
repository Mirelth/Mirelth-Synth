package llm

import (
	"context"
	"testing"
)

func TestFingerprintCache_NilResponseIsNotStored(t *testing.T) {
	c := NewFingerprintCache(8, 0)
	if err := c.Set(context.Background(), "q", nil); err != nil {
		t.Fatalf("Set(nil) returned error: %v", err)
	}
	if _, hit := c.Get(context.Background(), "q", 0); hit {
		t.Fatal("nil response must not produce a cache hit")
	}
}

func TestFingerprintCache_SetRequestNilIsSafe(t *testing.T) {
	c := NewFingerprintCache(8, 0)
	req := &CompletionRequest{Messages: []ChatMessage{{Role: "user", Content: "q"}}}
	if err := c.SetRequest(context.Background(), req, nil); err != nil {
		t.Fatalf("SetRequest(nil) returned error: %v", err)
	}
	if _, hit := c.GetRequest(context.Background(), req); hit {
		t.Fatal("nil response must not produce a cache hit")
	}
}

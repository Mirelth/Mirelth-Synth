package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/Mirelth/Mirelth-Synth/v7/synth/internal/llm"
)

func TestAddLLMUsage(t *testing.T) {
	got := addLLMUsage(llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}, llm.Usage{PromptTokens: 2, CompletionTokens: 3, TotalTokens: 5})
	if got.PromptTokens != 12 || got.CompletionTokens != 8 || got.TotalTokens != 20 {
		t.Fatalf("unexpected merged usage: %+v", got)
	}
}

func TestIsRetryableToolError(t *testing.T) {
	if !isRetryableToolError(context.DeadlineExceeded) {
		t.Fatal("deadline exceeded should be retryable")
	}
	if isRetryableToolError(errors.New("invalid parameter")) {
		t.Fatal("validation error must not be retryable")
	}
	if isRetryableToolError(context.Canceled) {
		t.Fatal("canceled must not be retryable")
	}
}

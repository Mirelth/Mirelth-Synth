package main

import (
	"context"
	"testing"
	"time"

	"github.com/Mirelth/Synth/v7/synth/cmd/example/demo"
	synth "github.com/Mirelth/Synth/v7/synth/pkg"
)

// TestSimpleExample 验证最简示例的核心逻辑不会腐烂
func TestSimpleExample(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	demoLLM := demo.NewDemoLLM("你好!我是AI助手!")
	a, err := synth.NewAgent("TestBot", "你是一个友好的AI助手。", demoLLM, synth.WithMaxTurns(3))
	if err != nil {
		t.Fatalf("NewAgent error: %v", err)
	}

	resp, err := a.Run(ctx, synth.UserMessage("你好"))
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response")
	}

	stats := a.Stats()
	if stats.Status == "" {
		t.Error("expected non-empty status")
	}
}

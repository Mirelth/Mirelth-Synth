package mcp

import (
	"context"
	"strings"
	"testing"
)

func TestMCPServer_RefusesNonLoopbackWithoutAPIKey(t *testing.T) {
	s := NewMCPServer()
	err := s.Start(context.Background(), ":0")
	if err == nil || !strings.Contains(err.Error(), "non-loopback") {
		t.Fatalf("expected refusal for non-loopback without API key, got %v", err)
	}
}

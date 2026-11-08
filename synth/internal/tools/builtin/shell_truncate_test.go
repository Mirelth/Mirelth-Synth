package builtin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestShell_OutputTruncation(t *testing.T) {
	dir := t.TempDir()
	shell := NewShell().WithBlacklist().WithAllowedWorkdirs([]string{dir})
	shell.maxOutputSize = 10

	args, _ := json.Marshal(map[string]string{
		"action":  "execute",
		"command": "go version",
		"workdir": dir,
	})

	result, err := shell.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "截断") {
		t.Errorf("expected truncation marker in output, got: %s", result.Content)
	}
}

func TestShell_OutputUnderLimit(t *testing.T) {
	dir := t.TempDir()
	shell := NewShell().WithBlacklist().WithAllowedWorkdirs([]string{dir})

	args, _ := json.Marshal(map[string]string{
		"action":  "execute",
		"command": "go env GOROOT",
		"workdir": dir,
	})

	result, err := shell.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", result.Content)
	}
	if strings.Contains(result.Content, "截断") {
		t.Error("short output should not be truncated")
	}
}

package builtin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func execFS(t *testing.T, fs *FileSystem, params map[string]any) *toolResultView {
	t.Helper()
	raw, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	res, err := fs.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	return &toolResultView{content: res.Content, isError: res.IsError}
}

type toolResultView struct {
	content string
	isError bool
}

// 2026-10-26 P1 回归：search_directory 逐文件 os.ReadFile 会跟随符号链接，
// root 内指向窗外的链接可绕过 openAndValidate 的 symlink 复核。
func TestSearchDirectory_SkipsSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("TOP_SECRET_TOKEN"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "looks-innocent.txt")
	if err := os.Symlink(secret, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "normal.txt"), []byte("normal"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := mustNewFS(t, root)
	res := execFS(t, fs, map[string]any{
		"action": "search_directory",
		"path":   ".",
		"query":  "TOP_SECRET_TOKEN",
	})
	if res.isError {
		t.Fatalf("search failed: %s", res.content)
	}
	if strings.Contains(res.content, "secret.txt") || strings.Contains(res.content, "TOP_SECRET_TOKEN:") {
		t.Fatalf("search_directory followed a symlink outside root: %s", res.content)
	}
}

// 2026-10-26 P2 回归：editFile 必须与 readFile 一样有读取上限。
func TestEditFile_RejectsOversizedFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "large.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 64)), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := mustNewFS(t, root)
	fs.maxReadSize = 16
	res := execFS(t, fs, map[string]any{
		"action":  "edit",
		"path":    "large.txt",
		"old_str": "x",
		"new_str": "y",
	})
	if !res.isError || !strings.Contains(res.content, "too large") {
		t.Fatalf("oversized edit was not rejected: %s", res.content)
	}
}

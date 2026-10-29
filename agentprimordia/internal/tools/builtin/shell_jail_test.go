package builtin

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// 2026-10-26 P0 回归：cp/mv/sort 会把 --opt=/abs 的值当绝对路径，
// 旧实现只检查 token 是否以 '/' 开头，导致 jail 被绕过。
func TestShellJail_RejectsOptionEmbeddedAbsolutePath(t *testing.T) {
	if _, err := exec.LookPath("cp"); err != nil {
		t.Skip("cp not installed")
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	sh := NewShell().WithWhitelist([]string{"cp"}).WithAllowedWorkdirs([]string{root})
	res := execShellResult(t, sh, map[string]any{
		"action":  "execute",
		"command": "cp",
		"args":    []string{"source.txt", "--target-directory=" + outside},
	})
	if !res.IsError || !strings.Contains(res.Content, "outside the allowed directories") {
		t.Fatalf("option-embedded absolute path was not rejected: %s", res.Content)
	}
	if _, err := os.Stat(filepath.Join(outside, "source.txt")); err == nil {
		t.Fatal("cp wrote outside the allowed workdir")
	}
}

// 2026-10-26 P1 回归：白名单只比较 argv[0] 的 basename，
// 带路径的 command 会执行窗外的同名程序。
func TestShellJail_RejectsArgv0PathOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX absolute /bin/ls path semantics differ on Windows")
	}
	root := t.TempDir()
	sh := NewShell().WithWhitelist([]string{"ls"}).WithAllowedWorkdirs([]string{root})
	res := execShellResult(t, sh, map[string]any{
		"action":  "execute",
		"command": "/bin/ls",
	})
	if !res.IsError || !strings.Contains(res.Content, "outside the allowed directories") {
		t.Fatalf("argv[0] path outside root was not rejected: %s", res.Content)
	}
}

// 2026-10-26 P1 回归：root 内指向窗外的符号链接目录不能作为 workdir。
func TestShellJail_RejectsSymlinkWorkdirEscape(t *testing.T) {
	if _, err := exec.LookPath("pwd"); err != nil {
		t.Skip("pwd not installed")
	}
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	sh := NewShell().WithWhitelist([]string{"pwd"}).WithAllowedWorkdirs([]string{root})
	res := execShellResult(t, sh, map[string]any{
		"action":  "execute",
		"command": "pwd",
		"workdir": link,
	})
	if !res.IsError || !strings.Contains(res.Content, "allowed directories") {
		t.Fatalf("symlink workdir escape was not rejected: %s", res.Content)
	}
}

// 2026-10-26 P2 回归：输出必须有界，不能先全量进内存再截断。
func TestShell_OutputIsCapped(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat not installed")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "big.txt"), []byte(strings.Repeat("A", 32*1024)), 0o644); err != nil {
		t.Fatal(err)
	}
	sh := NewShell().WithWhitelist([]string{"cat"}).WithAllowedWorkdirs([]string{root}).WithMaxOutputSize(64)
	res := execShellResult(t, sh, map[string]any{
		"action":  "execute",
		"command": "cat big.txt",
	})
	if res.IsError {
		t.Fatalf("cat failed: %s", res.Content)
	}
	if !strings.Contains(res.Content, "输出已截断") {
		t.Fatalf("output was not capped: %s", res.Content)
	}
	if len(res.Content) > 1024 {
		t.Fatalf("capped output unexpectedly large: %d bytes", len(res.Content))
	}
}

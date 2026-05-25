package api

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveWithinRoot(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveWithinRoot(root, "file.json"); err != nil {
		t.Errorf("expected ok, got %v", err)
	}
	if _, err := ResolveWithinRoot(root, "../etc/passwd"); !errors.Is(err, ErrUnsafePath) {
		t.Errorf("expected ErrUnsafePath, got %v", err)
	}
	if _, err := ResolveWithinRoot(root, ""); !errors.Is(err, ErrUnsafePath) {
		t.Errorf("expected ErrUnsafePath for empty, got %v", err)
	}
	// 绝对路径在不同平台的形态不同：Unix 的 "/etc/passwd" 在 Windows 上
	// filepath.IsAbs 会返回 false（Windows 绝对路径需要盘符或 UNC），
	// 这里按运行平台选择典型的"越界绝对路径"样本。
	absSample := "/etc/passwd"
	if runtime.GOOS == "windows" {
		absSample = `C:\Windows\System32\drivers\etc\hosts`
	}
	if _, err := ResolveWithinRoot(root, absSample); !errors.Is(err, ErrUnsafePath) {
		t.Errorf("expected ErrUnsafePath for abs path, got %v", err)
	}
	// nested ok
	nested := filepath.Join(root, "sub", "a.json")
	if _, err := ResolveWithinRoot(root, "sub/a.json"); err != nil {
		t.Errorf("expected ok for nested, got %v (nested=%s)", err, nested)
	}
}

func TestResolveLeafFile(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "ok.json")
	if err := os.WriteFile(target, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveLeafFile(root, "ok.json", "json"); err != nil {
		t.Errorf("expected ok, got %v", err)
	}
	if _, err := ResolveLeafFile(root, "ok.json", "toml"); !errors.Is(err, ErrUnsafePath) {
		t.Errorf("expected ext mismatch ErrUnsafePath, got %v", err)
	}
	if _, err := ResolveLeafFile(root, "../ok.json", "json"); !errors.Is(err, ErrUnsafePath) {
		t.Errorf("expected ErrUnsafePath for parent ref, got %v", err)
	}
	if _, err := ResolveLeafFile(root, "sub/ok.json", "json"); !errors.Is(err, ErrUnsafePath) {
		t.Errorf("expected ErrUnsafePath for nested name, got %v", err)
	}
	if _, err := ResolveLeafFile(root, "missing.json", "json"); !errors.Is(err, ErrUnsafePath) {
		t.Errorf("expected ErrUnsafePath for missing file, got %v", err)
	}
}

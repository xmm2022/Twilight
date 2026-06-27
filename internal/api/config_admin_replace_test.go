package api

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestReplaceConfigFileFallsBackForBusyBindMount(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.toml")
	tmp := filepath.Join(dir, "config.toml.tmp")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmp, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := replaceConfigFile(tmp, target, func(_, _ string) error {
		return &os.LinkError{Op: "rename", Old: tmp, New: target, Err: syscall.EBUSY}
	})
	if err != nil {
		t.Fatalf("replaceConfigFile returned error: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Fatalf("target content = %q, want new", content)
	}
	if _, err := os.Stat(tmp); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("tmp still exists or unexpected stat error: %v", err)
	}
}

func TestReplaceConfigFileKeepsRenameErrors(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.toml")
	tmp := filepath.Join(dir, "config.toml.tmp")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmp, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	renameErr := &os.LinkError{Op: "rename", Old: tmp, New: target, Err: syscall.EACCES}

	err := replaceConfigFile(tmp, target, func(_, _ string) error { return renameErr })
	if !errors.Is(err, syscall.EACCES) {
		t.Fatalf("error = %v, want EACCES", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "old" {
		t.Fatalf("target content = %q, want old", content)
	}
}

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommandPrintsResolvedRequestFilePath(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, cwd, "users.http", "GET http://example.com\n")
	t.Chdir(cwd)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCommand(stdout, stderr)
	cmd.SetArgs([]string{"users.http"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := filepath.Join(cwd, "users.http") + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootCommandRequiresRequestFileArgument(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCommand(stdout, stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want argument error")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Fatalf("error = %q, want argument error", err.Error())
	}
}

func TestRootCommandReturnsErrorForMissingRequestFile(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCommand(stdout, stderr)
	cmd.SetArgs([]string{"missing.http"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want request file error")
	}
	if !strings.Contains(err.Error(), "request file not found") {
		t.Fatalf("error = %q, want request file not found", err.Error())
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func writeFile(t *testing.T, basePath, name, contents string) {
	t.Helper()

	path := filepath.Join(basePath, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

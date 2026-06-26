package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommandPrintsParsedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("response body"))
	}))
	t.Cleanup(server.Close)

	cwd := t.TempDir()
	writeFile(t, cwd, "users.http", "GET "+server.URL+"\n")
	t.Chdir(cwd)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCommand(stdout, stderr)
	cmd.SetArgs([]string{"users.http"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := "response body"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootCommandDebugPrintsResolvedRequestFileAndParsedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("debug body"))
	}))
	t.Cleanup(server.Close)

	cwd := t.TempDir()
	writeFile(t, cwd, "users.http", "GET "+server.URL+"\nAccept: application/json\n")
	t.Chdir(cwd)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCommand(stdout, stderr)
	cmd.SetArgs([]string{"--debug", "users.http"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := filepath.Join(cwd, "users.http") + "\nGET " + server.URL + "\nAccept: application/json\nStatus: 200\ndebug body"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootCommandSendsNamedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		_, _ = w.Write([]byte("created"))
	}))
	t.Cleanup(server.Close)

	cwd := t.TempDir()
	writeFile(t, cwd, "users.http", "GET "+server.URL+"/first\n\n### Create user\nPOST "+server.URL+"/users\n")
	t.Chdir(cwd)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := NewRootCommand(stdout, stderr)
	cmd.SetArgs([]string{"users.http", "Create user"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if stdout.String() != "created" {
		t.Fatalf("stdout = %q, want created", stdout.String())
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
	if !strings.Contains(err.Error(), "accepts between 1 and 2 arg") {
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

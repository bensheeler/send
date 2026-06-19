package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bensheeler/send/core/scanner"
)

func TestScanRequestFileFindsRequestFile(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, cwd, "users.http", "GET http://example.com\n")

	result, err := ScanRequestFile(ScanRequestFileInput{
		CWD:      cwd,
		Selector: "users.http",
	})
	if err != nil {
		t.Fatalf("ScanRequestFile returned error: %v", err)
	}

	want := filepath.Join(cwd, "users.http")
	if result.Path != want {
		t.Fatalf("Path = %q, want %q", result.Path, want)
	}
}

func TestScanRequestFileReturnsScannerErrors(t *testing.T) {
	cwd := t.TempDir()

	_, err := ScanRequestFile(ScanRequestFileInput{
		CWD:      cwd,
		Selector: "missing.http",
	})
	if !errors.Is(err, scanner.ErrNotFound) {
		t.Fatalf("error = %v, want scanner.ErrNotFound", err)
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

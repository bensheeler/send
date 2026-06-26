package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestLoadRequestScansAndParsesRequestFile(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, cwd, "examples/http/single-line.http", "GET https://example.com/users/1\nAccept: application/json\n")

	result, err := LoadRequest(LoadRequestInput{
		CWD:      cwd,
		Selector: "single-line",
	})
	if err != nil {
		t.Fatalf("LoadRequest returned error: %v", err)
	}

	wantPath := filepath.Join(cwd, "examples", "http", "single-line.http")
	if result.Path != wantPath {
		t.Fatalf("Path = %q, want %q", result.Path, wantPath)
	}
	if result.Method != "GET" {
		t.Fatalf("Method = %q, want GET", result.Method)
	}
	if result.URL != "https://example.com/users/1" {
		t.Fatalf("URL = %q, want https://example.com/users/1", result.URL)
	}
	if len(result.Headers) != 1 {
		t.Fatalf("len(Headers) = %d, want 1", len(result.Headers))
	}
	if result.Headers[0].Name != "Accept" || result.Headers[0].Value != "application/json" {
		t.Fatalf("Headers[0] = %#v, want Accept: application/json", result.Headers[0])
	}
}

func TestLoadRequestReturnsFirstRequestFromMultiRequestFile(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, cwd, "requests/users.http", "GET https://example.com/users/1\n\n### createUser\nPOST https://example.com/users\n")

	result, err := LoadRequest(LoadRequestInput{
		CWD:      cwd,
		Selector: "requests/users.http",
	})
	if err != nil {
		t.Fatalf("LoadRequest returned error: %v", err)
	}

	if result.Method != "GET" {
		t.Fatalf("Method = %q, want GET", result.Method)
	}
	if result.URL != "https://example.com/users/1" {
		t.Fatalf("URL = %q, want first request URL", result.URL)
	}
}

func TestLoadRequestSelectsNamedRequest(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, cwd, "requests/users.http", "GET https://example.com/users/1\n\n### Create user\nPOST https://example.com/users\n")

	result, err := LoadRequest(LoadRequestInput{
		CWD:         cwd,
		Selector:    "requests/users.http",
		RequestName: "Create user",
	})
	if err != nil {
		t.Fatalf("LoadRequest returned error: %v", err)
	}

	if result.Method != "POST" {
		t.Fatalf("Method = %q, want POST", result.Method)
	}
	if result.URL != "https://example.com/users" {
		t.Fatalf("URL = %q, want named request URL", result.URL)
	}
}

func TestLoadRequestReturnsErrorForMissingRequestName(t *testing.T) {
	cwd := t.TempDir()
	writeFile(t, cwd, "requests/users.http", "GET https://example.com/users/1\n\n### createUser\nPOST https://example.com/users\n")

	_, err := LoadRequest(LoadRequestInput{
		CWD:         cwd,
		Selector:    "requests/users.http",
		RequestName: "CreateUser",
	})
	if !errors.Is(err, ErrRequestNameNotFound) {
		t.Fatalf("error = %v, want ErrRequestNameNotFound", err)
	}
}

func TestSendRequestLoadsAndRunsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("Authorization = %q, want Bearer token", got)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("sent"))
	}))
	t.Cleanup(server.Close)

	cwd := t.TempDir()
	writeFile(t, cwd, "requests/users.http", "GET "+server.URL+"\nAuthorization: Bearer token\n")

	result, err := SendRequest(context.Background(), SendRequestInput{
		CWD:      cwd,
		Selector: "requests/users.http",
	})
	if err != nil {
		t.Fatalf("SendRequest returned error: %v", err)
	}

	wantPath := filepath.Join(cwd, "requests", "users.http")
	if result.Path != wantPath {
		t.Fatalf("Path = %q, want %q", result.Path, wantPath)
	}
	if result.Method != http.MethodGet {
		t.Fatalf("Method = %q, want GET", result.Method)
	}
	if result.URL != server.URL {
		t.Fatalf("URL = %q, want %q", result.URL, server.URL)
	}
	if len(result.Headers) != 1 {
		t.Fatalf("len(Headers) = %d, want 1", len(result.Headers))
	}
	if result.Headers[0].Name != "Authorization" || result.Headers[0].Value != "Bearer token" {
		t.Fatalf("Headers[0] = %#v, want Authorization: Bearer token", result.Headers[0])
	}
	if result.StatusCode != http.StatusAccepted {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusAccepted)
	}
	if string(result.Body) != "sent" {
		t.Fatalf("Body = %q, want sent", result.Body)
	}
}

func TestSendRequestCarriesBodyThroughToRunner(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if string(body) != "{\"name\":\"Ada\"}" {
			t.Fatalf("body = %q, want JSON body", body)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	t.Cleanup(server.Close)

	cwd := t.TempDir()
	writeFile(t, cwd, "requests/create-user.http", "POST "+server.URL+"\nContent-Type: application/json\n\n{\"name\":\"Ada\"}\n")

	result, err := SendRequest(context.Background(), SendRequestInput{
		CWD:      cwd,
		Selector: "requests/create-user.http",
	})
	if err != nil {
		t.Fatalf("SendRequest returned error: %v", err)
	}

	if result.StatusCode != http.StatusCreated {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusCreated)
	}
	if string(result.Body) != "created" {
		t.Fatalf("Body = %q, want created", result.Body)
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

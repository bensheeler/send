package runner

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunSendsRequestAndReturnsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))
	t.Cleanup(server.Close)

	response, err := Run(context.Background(), server.Client(), Request{
		Method: http.MethodGet,
		URL:    server.URL,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if string(response.Body) != "hello" {
		t.Fatalf("Body = %q, want hello", response.Body)
	}
}

func TestRunUsesRequestMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(server.Close)

	_, err := Run(context.Background(), server.Client(), Request{
		Method: http.MethodPost,
		URL:    server.URL,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestRunSendsRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Values("X-Trace"); len(got) != 2 || got[0] != "one" || got[1] != "two" {
			t.Fatalf("X-Trace values = %#v, want [one two]", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("Authorization = %q, want Bearer token", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	_, err := Run(context.Background(), server.Client(), Request{
		Method: http.MethodGet,
		URL:    server.URL,
		Headers: []Header{
			{Name: "X-Trace", Value: "one"},
			{Name: "X-Trace", Value: "two"},
			{Name: "Authorization", Value: "Bearer token"},
		},
	})
	if err != nil {
		 t.Fatalf("Run returned error: %v", err)
	}
}

func TestRunSendsRequestBody(t *testing.T) {
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
	}))
	t.Cleanup(server.Close)

	_, err := Run(context.Background(), server.Client(), Request{
		Method: http.MethodPost,
		URL:    server.URL,
		Headers: []Header{
			{Name: "Content-Type", Value: "application/json"},
		},
		Body: []byte("{\"name\":\"Ada\"}"),
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestRunReturnsNon2xxStatusAndBodyWithoutError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("missing"))
	}))
	t.Cleanup(server.Close)

	response, err := Run(context.Background(), server.Client(), Request{
		Method: http.MethodGet,
		URL:    server.URL,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
	if string(response.Body) != "missing" {
		t.Fatalf("Body = %q, want missing", response.Body)
	}
}

func TestRunReturnsErrorForInvalidURL(t *testing.T) {
	_, err := Run(context.Background(), nil, Request{
		Method: http.MethodGet,
		URL:    "://bad-url",
	})
	if err == nil {
		t.Fatal("Run error = nil, want error")
	}
}

package parser

import "testing"

func TestParseRequestsParsesSingleRequestLine(t *testing.T) {
	requests, err := ParseRequests([]byte("GET https://example.com/users\n"))
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(requests))
	}
	if requests[0].Method != "GET" {
		t.Fatalf("Method = %q, want GET", requests[0].Method)
	}
	if requests[0].URL != "https://example.com/users" {
		t.Fatalf("URL = %q, want URL", requests[0].URL)
	}
}

func TestParseRequestsNormalizesMethodToUppercase(t *testing.T) {
	requests, err := ParseRequests([]byte("post https://example.com/users\n"))
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if requests[0].Method != "POST" {
		t.Fatalf("Method = %q, want POST", requests[0].Method)
	}
}

func TestParseRequestsRejectsUnsupportedMethod(t *testing.T) {
	_, err := ParseRequests([]byte("TRACE https://example.com/users\n"))
	if err == nil {
		t.Fatal("ParseRequests error = nil, want error")
	}
}

func TestParseRequestsRejectsMalformedRequestLine(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{name: "empty", contents: ""},
		{name: "blank only", contents: "\n\t \n"},
		{name: "missing url", contents: "GET\n"},
		{name: "extra token", contents: "GET https://example.com HTTP/1.1\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRequests([]byte(tt.contents))
			if err == nil {
				t.Fatal("ParseRequests error = nil, want error")
			}
		})
	}
}

func TestParseRequestsSkipsLeadingBlankAndCommentLines(t *testing.T) {
	contents := []byte("\n# users API\n// smoke request\n\t  GET\t https://example.com/users\n")

	requests, err := ParseRequests(contents)
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if requests[0].Method != "GET" {
		t.Fatalf("Method = %q, want GET", requests[0].Method)
	}
	if requests[0].URL != "https://example.com/users" {
		t.Fatalf("URL = %q, want URL", requests[0].URL)
	}
}

func TestParseRequestsRejectsCommentsOnlyInput(t *testing.T) {
	_, err := ParseRequests([]byte("# users API\n// smoke request\n"))
	if err == nil {
		t.Fatal("ParseRequests error = nil, want error")
	}
}

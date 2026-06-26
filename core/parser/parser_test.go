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

func TestParseRequestsParsesHeadersAfterRequestLine(t *testing.T) {
	requests, err := ParseRequests([]byte("GET https://example.com/users\nAuthorization: Bearer token\nAccept: application/json\n"))
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if len(requests[0].Headers) != 2 {
		t.Fatalf("len(Headers) = %d, want 2", len(requests[0].Headers))
	}
	if requests[0].Headers[0].Name != "Authorization" || requests[0].Headers[0].Value != "Bearer token" {
		t.Fatalf("Headers[0] = %#v, want Authorization: Bearer token", requests[0].Headers[0])
	}
	if requests[0].Headers[1].Name != "Accept" || requests[0].Headers[1].Value != "application/json" {
		t.Fatalf("Headers[1] = %#v, want Accept: application/json", requests[0].Headers[1])
	}
}

func TestParseRequestsParsesRawBodyAfterHeaders(t *testing.T) {
	requests, err := ParseRequests([]byte("POST https://example.com/users\nContent-Type: application/json\n\n{\"name\":\"Ada\"}\n"))
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(requests))
	}
	if requests[0].Method != "POST" {
		t.Fatalf("Method = %q, want POST", requests[0].Method)
	}
	if len(requests[0].Headers) != 1 {
		t.Fatalf("len(Headers) = %d, want 1", len(requests[0].Headers))
	}
	if requests[0].Headers[0].Name != "Content-Type" || requests[0].Headers[0].Value != "application/json" {
		t.Fatalf("Headers[0] = %#v, want Content-Type: application/json", requests[0].Headers[0])
	}
	if string(requests[0].Body) != "{\"name\":\"Ada\"}" {
		t.Fatalf("Body = %q, want raw JSON body", requests[0].Body)
	}
}

func TestParseRequestsParsesRequestsSeparatedBySeparatorLine(t *testing.T) {
	contents := []byte("GET https://example.com/users/1\n\n###\n\nPOST https://example.com/users\nContent-Type: application/json\n\n{\"name\":\"Ada\"}\n")

	requests, err := ParseRequests(contents)
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if len(requests) != 2 {
		t.Fatalf("len(requests) = %d, want 2", len(requests))
	}
	if requests[0].Method != "GET" || requests[0].URL != "https://example.com/users/1" {
		t.Fatalf("requests[0] = %#v, want GET users/1", requests[0])
	}
	if requests[1].Method != "POST" || requests[1].URL != "https://example.com/users" {
		t.Fatalf("requests[1] = %#v, want POST users", requests[1])
	}
	if string(requests[1].Body) != "{\"name\":\"Ada\"}" {
		t.Fatalf("requests[1].Body = %q, want raw JSON body", requests[1].Body)
	}
}

func TestParseRequestsUsesSeparatorTitleAsRequestName(t *testing.T) {
	contents := []byte("### createUser\nPOST https://example.com/users\n")

	requests, err := ParseRequests(contents)
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(requests))
	}
	if requests[0].Name != "createUser" {
		t.Fatalf("Name = %q, want createUser", requests[0].Name)
	}
}

func TestParseRequestsIgnoresCommentPreambleBeforeFirstSeparator(t *testing.T) {
	contents := []byte("# Users API\n\n### listUsers\nGET https://example.com/users\n")

	requests, err := ParseRequests(contents)
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(requests))
	}
	if requests[0].Name != "listUsers" {
		t.Fatalf("Name = %q, want listUsers", requests[0].Name)
	}
	if requests[0].URL != "https://example.com/users" {
		t.Fatalf("URL = %q, want users URL", requests[0].URL)
	}
}

func TestParseRequestsUsesNameMetadataAsRequestName(t *testing.T) {
	contents := []byte("# users API\n# @name createUser\nPOST https://example.com/users\n")

	requests, err := ParseRequests(contents)
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if requests[0].Name != "createUser" {
		t.Fatalf("Name = %q, want createUser", requests[0].Name)
	}
}

func TestParseRequestsIgnoresMetadataThatOnlyPrefixesName(t *testing.T) {
	contents := []byte("# @namex createUser\n# @namespace users\nPOST https://example.com/users\n")

	requests, err := ParseRequests(contents)
	if err != nil {
		t.Fatalf("ParseRequests returned error: %v", err)
	}

	if requests[0].Name != "" {
		t.Fatalf("Name = %q, want empty", requests[0].Name)
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

func TestParseRequestsRejectsMalformedHeaderLine(t *testing.T) {
	_, err := ParseRequests([]byte("GET https://example.com/users\nAuthorization Bearer token\n"))
	if err == nil {
		t.Fatal("ParseRequests error = nil, want error")
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

package parser

import "strings"

type Header struct {
	Name  string
	Value string
}

type Request struct {
	Name        string
	Method      string
	URL         string
	HTTPVersion string
	Headers     []Header
	Body        []byte
}

type requestBlock struct {
	Name  string
	Lines []string
}

type ParseError struct {
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}

func ParseRequests(contents []byte) ([]Request, error) {
	lines := strings.Split(string(contents), "\n")
	blocks := splitRequestBlocks(lines)
	requests := make([]Request, 0, len(blocks))
	for _, block := range blocks {
		request, err := parseRequestBlock(block.Name, block.Lines)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if len(requests) == 0 {
		return nil, &ParseError{Message: "request line not found"}
	}
	return requests, nil
}

func parseRequestBlock(name string, lines []string) (Request, error) {
	line, requestLineIndex, ok := firstRequestLine(lines)
	if !ok {
		return Request{}, &ParseError{Message: "request line not found"}
	}
	if metadataName := requestNameMetadata(lines[:requestLineIndex]); metadataName != "" {
		name = metadataName
	}

	method, url, httpVersion, ok := parseRequestLine(line)
	if !ok {
		return Request{}, &ParseError{Message: "malformed request line"}
	}
	if !isSupportedMethod(method) {
		return Request{}, &ParseError{Message: "unsupported HTTP method"}
	}
	if httpVersion != "" && !isSupportedHTTPVersion(httpVersion) {
		return Request{}, &ParseError{Message: "unsupported HTTP version"}
	}

	headers, body, err := parseHeadersAndBody(lines[requestLineIndex+1:])
	if err != nil {
		return Request{}, err
	}

	return Request{Name: name, Method: method, URL: url, HTTPVersion: httpVersion, Headers: headers, Body: body}, nil
}

func parseRequestLine(line string) (string, string, string, bool) {
	parts := strings.Fields(line)
	switch len(parts) {
	case 1:
		if !isURLOnlyRequestLine(parts[0]) {
			return "", "", "", false
		}
		return "GET", parts[0], "", true
	case 2:
		return strings.ToUpper(parts[0]), parts[1], "", true
	case 3:
		return strings.ToUpper(parts[0]), parts[1], parts[2], true
	default:
		return "", "", "", false
	}
}

func isURLOnlyRequestLine(line string) bool {
	lower := strings.ToLower(line)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func splitRequestBlocks(lines []string) []requestBlock {
	var blocks []requestBlock
	start := 0
	name := ""
	for index, line := range lines {
		if isSeparatorLine(line) {
			if hasRequestLineCandidate(lines[start:index]) {
				blocks = append(blocks, requestBlock{Name: name, Lines: lines[start:index]})
			}
			name = separatorTitle(line)
			start = index + 1
		}
	}
	if hasRequestLineCandidate(lines[start:]) {
		blocks = append(blocks, requestBlock{Name: name, Lines: lines[start:]})
	}
	return blocks
}

func hasRequestLineCandidate(lines []string) bool {
	_, _, ok := firstRequestLine(lines)
	return ok
}

func firstRequestLine(lines []string) (string, int, bool) {
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}
		return trimmed, index, true
	}
	return "", 0, false
}

func parseHeadersAndBody(lines []string) ([]Header, []byte, error) {
	var headers []Header
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			body := strings.Join(lines[index+1:], "\n")
			body = strings.TrimRight(body, "\r\n")
			return headers, []byte(body), nil
		}

		name, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(name) == "" {
			return nil, nil, &ParseError{Message: "malformed header line"}
		}
		headers = append(headers, Header{
			Name:  strings.TrimSpace(name),
			Value: strings.TrimSpace(value),
		})
	}
	return headers, nil, nil
}

func isCommentLine(line string) bool {
	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//")
}

func requestNameMetadata(lines []string) string {
	name := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		text, ok := commentText(trimmed)
		if !ok {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(text))
		if len(fields) >= 2 && fields[0] == "@name" {
			name = strings.Join(fields[1:], " ")
		}
	}
	return name
}

func commentText(line string) (string, bool) {
	if text, ok := strings.CutPrefix(line, "#"); ok {
		return text, true
	}
	if text, ok := strings.CutPrefix(line, "//"); ok {
		return text, true
	}
	return "", false
}

func isSeparatorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return separatorHashCount(trimmed) >= 3
}

func separatorTitle(line string) string {
	trimmed := strings.TrimSpace(line)
	return strings.TrimSpace(trimmed[separatorHashCount(trimmed):])
}

func separatorHashCount(line string) int {
	count := 0
	for _, char := range line {
		if char != '#' {
			break
		}
		count++
	}
	return count
}

func isSupportedMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "PATCH", "OPTIONS", "TRACE",
		"LOCK", "UNLOCK", "PROPFIND", "PROPPATCH", "COPY", "MOVE", "MKCOL", "MKCALENDAR", "ACL", "SEARCH":
		return true
	default:
		return false
	}
}

func isSupportedHTTPVersion(version string) bool {
	switch version {
	case "HTTP/1.1", "HTTP/2":
		return true
	default:
		return false
	}
}

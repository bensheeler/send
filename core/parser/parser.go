package parser

import "strings"

type Request struct {
	Method string
	URL    string
}

type ParseError struct {
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}

func ParseRequests(contents []byte) ([]Request, error) {
	line, ok := firstRequestLine(contents)
	if !ok {
		return nil, &ParseError{Message: "request line not found"}
	}

	parts := strings.Fields(line)
	if len(parts) != 2 {
		return nil, &ParseError{Message: "malformed request line"}
	}

	method := strings.ToUpper(parts[0])
	if !isSupportedMethod(method) {
		return nil, &ParseError{Message: "unsupported HTTP method"}
	}
	return []Request{{Method: method, URL: parts[1]}}, nil
}

func firstRequestLine(contents []byte) (string, bool) {
	for _, line := range strings.Split(string(contents), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}
		return trimmed, true
	}
	return "", false
}

func isCommentLine(line string) bool {
	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//")
}

func isSupportedMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

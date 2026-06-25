package parser

import "strings"

type Header struct {
	Name  string
	Value string
}

type Request struct {
	Method  string
	URL     string
	Headers []Header
	Body    []byte
}

type ParseError struct {
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}

func ParseRequests(contents []byte) ([]Request, error) {
	lines := strings.Split(string(contents), "\n")
	line, requestLineIndex, ok := firstRequestLine(lines)
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

	headers, body, err := parseHeadersAndBody(lines[requestLineIndex+1:])
	if err != nil {
		return nil, err
	}

	return []Request{{Method: method, URL: parts[1], Headers: headers, Body: body}}, nil
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

func isSupportedMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

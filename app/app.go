package app

import (
	"context"
	"errors"
	"os"

	"github.com/bensheeler/send/core/parser"
	"github.com/bensheeler/send/core/runner"
	"github.com/bensheeler/send/core/scanner"
)

const DefaultLookupDepth = 3

type ScanRequestFileInput struct {
	CWD      string
	Selector string
}

type ScanRequestFileResult struct {
	Path string
}

type LoadRequestInput struct {
	CWD         string
	Selector    string
	RequestName string
}

type Header struct {
	Name  string
	Value string
}

type LoadRequestResult struct {
	Path    string
	Method  string
	URL     string
	Headers []Header
	Body    []byte
}

type SendRequestInput struct {
	CWD         string
	Selector    string
	RequestName string
}

type SendRequestResult struct {
	Path       string
	Method     string
	URL        string
	Headers    []Header
	StatusCode int
	Body       []byte
}

var ErrRequestNameNotFound = errors.New("request name not found")

func ScanRequestFile(input ScanRequestFileInput) (ScanRequestFileResult, error) {
	result, err := scanner.FindRequestFile(input.CWD, input.Selector, scanner.Options{
		LookupDepth: DefaultLookupDepth,
	})
	if err != nil {
		return ScanRequestFileResult{}, err
	}

	return ScanRequestFileResult{Path: result.Path}, nil
}

func LoadRequest(input LoadRequestInput) (LoadRequestResult, error) {
	scanResult, err := ScanRequestFile(ScanRequestFileInput{
		CWD:      input.CWD,
		Selector: input.Selector,
	})
	if err != nil {
		return LoadRequestResult{}, err
	}

	contents, err := os.ReadFile(scanResult.Path)
	if err != nil {
		return LoadRequestResult{}, err
	}

	requests, err := parser.ParseRequests(contents)
	if err != nil {
		return LoadRequestResult{}, err
	}

	request := requests[0]
	if input.RequestName != "" {
		var ok bool
		request, ok = requestByName(requests, input.RequestName)
		if !ok {
			return LoadRequestResult{}, ErrRequestNameNotFound
		}
	}
	headers := make([]Header, 0, len(request.Headers))
	for _, header := range request.Headers {
		headers = append(headers, Header{Name: header.Name, Value: header.Value})
	}

	return LoadRequestResult{
		Path:    scanResult.Path,
		Method:  request.Method,
		URL:     request.URL,
		Headers: headers,
		Body:    request.Body,
	}, nil
}

func requestByName(requests []parser.Request, name string) (parser.Request, bool) {
	for _, request := range requests {
		if request.Name == name {
			return request, true
		}
	}
	return parser.Request{}, false
}

func SendRequest(ctx context.Context, input SendRequestInput) (SendRequestResult, error) {
	request, err := LoadRequest(LoadRequestInput{
		CWD:         input.CWD,
		Selector:    input.Selector,
		RequestName: input.RequestName,
	})
	if err != nil {
		return SendRequestResult{}, err
	}

	headers := make([]runner.Header, 0, len(request.Headers))
	for _, header := range request.Headers {
		headers = append(headers, runner.Header{Name: header.Name, Value: header.Value})
	}

	response, err := runner.Run(ctx, nil, runner.Request{
		Method:  request.Method,
		URL:     request.URL,
		Headers: headers,
		Body:    request.Body,
	})
	if err != nil {
		return SendRequestResult{}, err
	}

	return SendRequestResult{
		Path:       request.Path,
		Method:     request.Method,
		URL:        request.URL,
		Headers:    request.Headers,
		StatusCode: response.StatusCode,
		Body:       response.Body,
	}, nil
}

package app

import (
	"context"
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
	CWD      string
	Selector string
}

type LoadRequestResult struct {
	Path   string
	Method string
	URL    string
}

type SendRequestInput struct {
	CWD      string
	Selector string
}

type SendRequestResult struct {
	Path       string
	Method     string
	URL        string
	StatusCode int
	Body       []byte
}

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
	return LoadRequestResult{
		Path:   scanResult.Path,
		Method: request.Method,
		URL:    request.URL,
	}, nil
}

func SendRequest(ctx context.Context, input SendRequestInput) (SendRequestResult, error) {
	request, err := LoadRequest(LoadRequestInput{
		CWD:      input.CWD,
		Selector: input.Selector,
	})
	if err != nil {
		return SendRequestResult{}, err
	}

	response, err := runner.Run(ctx, nil, runner.Request{
		Method: request.Method,
		URL:    request.URL,
	})
	if err != nil {
		return SendRequestResult{}, err
	}

	return SendRequestResult{
		Path:       request.Path,
		Method:     request.Method,
		URL:        request.URL,
		StatusCode: response.StatusCode,
		Body:       response.Body,
	}, nil
}

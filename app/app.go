package app

import "github.com/bensheeler/send/core/scanner"

const DefaultLookupDepth = 3

type ScanRequestFileInput struct {
	CWD      string
	Selector string
}

type ScanRequestFileResult struct {
	Path string
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

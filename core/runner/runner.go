package runner

import (
	"context"
	"io"
	"net/http"
)

type Header struct {
	Name  string
	Value string
}

type Request struct {
	Method  string
	URL     string
	Headers []Header
}

type Response struct {
	StatusCode int
	Body       []byte
}

func Run(ctx context.Context, client *http.Client, request Request) (Response, error) {
	if client == nil {
		client = http.DefaultClient
	}

	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, request.URL, nil)
	if err != nil {
		return Response{}, err
	}
	for _, header := range request.Headers {
		httpRequest.Header.Add(header.Name, header.Value)
	}

	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return Response{}, err
	}
	defer httpResponse.Body.Close()

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return Response{}, err
	}

	return Response{
		StatusCode: httpResponse.StatusCode,
		Body:       body,
	}, nil
}

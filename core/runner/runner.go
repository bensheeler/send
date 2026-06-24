package runner

import (
	"context"
	"io"
	"net/http"
)

type Request struct {
	Method string
	URL    string
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

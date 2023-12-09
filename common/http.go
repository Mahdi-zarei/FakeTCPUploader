package common

import (
	"context"
	"io"
	"net/http"
)

func CreateHttpPostRequest(ctx context.Context, contentType string, url string, reader *RateLimitReader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(reader)
	req.ContentLength = reader.TotalBytes()
	snapshot := reader
	req.GetBody = func() (io.ReadCloser, error) {
		r := snapshot
		return io.NopCloser(r), nil
	}

	return req, nil
}

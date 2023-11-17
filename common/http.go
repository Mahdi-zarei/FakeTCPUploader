package common

import (
	"io"
	"net/http"
)

func CreateHttpPostRequest(contentType string, url string, b []byte, maxrate int64) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	limitReader := NewRateLimiterReader(b, maxrate)
	req.Body = io.NopCloser(limitReader)
	req.ContentLength = int64(len(b))
	snapshot := *limitReader
	req.GetBody = func() (io.ReadCloser, error) {
		r := snapshot
		return io.NopCloser(&r), nil
	}

	return req, nil
}

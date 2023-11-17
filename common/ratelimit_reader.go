package common

import (
	"bytes"
	"io"
	"time"
)

const defaultReadAmount = 1 << 20

type RateLimitReader struct {
	*bytes.Reader
	maxrate  int64
	lastRead time.Time
}

func NewRateLimiterReader(b []byte, maxRate int64) *RateLimitReader {
	return &RateLimitReader{
		Reader:  bytes.NewReader(b),
		maxrate: maxRate,
	}
}

func (r *RateLimitReader) allowedReadAmount() int64 {
	if r.lastRead.IsZero() {
		return defaultReadAmount
	}

	allowedRead := int64(float64(r.maxrate) * time.Since(r.lastRead).Seconds())
	allowedRead = max(defaultReadAmount, allowedRead)
	return allowedRead
}

func (r *RateLimitReader) Read(p []byte) (n int, err error) {
	readSize := r.allowedReadAmount()
	readSize = min(int64(len(p)), readSize)
	r.lastRead = time.Now()
	return r.Reader.Read(p[:readSize])
}

func (r *RateLimitReader) WriteTo(w io.Writer) (int64, error) {
	var totalWrite int64
	for {
		readSize := r.allowedReadAmount()
		readSize = min(int64(r.Reader.Len()), readSize)
		r.lastRead = time.Now()
		b := make([]byte, readSize)
		n, e := r.Reader.Read(b)
		if e != nil && e != io.EOF {
			return totalWrite, e
		}
		m, e2 := w.Write(b[:n])
		totalWrite += int64(m)
		if m < n && e2 == nil {
			return totalWrite, io.ErrShortWrite
		}
		if e2 != nil {
			return totalWrite, e2
		}
		if e == io.EOF {
			return totalWrite, nil
		}
	}
}

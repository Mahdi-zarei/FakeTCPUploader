package common

import (
	"bytes"
	"io"
	"math/rand"
	"time"
)

const defaultReadAmount = 1 << 15

type RateLimitReader struct {
	*bytes.Reader
	maxrate  int64
	lastRead time.Time
}

func NewRateLimiterReader(b []byte, maxRate int64) *RateLimitReader {
	rateLimiter := &RateLimitReader{
		Reader:  bytes.NewReader(b),
		maxrate: maxRate,
	}
	return rateLimiter
}

func (r *RateLimitReader) allowedReadAmount() int64 {
	if r.lastRead.IsZero() {
		return defaultReadAmount
	}
	timeToSleep := float64(1000)/float64((r.maxrate+defaultReadAmount-1)/defaultReadAmount) - float64(time.Since(r.lastRead).Milliseconds())
	time.Sleep(time.Duration(int64(timeToSleep)) * time.Millisecond)

	extraCoef := float64(rand.Int63n(4000)) / 10000

	return int64((1 + extraCoef) * float64(r.maxrate) * time.Since(r.lastRead).Seconds())
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

func (r *RateLimitReader) TotalBytes() int64 {
	return r.Reader.Size()
}

func (r *RateLimitReader) BytesRead() int64 {
	return r.Reader.Size() - int64(r.Reader.Len())
}

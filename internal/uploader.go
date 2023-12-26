package internal

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/logs"
	"bytes"
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Uploader struct {
	chunkSize      int64
	watcher        *RateWatcher
	baseData       []byte
	counter        int
	parallelFactor int
}

func NewUploader(chunkSize int64, parallelFactor int, watcher *RateWatcher) *Uploader {
	b := []byte("0987654321asdfghjklqwertyuiozxcvbnm")
	baseData := bytes.Repeat(b, int(chunkSize)/len(b))
	baseData = append(baseData, b[:int(chunkSize)-len(baseData)]...)
	return &Uploader{
		chunkSize:      chunkSize,
		watcher:        watcher,
		baseData:       baseData,
		parallelFactor: parallelFactor,
	}
}

func (u *Uploader) SendData(address string, maxRate int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := time.Now()
	wg := &sync.WaitGroup{}
	sentVal := atomic.Int64{}
	wg.Add(u.parallelFactor)
	rate := maxRate / int64(u.parallelFactor)
	for i := 0; i < u.parallelFactor; i++ {
		go func() {
			r, e := u.sendDataMultiConn(ctx, address, rate, u.parallelFactor, wg)
			if e != nil {
				logs.Logger.Printf("error sending for %v : %v", address, e)
			}
			sentVal.Add(r)
		}()
	}
	wg.Wait()
	u.watcher.WatchQuality(address, maxRate, sentVal.Load(), time.Since(start))
	return nil
}

func (u *Uploader) sendDataMultiConn(ctx context.Context, address string, maxRate int64, splitFactor int, wg *sync.WaitGroup) (int64, error) {
	defer wg.Done()
	rateLimitedBody := common.NewRateLimiterReader(u.baseData[:len(u.baseData)/splitFactor], maxRate)
	req, err := common.CreateHttpPostRequest(ctx, "application/octet-stream", address, rateLimitedBody)
	if err != nil {
		return rateLimitedBody.BytesRead(), err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return rateLimitedBody.BytesRead(), err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return rateLimitedBody.BytesRead(), errors.New("non 200 status code: " + strconv.Itoa(resp.StatusCode))
	}

	return rateLimitedBody.BytesRead(), nil
}

func (u *Uploader) SendParallel(count int, maxTransferRate int64) {
	if constants.DEBUG {
		logs.Logger.Println("sending ", count)
	}
	if count == 0 {
		return
	}
	transferRate := maxTransferRate / int64(count)
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := u.watcher.GetAddr()
			_ = u.SendData(addr, transferRate)
		}()
	}
	wg.Wait()
}

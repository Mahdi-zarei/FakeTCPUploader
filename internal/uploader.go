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
	"time"
)

type Uploader struct {
	chunkSize int64
	watcher   *RateWatcher
	baseData  []byte
	counter   int
}

func NewUploader(chunkSize int64, addresses []string) *Uploader {
	b := []byte("0987654321asdfghjklqwertyuiozxcvbnm")
	baseData := bytes.Repeat(b, int(chunkSize)/len(b))
	baseData = append(baseData, b[:int(chunkSize)-len(baseData)]...)
	return &Uploader{
		chunkSize: chunkSize,
		watcher:   NewRateWatcher(addresses, 4, 5),
		baseData:  baseData,
	}
}

func (u *Uploader) SendData(address string, maxRate int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	start := time.Now()
	rateLimitedBody := common.NewRateLimiterReader(u.baseData, maxRate)
	req, err := common.CreateHttpPostRequest(ctx, "application/octet-stream", address, rateLimitedBody)
	if err != nil {
		u.watcher.WatchQuality(address, maxRate, rateLimitedBody.BytesRead(), time.Since(start))
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		u.watcher.WatchQuality(address, maxRate, rateLimitedBody.BytesRead(), time.Since(start))
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		u.watcher.WatchQuality(address, maxRate, rateLimitedBody.BytesRead(), time.Since(start))
		return errors.New("non 200 status code: " + strconv.Itoa(resp.StatusCode))
	}
	u.watcher.WatchQuality(address, maxRate, rateLimitedBody.BytesRead(), time.Since(start))
	return nil
}

func (u *Uploader) SendParallel(count int, maxTransferRate int64) {
	if constants.DEBUG {
		logs.Logger.Println("sending ", count)
	}
	transferRate := maxTransferRate / int64(count)
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startTime := time.Now()
			addr := u.watcher.GetAddr()
			err := u.SendData(addr, transferRate)
			t := time.Since(startTime)
			if err != nil {
				logs.Logger.Printf("error in sending after %v: %v", common.FormatFloat64(t.Seconds(), 2), err)
			} else {
				logs.Logger.Printf("finished sending for %v in %v", addr, common.FormatFloat64(t.Seconds(), 2))
			}
		}()
	}
	wg.Wait()
}

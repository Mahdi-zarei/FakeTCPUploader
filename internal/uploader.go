package internal

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/logs"
	"bytes"
	"context"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Uploader struct {
	chunkSize int64
	addresses []string
	baseData  []byte
	counter   int
}

func NewUploader(chunkSize int64, addresses []string) *Uploader {
	b := []byte("0987654321asdfghjklqwertyuiozxcvbnm")
	baseData := bytes.Repeat(b, int(chunkSize)/len(b))
	baseData = append(baseData, b[:int(chunkSize)-len(baseData)]...)
	return &Uploader{
		chunkSize: chunkSize,
		addresses: addresses,
		baseData:  baseData,
	}
}

func (u *Uploader) getRandomAddress() string {
	return u.addresses[rand.Int()%len(u.addresses)]
}

func (u *Uploader) SendData(address string, maxRate int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	req, err := common.CreateHttpPostRequest(ctx, "application/octet-stream", address, u.baseData, maxRate)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("non 200 status code: " + strconv.Itoa(resp.StatusCode))
	}
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
		go func(addrIdx int) {
			defer wg.Done()
			startTime := time.Now()
			err := u.SendData(u.addresses[addrIdx], transferRate)
			t := time.Since(startTime)
			if err != nil {
				logs.Logger.Printf("error in sending after %v: %v", t.Seconds(), err)
			} else {
				logs.Logger.Printf("finished sending for %v in %v", u.addresses[addrIdx], t.Seconds())
			}
		}(u.counter)
		u.counter++
		u.counter %= len(u.addresses)
	}
	wg.Wait()
}

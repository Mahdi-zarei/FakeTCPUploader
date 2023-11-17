package internal

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/logs"
	"bytes"
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
	req, err := common.CreateHttpPostRequest("application/octet-stream", address, u.baseData, maxRate)
	if err != nil {
		return err
	}

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if constants.DEBUG {
		logs.Logger.Println(address, " took ", time.Since(startTime).Milliseconds(), "ms")
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
			err := u.SendData(u.addresses[addrIdx], transferRate)
			if err != nil {
				logs.Logger.Println("error in sending: ", err)
			}
		}(u.counter)
		u.counter++
		u.counter %= len(u.addresses)
	}
	wg.Wait()
}

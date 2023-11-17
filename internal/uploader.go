package internal

import (
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
}

func NewUploader(chunkSize int64, addresses []string) *Uploader {
	return &Uploader{
		chunkSize: chunkSize,
		addresses: addresses,
		baseData:  []byte("0987654321asdfghjklqwertyuiozxcvbnm"),
	}
}

func (u *Uploader) getRandomAddress() string {
	return u.addresses[rand.Int()%len(u.addresses)]
}

func (u *Uploader) SendData() error {
	data := bytes.Repeat(u.baseData, int(float64(u.chunkSize)/float64(len(u.baseData))))
	data = append(data, u.baseData[:u.chunkSize-int64(len(data))]...)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.getRandomAddress(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Add("contentType", "application/octet-stream")

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

func (u *Uploader) SendParallel(count int) {
	if constants.DEBUG {
		logs.Logger.Println("sending ", count)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := u.SendData()
			if err != nil {
				logs.Logger.Println("error in sending: ", err)
			}
		}()
	}
	wg.Wait()
}

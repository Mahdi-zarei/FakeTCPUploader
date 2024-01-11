package internal

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/logs"
	"sync"
)

type Calculator struct {
	offset    int64
	currentRX int64
	currentTX int64
	mu        sync.RWMutex
	ratio     int64
}

func NewCalculator(ratio int64, offset int64, maxRate int64) *Calculator {
	return &Calculator{
		offset: offset,
		ratio:  ratio,
	}
}

func (c *Calculator) RegisterNew(RX, TX int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentRX = RX
	c.currentTX = TX
}

func (c *Calculator) GetNeededWrite() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	currRX := common.GBtoBytes(c.offset) + c.currentRX
	if currRX == 0 {
		return 0
	}
	currentRatio := int64(float64(c.currentTX) / float64(currRX))
	if constants.DEBUG {
		logs.Logger.Println("current total ratio: ", currentRatio)
	}
	neededBytes := (c.ratio - currentRatio) * c.currentRX
	if neededBytes < 0 {
		return 0
	}
	if neededBytes == 0 {
		return common.MBtoBytes(1)
	}
	return neededBytes
}

func (c *Calculator) GetLocalNeededWrite(rx, tx int64) (int64, int64) {
	if rx == 0 || tx == 0 {
		return 0, 0
	}
	currentRatio := int64(float64(tx) / float64(rx))
	if constants.DEBUG {
		logs.Logger.Println("current local ratio: ", currentRatio)
	}
	neededBytes := (c.ratio - currentRatio) * rx
	if neededBytes < 0 {
		return 0, currentRatio
	}
	if neededBytes == 0 {
		return common.MBtoBytes(1), currentRatio
	}
	return neededBytes, currentRatio
}

package internal

import (
	"FakeTCPUploader/common"
	"sync"
)

type Calculator struct {
	offset    int64
	maxRate   int64
	currentRX int64
	currentTX int64
	mu        sync.RWMutex
	ratio     int64
}

func NewCalculator(ratio int64, offset int64, maxRate int64) *Calculator {
	return &Calculator{
		offset:  offset,
		maxRate: maxRate,
		ratio:   ratio,
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
	neededBytes := (c.ratio - currentRatio) * c.currentRX
	if neededBytes < 0 {
		return 0
	}
	if neededBytes == 0 {
		return common.MBtoBytes(1)
	}
	return min(neededBytes, c.maxRate)
}

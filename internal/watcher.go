package internal

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/logs"
	"sync"
	"time"
)

type RateWatcher struct {
	addresses  []string
	foreign    []string
	banMap     map[string]time.Time
	limitCoef  int
	banMinutes int
	counter    int
	mu         sync.Mutex
}

func NewRateWatcher(addresses []string, foreign []string, limitCoef int, banMinutes int) *RateWatcher {
	watcher := &RateWatcher{
		addresses:  addresses,
		foreign:    foreign,
		limitCoef:  limitCoef,
		banMinutes: banMinutes,
		banMap:     make(map[string]time.Time),
	}

	go watcher.watchAddresses()

	return watcher
}

func (r *RateWatcher) watchAddresses() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			r.reviseMap()
			r.mu.Unlock()
		}
	}
}

func (r *RateWatcher) reviseMap() {
	for addr, banTime := range r.banMap {
		if int(time.Since(banTime).Minutes()) > r.banMinutes {
			delete(r.banMap, addr)
		}
	}
}

func (r *RateWatcher) GetAddr() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	mp := r.addresses
	if r.isDomesticExhausted() {
		mp = append(mp, r.foreign...)
	}
	for {
		addr := mp[r.counter]
		r.counter++
		r.counter %= len(mp)
		if _, ok := r.banMap[addr]; ok {
			continue
		}
		return addr
	}
}

func (r *RateWatcher) isDomesticExhausted() bool {
	cnt := 0
	for _, addr := range r.addresses {
		if _, ok := r.banMap[addr]; ok {
			cnt++
		}
	}
	if cnt > len(r.addresses)/2 {
		return true
	}
	return false
}

func (r *RateWatcher) WatchQuality(addr string, wantedRate int64, readBytes int64, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	currRate := int64(float64(readBytes) / d.Seconds())
	if currRate == 0 || int(float64(wantedRate)/float64(currRate)) >= r.limitCoef {
		r.banMap[addr] = time.Now()
		logs.Logger.Printf("address %v banned, wanted rate %v, current rate %v, took %v", addr, common.BytesToMB(wantedRate), common.BytesToMB(currRate), common.FormatFloat64(d.Seconds(), 2))
	} else {
		if constants.DEBUG {
			logs.Logger.Printf("addr %v wanted %v got %v, took %v", addr, common.BytesToMB(wantedRate), common.BytesToMB(currRate), common.FormatFloat64(d.Seconds(), 2))
		}
	}
}

func (r *RateWatcher) GetActiveAddressesCount() int {
	if r.isDomesticExhausted() {
		return len(r.addresses) + len(r.foreign) - len(r.banMap)
	}
	return len(r.addresses) - len(r.banMap)
}

package internal

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/logs"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type NetworkWatcher struct {
	interfaceName  string
	snapShotTime   time.Time
	snapShotRXDiff int64
	snapShotTXDiff int64
	snapAccuData   int64
}

func NewNetworkWatcher(interfaceName string, snapInterval int) *NetworkWatcher {
	nw := &NetworkWatcher{
		interfaceName: interfaceName,
	}

	go nw.snapShotter(snapInterval)

	return nw
}

func (n *NetworkWatcher) snapShotter(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	n.takeSnapShot()
	for {
		select {
		case <-ticker.C:
			n.takeSnapShot()
		}
	}
}

func (n *NetworkWatcher) takeSnapShot() {
	if constants.DEBUG {
		logs.Logger.Println("taking snapshot...")
	}

	dl, e := n.GetDownloadedBytes()
	up, e2 := n.GetUploadedBytes()
	if e != nil || e2 != nil {
		logs.Logger.Println("failed to get snapshot: ", common.NotNil(e, e2))
		return
	}
	n.calcAccuData(n.snapShotRXDiff, n.snapShotTXDiff)

	n.snapShotTime = time.Now()
	n.snapShotRXDiff = dl - n.snapShotRXDiff
	n.snapShotTXDiff = up - n.snapShotTXDiff
}

func (n *NetworkWatcher) calcAccuData(currRX, currTX int64) {
	diff := (currTX - 10*currRX) / 10
	n.snapAccuData -= diff
	if constants.DEBUG {
		logs.Logger.Printf("currTX %v, currRX %v, last accu %v, curr accu %v", currTX, currRX, n.snapAccuData+diff, n.snapAccuData)
	}
	n.snapAccuData = max(0, n.snapAccuData)
}

func (n *NetworkWatcher) GetSnapShotData() (rx int64, tx int64) {
	return n.snapShotRXDiff + n.snapAccuData, n.snapShotTXDiff
}

func (n *NetworkWatcher) GetDownloadedBytes() (int64, error) {
	f, e := os.Open("/sys/class/net/" + n.interfaceName + "/statistics/rx_bytes")
	if e != nil {
		return 0, e
	}
	defer f.Close()
	all, e := io.ReadAll(f)
	if e != nil {
		return 0, e
	}
	rx := common.MustVal(strconv.ParseInt(strings.ReplaceAll(string(all), "\n", ""), 10, 64))

	return rx, nil
}

func (n *NetworkWatcher) GetUploadedBytes() (int64, error) {
	f, e := os.Open("/sys/class/net/" + n.interfaceName + "/statistics/tx_bytes")
	if e != nil {
		return 0, e
	}
	defer f.Close()
	all, e := io.ReadAll(f)
	if e != nil {
		return 0, e
	}
	tx := common.MustVal(strconv.ParseInt(strings.ReplaceAll(string(all), "\n", ""), 10, 64))

	return tx, nil
}

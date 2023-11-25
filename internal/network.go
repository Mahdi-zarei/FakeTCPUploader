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

	n.snapShotTime = time.Now()
	n.snapShotRXDiff = dl - n.snapShotRXDiff
	n.snapShotTXDiff = up - n.snapShotTXDiff
}

func (n *NetworkWatcher) GetSnapShotData() (rx int64, tx int64) {
	return n.snapShotRXDiff, n.snapShotTXDiff
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

	if constants.DEBUG {
		logs.Logger.Println("current dl ", rx)
	}

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

	if constants.DEBUG {
		logs.Logger.Println("current up ", tx)
	}

	return tx, nil
}

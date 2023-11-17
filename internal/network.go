package internal

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/logs"
	"io"
	"os"
	"strconv"
	"strings"
)

type NetworkWatcher struct {
	interfaceName string
}

func NewNetworkWatcher(interfaceName string) *NetworkWatcher {
	return &NetworkWatcher{
		interfaceName: interfaceName,
	}
}

func (n *NetworkWatcher) GetDownloadedBytes() (int64, error) {
	f, e := os.Open("/sys/class/net/ens3/statistics/rx_bytes")
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
	f, e := os.Open("/sys/class/net/ens3/statistics/tx_bytes")
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

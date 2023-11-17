package internal

import (
	"FakeTCPUploader/constants"
	"FakeTCPUploader/logs"
	"errors"
	"github.com/akhenakh/statgo"
)

type NetworkWatcher struct {
	interfaceName string
	stats         *statgo.Stat
}

func NewNetworkWatcher(interfaceName string) *NetworkWatcher {
	return &NetworkWatcher{
		interfaceName: interfaceName,
		stats:         statgo.NewStat(),
	}
}

func (n *NetworkWatcher) GetDownloadedBytes() (int64, error) {
	stats := n.getNetStats()
	if stats == nil {
		return 0, errors.New("no interface found with the given name " + n.interfaceName)
	}
	if constants.DEBUG {
		logs.Logger.Println("current dl ", stats.RX)
	}

	return int64(stats.RX), nil
}

func (n *NetworkWatcher) GetUploadedBytes() (int64, error) {
	stats := n.getNetStats()
	if stats == nil {
		return 0, errors.New("no interface found with the given name " + n.interfaceName)
	}
	if constants.DEBUG {
		logs.Logger.Println("current up ", stats.TX)
	}

	return int64(stats.TX), nil
}

func (n *NetworkWatcher) getNetStats() *statgo.NetIOStats {
	allStats := n.stats.NetIOStats()
	if constants.DEBUG {
		for _, s := range allStats {
			logs.Logger.Println(*s)
		}
	}
	for _, val := range allStats {
		if val.IntName == n.interfaceName {
			return val
		}
	}

	return nil
}

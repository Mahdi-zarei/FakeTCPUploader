package main

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/internal"
	"FakeTCPUploader/logs"
	"flag"
	"log"
	"time"
)

var addresses = []string{
	"http://ookla-speedtest.shatel.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest1.pishgaman.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://sp2.petiak.com.prod.hosts.ooklaserver.net:8080/upload",
	"http://sp1.hiweb.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://rhaspd2.mci.ir:8080/upload",
	"http://ookla-tehran.tci.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest.techno2000.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://sptest.hostiran.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest1.irancell.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://turbo.nakhl.net.prod.hosts.ooklaserver.net:8080/upload",
}

func main() {
	_ratio := flag.Int("ratio", 10, "ratio of upload to download")
	_maxSpeed := flag.Int64("maxSpeed", 368, "max upload speed")
	_chunkSize := flag.Int64("chunkSize", common.MBtoBytes(32), "size of uploaded data in each post request")
	_interfaceName := flag.String("interface", "ens3", "name of interface to monitor")
	_offset := flag.Int64("offset", 0, "offset for download in GB")
	_debug := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()
	logs.Logger = log.Default()
	constants.DEBUG = *_debug
	ratio := *_ratio
	maxSpeed := common.MBtoBytes(*_maxSpeed)
	chunkSize := *_chunkSize
	interfaceName := *_interfaceName
	offset := *_offset

	if constants.DEBUG {
		logs.Logger.Printf("starting with ratio %v, maxSpeed %v, chunkSize %v, interfaceName %v, offset %v",
			ratio,
			maxSpeed,
			chunkSize,
			interfaceName,
			offset)
	}

	calulator := internal.NewCalculator(int64(ratio), offset, maxSpeed)
	networkWatcher := internal.NewNetworkWatcher(interfaceName)
	uploader := internal.NewUploader(chunkSize, addresses)

	for {
		calulator.RegisterNew(common.MustVal(networkWatcher.GetDownloadedBytes()), common.MustVal(networkWatcher.GetUploadedBytes()))
		needed := calulator.GetNeededWrite()
		if needed == 0 {
			if constants.DEBUG {
				logs.Logger.Println("no needed data")
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		writeCount := (needed + chunkSize) / chunkSize
		writeCount = min(writeCount, int64(len(addresses)*10))
		uploader.SendParallel(int(writeCount))
		time.Sleep(1 * time.Second)
	}
}

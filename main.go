package main

import (
	"FakeTCPUploader/common"
	"FakeTCPUploader/constants"
	"FakeTCPUploader/internal"
	"FakeTCPUploader/logs"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var addresses = []string{
	"http://ookla-speedtest.shatel.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest1.pishgaman.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://sp2.petiak.com.prod.hosts.ooklaserver.net:8080/upload",
	"http://sp1.hiweb.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://rhaspd2.mci.ir:8080/upload",
	"http://ookla-tehran.tci.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest1.irancell.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://turbo.nakhl.net.prod.hosts.ooklaserver.net:8080/upload",
}

func main() {
	go func() {
		http.ListenAndServe(":3436", nil)
	}()
	_ratio := flag.Int("ratio", 10, "ratio of upload to download")
	_maxSpeed := flag.Int64("maxSpeed", 32, "max upload speed in MB")
	_chunkSize := flag.Int64("chunkSize", 64, "size of uploaded data in MB in each post request")
	_interfaceName := flag.String("interface", "ens3", "name of interface to monitor")
	_offset := flag.Int64("offset", 0, "offset for download in GB")
	_sleeptime := flag.Int("sleepTime", 1000, "sleep time between checker loops in ms")
	_extraCount := flag.Int("extra", 1, "extra chunks uploaded per checker loop when ratio is already satisfied")
	_debug := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()
	logs.Logger = log.Default()
	constants.DEBUG = *_debug
	ratio := *_ratio
	maxSpeed := common.MBtoBytes(*_maxSpeed)
	chunkSize := common.MBtoBytes(*_chunkSize)
	interfaceName := *_interfaceName
	offset := *_offset
	sleeptime := *_sleeptime
	extraCount := *_extraCount

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
				logs.Logger.Println("no needed data, going on with delay and half speed")
			}
			time.Sleep(200 * time.Millisecond)
			uploader.SendParallel(1, maxSpeed/2)
			continue
		}
		writeCount := (needed + int64(extraCount)*chunkSize) / chunkSize
		writeCount = min(writeCount, int64(len(addresses)*4))
		uploader.SendParallel(int(writeCount), maxSpeed)
		time.Sleep(time.Duration(sleeptime) * time.Millisecond)
	}
}

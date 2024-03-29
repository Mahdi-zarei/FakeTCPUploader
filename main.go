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
	"http://speedtest1.pishgaman.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://sp2.petiak.com.prod.hosts.ooklaserver.net:8080/upload",
	"http://sp1.hiweb.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://rhaspd2.mci.ir:8080/upload",
	"http://ookla-tehran.tci.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest1.irancell.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://sptest.hostiran.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest.techno2000.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://turbo.nakhl.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest.systec-co.com.prod.hosts.ooklaserver.net:8080/upload",
	"http://sp2.petiak.com:8080/upload",
	"http://speed-test.respina.net:8080/upload",
	"http://punak.nakhldns.ir.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest.asiatech.com:8080/upload",
	"http://speedtest.abramad.com:8080/upload",
	"http://sp.zi-tel.com:8080/upload",
	"http://speedtestER-ookla.tcfars.ir:8080/upload",
	"http://speedtest.hexaserver.cloud:8080/upload",
	"http://teh.ticco.ae:8080/upload",
	"http://speedtest.electrotm.org:8080/upload",
	"http://speedtest.hostida.com:8080/upload",
	"http://tehranwest1.irancell.ir:8080/upload",
	"http://szaspd2.mci.ir:8080/upload",
	"http://speedtest-ookla.tcfars.ir:8080/upload",
	"http://es.ticco.ae:8080/upload",
	"http://mashhad1.irancell.ir:8080/upload",
	"http://mshspd2.mci.ir:8080/upload",
	"http://tzbspd2.mci.ir:8080/upload",
	"http://tabriz1.irancell.ir:8080/upload",
}

var foreign = []string{
	"http://speedtest-paris.whiteprovider.net:8080/upload",
	"http://montsouris3.speedtest.orange.fr.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest-prs.vts.bf.prod.hosts.ooklaserver.net:8080/upload",
	"http://lg-par2.fdcservers.net:8080/upload",
	"http://sp1.asthriona.com.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest.lekloud.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest.sewan.fr.prod.hosts.ooklaserver.net:8080/upload",
	"http://speedtest-ookla-par.as62000.net.prod.hosts.ooklaserver.net:8080/upload",
	"http://perf.keyyo.net:8080/upload",
	"http://speedtest.milkywan.fr.prod.hosts.ooklaserver.net:8080/upload",
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
	_snapInterval := flag.Int("snapInterval", 10, "interval between snapshots in minute")
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
	snapInterval := *_snapInterval

	if constants.DEBUG {
		logs.Logger.Printf("starting with ratio %v, maxSpeed %v, chunkSize %v, interfaceName %v, offset %v",
			ratio,
			maxSpeed,
			chunkSize,
			interfaceName,
			offset)
	}

	rateWatcher := internal.NewRateWatcher(addresses, foreign, 2, 10)
	calulator := internal.NewCalculator(int64(ratio), offset, maxSpeed)
	networkWatcher := internal.NewNetworkWatcher(interfaceName, snapInterval)
	uploader := internal.NewUploader(chunkSize, 2, rateWatcher)

	for {
		calulator.RegisterNew(common.MustVal(networkWatcher.GetDownloadedBytes()), common.MustVal(networkWatcher.GetUploadedBytes()))
		needed := calulator.GetNeededWrite()
		localNeeded, localRatio := calulator.GetLocalNeededWrite(networkWatcher.GetSnapShotData())
		needed = max(needed, localNeeded)
		if needed == 0 {
			if constants.DEBUG {
				logs.Logger.Println("no needed data, going on with delay and 1/8 speed")
			}
			time.Sleep(500 * time.Millisecond)
			uploader.SendParallel(1, maxSpeed/8)
			continue
		}
		speed := maxSpeed
		if int(localRatio) <= ratio/2 {
			speed *= 4
		}
		writeCount := (needed + int64(extraCount)*chunkSize) / chunkSize
		writeCount = max(min(min(writeCount, int64(rateWatcher.GetActiveAddressesCount()*4)), 4), 4)
		uploader.SendParallel(int(writeCount), speed)
		time.Sleep(time.Duration(sleeptime) * time.Millisecond)
	}
}

package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vmselect"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vmstorage"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/buildinfo"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/envflag"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/fs"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/httpserver"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/procutil"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/storage"
)

var (
	// HTTP 端口定义
	httpListenAddr    = flag.String("httpListenAddr", ":8428", "TCP address to listen for http connections")
	minScrapeInterval = flag.Duration("dedup.minScrapeInterval", 0, "Remove superflouos samples from time series if they are located closer to each other than this duration. "+
		"This may be useful for reducing overhead when multiple identically configured Prometheus instances write data to the same VictoriaMetrics. "+
		"Deduplication is disabled if the -dedup.minScrapeInterval is 0")
)

func main() {
	envflag.Parse()
	buildinfo.Init()
	// 日志服务初始化
	logger.Init()
	logger.Infof("starting VictoriaMetrics at %q...", *httpListenAddr)
	startTime := time.Now()
	storage.SetMinScrapeIntervalForDeduplication(*minScrapeInterval)
	// 初始化存储服务
	vmstorage.Init()
	// 初始化查询服务,初始化查询结果缓存
	vmselect.Init()
	// 数据写入服务初始化,这块主要是启动Http服务，并注册处理函数
	vminsert.Init()

	// 自己的服务指标上报
	startSelfScraper()

	// 启动HTTP server，并注册好处理器
	// 这里用的是原生的HTTP 处理器
	// 劣势就是整个服务启动了太多的server，每个服务有自己的端口，在我们配置域名接入的时候，我们需要分别为每个服务配置自己的端口
	go httpserver.Serve(*httpListenAddr, requestHandler)
	logger.Infof("started VictoriaMetrics in %.3f seconds", time.Since(startTime).Seconds())

	// 等待停止信号
	sig := procutil.WaitForSigterm()
	logger.Infof("received signal %s", sig)

	// 收到停止信号以后，清理资源，释放空间
	stopSelfScraper()

	logger.Infof("gracefully shutting down webservice at %q", *httpListenAddr)
	startTime = time.Now()
	if err := httpserver.Stop(*httpListenAddr); err != nil {
		logger.Fatalf("cannot stop the webservice: %s", err)
	}
	vminsert.Stop()
	logger.Infof("successfully shut down the webservice in %.3f seconds", time.Since(startTime).Seconds())

	vmstorage.Stop()
	vmselect.Stop()

	fs.MustStopDirRemover()

	logger.Infof("the VictoriaMetrics has been stopped in %.3f seconds", time.Since(startTime).Seconds())
}

// HTTP 请求处理器注册
func requestHandler(w http.ResponseWriter, r *http.Request) bool {
	// 数据写入？
	if vminsert.RequestHandler(w, r) {
		return true
	}
	// 数据查询
	if vmselect.RequestHandler(w, r) {
		return true
	}
	// 这个是什么？存储相关的嘛？
	if vmstorage.RequestHandler(w, r) {
		return true
	}
	return false
}

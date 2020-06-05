package writeconcurrencylimiter

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/httpserver"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/timerpool"
	"github.com/VictoriaMetrics/metrics"
)

var (
	maxConcurrentInserts = flag.Int("maxConcurrentInserts", runtime.GOMAXPROCS(-1)*4, "The maximum number of concurrent inserts; see also -insert.maxQueueDuration")
	maxQueueDuration     = flag.Duration("insert.maxQueueDuration", time.Minute, "The maximum duration for waiting in the queue for insert requests due to -maxConcurrentInserts")
)

// ch is the channel for limiting concurrent calls to Do.
var ch chan struct{}

// Init initializes concurrencylimiter.
//
// Init must be called after flag.Parse call.
func Init() {
	ch = make(chan struct{}, *maxConcurrentInserts)
}

// Do calls f with the limited concurrency.
func Do(f func() error) error {
	// Limit the number of conurrent f calls in order to prevent from excess
	// memory usage and CPU trashing.
	// Pool是一个阻塞队列，只有当能丢进去东西的时候，才认为可以抢到锁
	select {
	case ch <- struct{}{}:  // 往队列里面占队，拿到执行请求的权限
		err := f()         // 执行对应的请求处理回调函数
		<-ch                // 处理完成后撤销资源的占用
		return err          // 返回处理的结果
	default:
		// 没拿到资源，继续往下
	}

	// All the workers are busy.
	// Sleep for up to *maxQueueDuration.
	concurrencyLimitReached.Inc()
	// 从池里面取一个计时器，请求堵塞等待
	t := timerpool.Get(*maxQueueDuration)
	select {
	case ch <- struct{}{}:
		timerpool.Put(t)
		err := f()
		<-ch
		return err
	case <-t.C:
		timerpool.Put(t)
		concurrencyLimitTimeout.Inc()
		return &httpserver.ErrorWithStatusCode{
			Err: fmt.Errorf("cannot handle more than %d concurrent inserts during %s; possible solutions: "+
				"increase `-insert.maxQueueDuration`, increase `-maxConcurrentInserts`, increase server capacity", *maxConcurrentInserts, *maxQueueDuration),
			StatusCode: http.StatusServiceUnavailable,
		}
	}
}

var (
	concurrencyLimitReached = metrics.NewCounter(`vm_concurrent_insert_limit_reached_total`)
	concurrencyLimitTimeout = metrics.NewCounter(`vm_concurrent_insert_limit_timeout_total`)

	_ = metrics.NewGauge(`vm_concurrent_insert_capacity`, func() float64 {
		return float64(cap(ch))
	})
	_ = metrics.NewGauge(`vm_concurrent_insert_current`, func() float64 {
		return float64(len(ch))
	})
)

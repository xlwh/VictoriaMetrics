package procutil

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitForSigterm waits for either SIGTERM or SIGINT
//
// Returns the caught signal.
func WaitForSigterm() os.Signal {
	ch := make(chan os.Signal, 1)
	// 巧妙的利用了chan是阻塞的原理,ch只有接收到signal发送的term信号，函数才会返回
	// 然后说明我们的进程需要退出了，这时候进行服务退出即可
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	return <-ch
}

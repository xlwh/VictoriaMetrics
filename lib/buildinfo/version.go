package buildinfo

import (
	"flag"
	"fmt"
	"os"
)

var version = flag.Bool("version", false, "Show VictoriaMetrics version")

// Version must be set via -ldflags '-X'
// go build -ldflags "-w -s -X main.Version=${VERSION} -X main.Build=${BUILD}"
// 在构建时期注入程序的版本号
var Version string

// Init must be called after flag.Parse call.
func Init() {
	if *version {
		printVersion()
		os.Exit(0)
	}
}

// 生成程序的flag
func init() {
	oldUsage := flag.Usage
	flag.Usage = func() {
		printVersion()
		oldUsage()
	}
}

func printVersion() {
	fmt.Fprintf(flag.CommandLine.Output(), "%s\n", Version)
}

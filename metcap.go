package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/blufor/metcap"
)

var (
	Version string
	Build   string
)

func main() {

	// cmdline options
	cfg := flag.String("config", "/etc/metcap/main.conf", "Path to config file")
	cores := flag.Int("cores", runtime.NumCPU(), "Number of cores to use")
	version := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *version {
		fmt.Println("MetCap version " + Version + " (build " + Build + ")")
	} else {
		runtime.GOMAXPROCS(*cores)
		mc := metcap.NewEngine(*cfg)
		mc.Run()
	}
}

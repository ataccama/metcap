package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	// "runtime/pprof"

	"github.com/blufor/metcap"
	"github.com/pkg/profile"
)

var (
	// Version is used to specify version number on build
	Version string
	// Build is used to specify commit sub-hash on build
	Build   string
)

func main() {

	exitCode := make(chan int, 1)
	var p interface{ Stop() }

	// cmdline options
	cfg := flag.String("config", "/etc/metcap/main.conf", "Path to config file")
	cores := flag.Int("cores", runtime.NumCPU(), "Number of cores to use")
	prof := flag.String("prof", "", "Run with profiling enabled, can be either one of: cpu,mem,blk,trace")
	version := flag.Bool("version", false, "Show version")

	flag.Parse()

	if *version {
		fmt.Println("MetCap version " + Version + " (build " + Build + ")")
		return
	}

	switch *prof {
	case "cpu":
		p = profile.Start(profile.NoShutdownHook, profile.CPUProfile)
	case "mem":
		p = profile.Start(profile.NoShutdownHook, profile.MemProfile)
	case "blk":
		p = profile.Start(profile.NoShutdownHook, profile.BlockProfile)
	default:
	}

	runtime.GOMAXPROCS(*cores)

	mc := metcap.NewEngine(*cfg, exitCode)
	mc.Run()

	if *prof != "" {
		p.Stop()
	}

	os.Exit(<-exitCode)
}

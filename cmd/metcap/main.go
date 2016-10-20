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
	Build string
)

func main() {
	var p interface {
		Stop()
	}
	cfg := flag.String("config", "/etc/metcap/main.conf", "Path to config file")
	cores := flag.Int("cores", runtime.NumCPU(), "Number of cores to use")
	prof := flag.String("prof", "", "Run with profiling enabled, can be either one of: cpu,mem,blk,trace")
	version := flag.Bool("version", false, "Show version")
	flag.Parse()
	if *version {
		fmt.Printf("MetCap version %s (build %s)\n", Version, Build)
		return
	}
	config := metcap.ReadConfig(cfg)
	switch *prof {
	case "":
	case "cpu":
		p = profile.Start(profile.NoShutdownHook, profile.CPUProfile)
	case "mem":
		p = profile.Start(profile.NoShutdownHook, profile.MemProfile)
	case "blk":
		p = profile.Start(profile.NoShutdownHook, profile.BlockProfile)
	case "trace":
		p = profile.Start(profile.NoShutdownHook, profile.TraceProfile)
	default:
		fmt.Printf("ERROR: Unknown profiling type '%s'. Use one of: cpu,mem,blk,trace\n", *prof)
		os.Exit(1)
	}
	runtime.GOMAXPROCS(*cores)
	mc, exitCode := metcap.NewEngine(config)
	mc.Run()
	codeNum := <-exitCode
	if *prof != "" {
		p.Stop()
	}
	os.Exit(codeNum)
}

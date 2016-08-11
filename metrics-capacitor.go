package main

import (
  "fmt"
  "flag"
  "runtime"
  "github.com/metrics-capacitor/metrics-capacitor"
)

var (
  Version string
  Build string
)

func main() {

  // cmdline options
  cfg := flag.String("config", "/etc/metrics-capacitor/main.conf", "Path to config file")
  daemon := flag.Bool("daemonize", false, "Run on background")
  cores := flag.Int("cores", runtime.NumCPU(), "Number of cores to use")
  version := flag.Bool("version", false, "Show version")
  flag.Parse()

  if *version {
    fmt.Println("Metrics Capacitor version " + Version + " (build " + Build + ")")
  } else {
    runtime.GOMAXPROCS(*cores)
    mc := metcap.NewEngine(*cfg, *daemon)
    mc.Run()
  }
}

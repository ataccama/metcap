package main

import (
  "flag"
  "runtime"
  "github.com/metrics-capacitor/metrics-capacitor"
)

func main() {
  // cmdline options
  cfg := flag.String("config", "/etc/metrics-capacitor/main.conf", "Path to config file")
  daemon := flag.Bool("daemonize", false, "Run on background")
  cores := flag.Int("cores", runtime.NumCPU(), "Number of cores to use")
  flag.Parse()

  runtime.GOMAXPROCS(*cores)

  // metrics-capacitor Engine
  mc := metcap.NewEngine(*cfg, *daemon)
  mc.Run()
}

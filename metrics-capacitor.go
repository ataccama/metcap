package main

import (
  "flag"
  "github.com/metrics-capacitor/metrics-capacitor"
)

func main() {
  cfg := flag.String("config", "/etc/metrics-capacitor/main.conf", "Path to config file")
  daemon := flag.Bool("daemonize", false, "Run on background")
  flag.Parse()
  mc := metcap.NewEngine(*cfg, *daemon)
  mc.Run()
}

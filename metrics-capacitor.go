package main

import (
  "flag"
  "github.com/metrics-capacitor/metrics-capacitor"
)

func main() {
  configfile := flag.String("configfile", "/etc/metrics-capacitor/main.conf", "Path to config file")
  engine := NewEngine(&configfile)
  engine.Run()
}

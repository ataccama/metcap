package main

import (
  "fmt"
  // "os"
  // "time"
  "encoding/json"
  // "flag"
  // "net/http"
  // "log/syslog"
  "github.com/jrallison/go-workers"
)

type Config struct {
  redis_conn    string
  redis_queue   string
  influx_pool   string
  influx_db     string
  concurrency   uint32
  debug, syslog bool
}

//
// Measurement objects
//
type Measurements []struct {
  Metric        string              `json:"metric"`
  Timestamp     uint64              `json:"timestamp"`
  Value         float64             `json:"value"`
  Tags
}

// Variable tags
type Tags struct {}

// Ingest JSON from Sensu
func (m Measurements) DecodeSensuJSON(data []byte) error {
  var input Measurements
  if err := json.Unmarshal(data, &input); err != nil {
    return fmt.Errorf("Measurements decode failed: %v", err)
  }
  m = input
  return nil
}

// Returns InfluxDB Line Protocol formatted data
// func (m *Measurements) EncodeInfluxLineProto() (string, error) {}


//
// WORKER FUNC
//
func flux(message *workers.Msg) {
  input := message.Args()
  // jid := message.Jid()

  fmt.Println(input)

  var m Measurements
  // m.DecodeSensuJSON(input)

  fmt.Println(m)
}

//
// MAIN
//
func init()  {
  // read config file
  // flag.Parse() ...
  // initialize logging
}

func main()  {
  workers.Configure(map[string]string{
    "server":   "localhost:6379",
    "database": "0",
    "pool":     "100",
    "process":  "1",
  })
  workers.Process("metrics", flux, 1000)
  go workers.StatsServer(8090)
  workers.Run()
}


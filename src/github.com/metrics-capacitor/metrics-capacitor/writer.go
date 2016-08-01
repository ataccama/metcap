package metcap

import (
  "fmt"
  "sync"
  // "gopkg.in/olivere/elastic.v3"
)

type Writer struct {
  Queue chan Metric
}

func RunWriter(workers *sync.WaitGroup) {
  workers.Add(1)
  fmt.Println("Starting writer...")
  defer fmt.Println("Stopping writer...")
  // code
  workers.Done()
}

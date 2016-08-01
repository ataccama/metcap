package metcap

import (
  "fmt"
  "sync"
)

type Listener struct {

}

func RunListener(workers *sync.WaitGroup)  {
  workers.Add(1)
  fmt.Println("Starting listener...")
  defer fmt.Println("Stopping listener...")
  // code
  workers.Done()
}

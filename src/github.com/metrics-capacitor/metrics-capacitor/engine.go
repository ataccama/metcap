package metcap

import (
  "os"
  "syscall"
  "sync"
  "fmt"
)

type Engine struct {
  Config      Config
  Daemon      *bool
  Workers     *sync.WaitGroup
  SignalChan  chan os.Signal
  ExitChan    chan int
}

func NewEngine(configfile string, daemon bool) Engine {
  return Engine{
    Config:     ReadConfig(&configfile),
    Daemon:     &daemon,
    Workers:    &sync.WaitGroup{},
    SignalChan: make(chan os.Signal, 1),
    ExitChan:   make(chan int)}
}

func (e *Engine) Run() {
  fmt.Println("INFO:  MetricsCapacitor Engine is starting")
  // initialize buffer
  b := NewBuffer(&e.Config.Buffer)
  fmt.Println("INFO:  buffer initialized")

  // initialize & start writer
  w := NewWriter(&e.Config.Writer, b, e.Workers)
  go w.Run()
  fmt.Println("INFO:  writer initialized & started")

  // initialize & start listeners
  if len(e.Config.Listener) > 0 {
    fmt.Println("INFO:  initilizing listeners...")
    for l_name, cfg := range e.Config.Listener {
      l := NewListener(l_name, cfg, b, e.Workers)
      go l.Run()
      fmt.Println("INFO:  listener '" + l_name + "' initialized")
    }
  }

  fmt.Println("INFO:  engine started :-)")

  // signal handling
  go func() {
    for {
      s := <-e.SignalChan
      switch s {
      case syscall.SIGINT:
        e.ExitChan <- 0
      case syscall.SIGTERM:
        e.ExitChan <- 0
      default:
        e.ExitChan <- 1
      }
    }
  }()

  // exit code semaphore
  exit := <-e.ExitChan

  // wait for all workers to finish
  e.Workers.Wait()

  // close buffer connection
  b.Close()

  // exit to the system :)
  os.Exit(exit)
}

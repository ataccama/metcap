package metcap

import (
  "os"
  "syscall"
  "sync"
)

type Engine struct {
  Config      Config
  Daemon      *bool
  Workers     *sync.WaitGroup
  SignalChan  chan os.Signal
  ExitChan    chan int
}

func NewEngine(configfile *string, daemon *bool) Engine {
  return Engine{
    Config:     ReadConfig(configfile),
    Daemon:     daemon,
    Workers:    &sync.WaitGroup{},
    SignalChan: make(chan os.Signal, 1),
    ExitChan:   make(chan int)}
}

func (e *Engine) Run() {
  // initialize buffer
  b := NewBuffer(&e.Config.Redis)

  // initialize & start writer
  w := NewWriter(&e.Config.Writer, b, e.Workers)
  go w.Run()

  // initialize & start listeners
  for name, cfg := range e.Config.Listener {
    l := NewListener(&name, &cfg, b, e.Workers)
    go l.Run()
  }

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

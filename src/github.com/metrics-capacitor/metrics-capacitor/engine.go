package metcap

import (
  "os"
  "syscall"
  "sync"
)

type Engine struct {
  Config      Config
  Workers     *sync.WaitGroup
  SignalChan  chan os.Signal
  ExitChan    chan int
}

func NewEngine(configfile *string) Engine {
  return Engine{
    Config:     ReadConfig(configfile),
    Workers:    &sync.WaitGroup{},
    SignalChan: make(chan os.Signal, 1),
    ExitChan:   make(chan int)}
}

func (e *Engine) Run() {
  buffer := NewBuffer(&e.Config)

  go RunWriter(e.Workers)
  go RunListener(e.Workers)

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
  exit := <-e.ExitChan
  e.Workers.Wait()
  buffer.Close()
  os.Exit(exit)
}

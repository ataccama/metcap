package metcap

import (
	"os"
	"os/signal"
	"syscall"
	"sync"
)

type Engine struct {
	Config     Config
	Daemon     *bool
	Workers    *sync.WaitGroup
	SignalChan chan os.Signal
	ExitChans  []*chan bool
}

func NewEngine(configfile string, daemon bool) Engine {
	return Engine{
		Config:     ReadConfig(&configfile),
		Daemon:     &daemon,
		Workers:    &sync.WaitGroup{},
		SignalChan: make(chan os.Signal, 1),
	}
}

func (e *Engine) Run() {
	log := NewLogger(&e.Config.Syslog, &e.Config.Debug)
	go log.Run()

	log.Info("[engine] Starting...")

	signal.Notify(e.SignalChan, syscall.SIGINT, syscall.SIGTERM)

	// initialize buffer
	buffer := NewBuffer(&e.Config.Buffer, log)

	// initialize & start writer
	if e.Config.Writer.Urls != nil {
		w_exit_chan := make(chan bool, 1)
		e.ExitChans = append(e.ExitChans, &w_exit_chan)
		writer := NewWriter(&e.Config.Writer, buffer, e.Workers, log, w_exit_chan)
		go writer.Start()
	}

	// initialize & start listeners
	if len(e.Config.Listener) > 0 {
		for l_name, cfg := range e.Config.Listener {
			l_exit_chan := make(chan bool, 1)
			e.ExitChans = append(e.ExitChans, &l_exit_chan)
			l := NewListener(l_name, cfg, buffer, e.Workers, log, l_exit_chan)
			go l.Start()
		}
	}

	log.Info("[engine] Started")

	for {
		sig := <-e.SignalChan
		switch {
		case sig == syscall.SIGINT || sig == syscall.SIGTERM:
			log.Info("[engine] Caught signal to shutdown...")
			log.Debug("[engine] Waiting for workers to stop")
			for _, c := range e.ExitChans {
				*c <- true
			}
			e.Workers.Wait()
			buffer.Close()
			log.Info("[engine] Exiting...")
			os.Exit(0)
		default:
		}
	}
}

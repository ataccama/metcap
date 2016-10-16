package metcap

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Engine struct {
	Config     Config
	Workers    *sync.WaitGroup
	ExitCode   chan int
	SignalChan chan os.Signal
}

func NewEngine(configfile string, exitChan chan int) Engine {
	return Engine{
		Config:     ReadConfig(&configfile),
		Workers:    &sync.WaitGroup{},
		ExitCode:   exitChan,
		SignalChan: make(chan os.Signal, 1),
	}
}

func (e *Engine) Run() {
	debugFlag := &Flag{new(sync.Mutex), e.Config.Debug}
	exitFlag := &Flag{new(sync.Mutex), false}
	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	}
	signal.Notify(e.SignalChan, signals...)

	logger := NewLogger(&e.Config.Syslog, debugFlag)
	go logger.Run()

	logger.Info("[engine] Starting...")

	var listenerEnabled, writerEnabled bool = false, false
	var transport Transport
	var listeners []*Listener
	var writers []*Writer

	if e.Config.Writer.URLs != nil {
		writerEnabled = true
	}
	if len(e.Config.Listener) > 0 {
		listenerEnabled = true
	}

	// initialize transport
	logger.Infof("[engine] Using '%s' transport", e.Config.Transport.Type)
	var err error
	switch e.Config.Transport.Type {
	case "channel":
		if listenerEnabled == false || writerEnabled == false {
			logger.Alert("[engine] Channel transport requires you to have both listener and writer enabled!")
			e.ExitCode <- 1
			return
		}
		transport = NewChannelTransport(&e.Config.Transport, logger)
	case "redis":
		transport, err = NewRedisTransport(&e.Config.Transport, listenerEnabled, writerEnabled, exitFlag, logger)
	case "amqp":
		transport, err = NewAMQPTransport(&e.Config.Transport, listenerEnabled, writerEnabled, exitFlag, logger)
	default:
		logger.Alertf("[engine] Transport '%s' not implemented", e.Config.Transport.Type)
		e.ExitCode <- 1
		return
	}
	if err != nil {
		logger.Alertf("[engine] Failed to set-up transport: %v", err)
		e.ExitCode <- 1
		return
	}

	// initialize & start writer
	if writerEnabled {
		writer, err := NewWriter(&e.Config.Writer, transport, e.Workers, logger, exitFlag)
		if err != nil {
			logger.Alert("[engine] Failed to initialize writer. Exiting")
			e.ExitCode <- 1
			return
		}
		writers = append(writers, &writer)
		go writer.Start()
	}

	// initialize & start listeners
	if listenerEnabled {
		for lName, cfg := range e.Config.Listener {
			listener, err := NewListener(lName, cfg, transport, e.Workers, logger, exitFlag)
			if err != nil {
				logger.Alertf("[engine] Failed to initialize listener '%s'", lName)
				continue
			}
			listeners = append(listeners, &listener)
			go listener.Start()
		}
	}

	// start transport
	transport.Start()

	stopReporter := make(chan struct{}, 1)
	// stats report goroutine
	go func() {
		// report func
		report := func() {
			for _, listener := range listeners {
				listener.LogReport()
			}
			transport.LogReport()
			for _, writer := range writers {
				writer.LogReport()
			}
		}
		// sleepTime between reports
		var sleepTime time.Duration
		if e.Config.ReportEvery.Duration > 0 {
			sleepTime = e.Config.ReportEvery.Duration
		} else {
			sleepTime = time.Duration(1 * time.Minute)
		}

		// ticker
		tick := make(chan struct{}, 1)
		go func() {
			for {
				tick <- struct{}{}
				time.Sleep(sleepTime)
			}
		}()

		// report select-loop
		for {
			select {
			case <-stopReporter:
				report()
				close(stopReporter)
				return
			case <-tick:
				report()
			}
		}
	}()

	// signal handler
	for {
		sig := <-e.SignalChan
		switch {
		case sig == syscall.SIGINT || sig == syscall.SIGTERM:
			if sig == syscall.SIGINT {
				logger.Info("[engine] Received SIGINT - shutting down")
			} else {
				logger.Info("[engine] Received SIGTERM - shutting down")
			}
			exitFlag.Raise()

			e.Workers.Wait()

			logger.Debug("[engine] Waiting for transport to terminate")
			transport.Stop()

			stopReporter <- struct{}{}
			time.Sleep(100 * time.Millisecond)
			<-stopReporter

			logger.Info("[engine] Exiting...")
			time.Sleep(100 * time.Millisecond)
			e.ExitCode <- 0
			return

		case sig == syscall.SIGUSR1:
			if debugFlag.Get() {
				logger.Info("[engine] Received SIGUSR1 - disabling DEBUG mode")
			} else {
				logger.Info("[engine] Received SIGUSR1 - enabling DEBUG mode")
			}
			debugFlag.Flip()

		case sig == syscall.SIGUSR2:
			logger.Info("[engine] Resetting counters")
			// do

		default:
			logger.Errorf("[engine] Unknown signal: %v", sig)
		}
	}
}

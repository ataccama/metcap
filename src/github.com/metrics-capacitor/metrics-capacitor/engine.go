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
	Daemon     *bool
	Workers    *sync.WaitGroup
	SignalChan chan os.Signal
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

	if e.Config.Writer.URLs != nil {
		writerEnabled = true
	}
	if len(e.Config.Listener) > 0 {
		listenerEnabled = true
	}

	// initialize transport
	logger.Infof("[engine] Using '%s' transport", e.Config.Transport.Type)
	switch e.Config.Transport.Type {
	case "channel":
		if listenerEnabled == false || writerEnabled == false {
			logger.Alert("[engine] Channel transport requires you to have both listener and writer enabled!")
			os.Exit(1)
		}
		transport = NewChannelTransport(&e.Config.Transport)
	case "redis":
		transport = NewRedisTransport(&e.Config.Transport, listenerEnabled, writerEnabled, exitFlag)
	case "amqp":
		var err error
		transport, err = NewAMQPTransport(&e.Config.Transport, listenerEnabled, writerEnabled, exitFlag, logger)
		if err != nil {
			logger.Alertf("[engine] Failed to set-up transport: %v", err)
			os.Exit(1)
		}
	default:
		logger.Alertf("[engine] Transport '%s' not implemented", e.Config.Transport.Type)
		os.Exit(1)
	}

	// initialize & start writer
	if writerEnabled {
		writer, err := NewWriter(&e.Config.Writer, transport, e.Workers, logger, exitFlag)
		if err != nil {
			logger.Alert("[engine] Failed to initialize writer. Exiting")
			os.Exit(1)
		}
		go writer.Start()
	}

	// initialize & start listeners
	if listenerEnabled {
		for lName, cfg := range e.Config.Listener {
			listener, err := NewListener(lName, cfg, transport, e.Workers, logger, exitFlag)
			if err != nil {
				logger.Alertf("[engine] Failed to initialize listener '%s'", lName)
			} else {
				go listener.Start()
			}
		}
	}

	// start transport
	transport.Start()

	// signal handler
	for {
		sig := <-e.SignalChan
		switch {
		case sig == syscall.SIGINT || sig == syscall.SIGTERM:
			logger.Info("[engine] Caught signal to shutdown...")
			exitFlag.Raise()
			logger.Debug("[engine] Waiting for workers to stop")
			e.Workers.Wait()
			logger.Debug("[engine] Waiting for transport to terminate")
			transport.Stop()
			logger.Info("[engine] Exiting...")
			time.Sleep(1 * time.Second)
			os.Exit(0)
		case sig == syscall.SIGUSR1:
			logger.Info("[engine] Enabling debug mode via signal")
			debugFlag.Raise()
		case sig == syscall.SIGUSR2:
			logger.Info("[engine] Disabling debug mode via signal")
			debugFlag.Lower()
		default:
			logger.Errorf("[engine] Unknown signal %v", sig)
		}
	}
}

type Flag struct {
	*sync.Mutex
	val bool
}

func (f *Flag) Get() bool {
	f.Lock()
	defer f.Unlock()
	return f.val
}

func (f *Flag) Raise() {
	f.Lock()
	f.val = true
	f.Unlock()
}

func (f *Flag) Lower() {
	f.Lock()
	f.val = false
	f.Unlock()
}

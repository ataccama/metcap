package metcap

import (
	// "bytes"
	"net"
	"strconv"
	"sync"
	"time"
)

type Listener struct {
	Name      string
	Socket    net.Listener
	Config    ListenerConfig
	ConnWg    sync.WaitGroup
	ModuleWg  *sync.WaitGroup
	Transport Transport
	Codec     Codec
	Logger    *Logger
	ExitFlag  *Flag
}

type ListenerStatus struct {
}

func NewListener(name string, c ListenerConfig, t Transport, moduleWg *sync.WaitGroup, logger *Logger, exitFlag *Flag) (Listener, error) {
	logger.Infof("[listener:%s] Starting [%s://0.0.0.0:%d/%s]", name, c.Protocol, c.Port, c.Codec)

	sock, err := net.Listen("tcp", ":"+strconv.Itoa(c.Port))
	if err != nil {
		logger.Alertf("[listener:%s] Couldn't start listener: %v", name, err)
		return Listener{}, err
	}

	var codec Codec

	switch c.Codec {
	case "graphite":
		logger.Debugf("[listener:%s] Detected graphite codec, loading mutator config", name)
		codec, err = NewGraphiteCodec(c.MutatorFile)
	case "influx":
		logger.Debugf("[listener:%s] Detected influx codec", name)
		// codec, err := NewInfluxCodec()
	}
	if err != nil {
		logger.Alertf("[listener:%s] Failed to initialize codec: %v", name, err)
		return Listener{}, err
	}

	// var wg sync.WaitGroup

	return Listener{
		Name:      name,
		Socket:    sock,
		Config:    c,
		ConnWg:    sync.WaitGroup{},
		ModuleWg:  moduleWg,
		Transport: t,
		Codec:     codec,
		Logger:    logger,
		ExitFlag:  exitFlag,
	}, nil
}

func (l *Listener) Start() {
	l.ModuleWg.Add(1)
	defer l.ModuleWg.Done()

	l.Logger.Infof("[listener:%s] Starting to accept connections", l.Name)

	connPipe := make(chan net.Conn)
	exitChan := make(chan bool, 1)

	// connection acceptor
	go func() {
		for {
			conn, err := l.Socket.Accept()
			switch {
			case err == nil && conn != nil: // connection
				l.Logger.Debugf("[listener:%s] Accepted connection from %s", l.Name, conn.RemoteAddr().String())
				connPipe <- conn
			case err != nil && conn != nil: // other error
				l.Logger.Errorf("[listener:%s] Can't accept connection from %s: %v", l.Name, conn.RemoteAddr().String(), err)
			case err != nil && conn == nil: // exiting
				return
			}
		}
	}()

	// shutdown handler
	go func() {
		for {
			if l.ExitFlag.Get() {
				l.Logger.Infof("[listener:%s] Stopping...", l.Name)
				l.Logger.Debugf("[listener:%s] Closing LISTEN socket", l.Name)
				l.Socket.Close()
				l.Logger.Infof("[listener:%s] Socket closed", l.Name)
				l.Logger.Debugf("[listener:%s] Processing remaining metrics", l.Name)
				exitChan <- true
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// listener processing loop
	for {
		select {
		case conn := <-connPipe:
			go l.handleConnection(conn)
		case <-exitChan:
			l.Logger.Debugf("[listener:%s] Remaining metrics processed", l.Name)
			l.ConnWg.Wait()
			l.Logger.Debugf("[listener:%s] Stopped", l.Name)
			return
		}
	}
}

func (l *Listener) handleConnection(conn net.Conn) {
	l.ConnWg.Add(1)
	defer conn.Close()
	defer l.ConnWg.Done()
	metrics, took, errs := l.Codec.Decode(conn)
	if len(errs) > 0 {

	}
	l.Logger.Debugf("[listener:%s] Decoded %d metrics, took %dms", l.Name, len(metrics), took)
	for _, metric := range metrics {
		l.Transport.ListenerChan() <- &metric
	}
	l.Logger.Debugf("[listener:%s] Closing connection to %s", l.Name, conn.RemoteAddr().String())
}

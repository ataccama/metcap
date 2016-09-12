package metcap

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type Listener struct {
	Name      string
	Socket    net.Listener
	Config    ListenerConfig
	DataWg    sync.WaitGroup
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
		codec, err = NewInfluxCodec()
	}
	if err != nil {
		logger.Alertf("[listener:%s] Failed to initialize codec: %v", name, err)
		return Listener{}, err
	}

	return Listener{
		Name:      name,
		Socket:    sock,
		Config:    c,
		DataWg:    sync.WaitGroup{},
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

	dataPipe := make(chan *bytes.Buffer)
	exitChan := make(chan bool, 1)

	// connection acceptor
	go func() {
		for {
			conn, err := l.Socket.Accept()
			switch {
			case err == nil && conn != nil: // connection
				l.Logger.Debugf("[listener:%s] Accepted connection from %s", l.Name, conn.RemoteAddr().String())
				iBuf := bufio.NewReader(conn)
				var oBuf bytes.Buffer
				_, err := io.Copy(&oBuf, iBuf)
				conn.Close()
				if err != nil {
					l.Logger.Alertf("[listener:%s] Error handling connection data: %v", err)
					continue
				}
				l.Logger.Debugf("[listener:%s] Closing connection to %s", l.Name, conn.RemoteAddr().String())
				dataPipe <- &oBuf
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
				l.DataWg.Wait()
				exitChan <- true
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// listener processing loop
	for {
		select {
		case data := <-dataPipe:
			go func(data *bytes.Buffer) {
				l.DataWg.Add(1)
				defer l.DataWg.Done()
				metrics, took, errs := l.Codec.Decode(bytes.NewReader(data.Bytes()))
				if len(errs) > 0 {
					l.Logger.Errorf("[listener:%s] Failed to decode %d metrics!", l.Name, len(errs))
					// log the metric raw data?
				}
				l.Logger.Debugf("[listener:%s] Decoded %d metrics, took %v", l.Name, len(metrics), took)
				for _, metric := range metrics {
					l.Transport.ListenerChan() <- &metric
				}
			}(data)
		case <-exitChan:
			l.Logger.Infof("[listener:%s] Remaining metrics processed", l.Name)
			l.DataWg.Wait()
			l.Logger.Infof("[listener:%s] Stopped", l.Name)
			return
		}
	}

}

package metcap

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type Listener struct {
	Name            string
	Socket          net.Listener
	Config          ListenerConfig
	ConnWg          sync.WaitGroup
	ModuleWg        *sync.WaitGroup
	Transport       Transport
	GraphiteMutator *[]string
	Logger          *Logger
	ExitFlag        *Flag
}

func NewListener(name string, c ListenerConfig, t Transport, module_wg *sync.WaitGroup, logger *Logger, exitFlag *Flag) (Listener, error) {
	logger.Infof("[listener:%s] Starting [%s://0.0.0.0:%d/%s]", name, c.Protocol, c.Port, c.Codec)

	sock, err := net.Listen("tcp", ":"+strconv.Itoa(c.Port))
	if err != nil {
		logger.Alertf("[listener:%s] Couldn't start listener: %v", name, err)
		return Listener{}, err
	}

	var mut []string
	if c.Codec == "graphite" {
		logger.Debugf("[listener:%s] Detected graphite codec, loading mutator config", name)
		mut_file, err := os.Open(c.MutatorFile)
		if err != nil {
			logger.Alertf("[listener:%s] Couldn't open mutator config: %v", name, err)
			return Listener{}, err
		} else {
			scn := bufio.NewScanner(mut_file)
			for scn.Scan() {
				mut = append(mut, scn.Text())
			}
			logger.Debugf("[listener:%s] Loaded mutator rules", name)
		}
	}

	var wg sync.WaitGroup

	return Listener{
		Name:            name,
		Socket:          sock,
		Config:          c,
		ConnWg:          wg,
		ModuleWg:        module_wg,
		Transport:       t,
		GraphiteMutator: &mut,
		Logger:          logger,
		ExitFlag:        exitFlag,
	}, nil
}

func (l *Listener) Start() {
	l.ModuleWg.Add(1)
	defer l.ModuleWg.Done()

	l.Logger.Infof("[listener:%s] Starting to accept connections", l.Name)

	connPipe := make(chan net.Conn, 100)
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
			} else {
				time.Sleep(10 * time.Millisecond)
			}
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
	scn := bufio.NewScanner(conn)
	for scn.Scan() {
		line := scn.Text()
		metric, err := NewMetricFromLine(line, l.Config.Codec, l.GraphiteMutator)
		if err == nil {
			if metric.OK {
				l.Transport.ListenerChan() <- &metric
			} else {
				l.Logger.Debugf("[listener:%s] Malformed line, skipping", l.Name)
			}
		} else {
			l.Logger.Errorf("[listener:%s] %v", l.Name, err)
		}
	}
	conn.Close()
	l.Logger.Debugf("[listener:%s] Closed connection from %s", l.Name, conn.RemoteAddr().String())
	l.ConnWg.Done()
}

package metcap

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"sync"
)

type Listener struct {
	Name            string
	Socket          net.Listener
	Config          ListenerConfig
	ConnWg          sync.WaitGroup
	ModuleWg				*sync.WaitGroup
	Buffer          *Buffer
	GraphiteMutator *[]string
	Logger          *Logger
	ExitChan				<-chan bool
}

func NewListener(name string, c ListenerConfig, b *Buffer, module_wg *sync.WaitGroup, logger *Logger, exit_chan <-chan bool) Listener {
	logger.Infof("[listener:%s] Starting [%s://0.0.0.0:%d/%s]", name, c.Protocol, c.Port, c.Codec)

	sock, err := net.Listen("tcp", ":"+strconv.Itoa(c.Port))
	if err != nil {
		logger.Alertf("[listener:%s] Couldn't start listener: %v", name, err)
	}

	var mut []string
	if c.Codec == "graphite" {
		logger.Debugf("[listener:%s] Detected graphite codec, loading mutator config", name)
		mut_file, err := os.Open(c.MutatorFile)
		if err != nil {
			logger.Alertf("[listener:%s] Couldn't open mutator config: %v", name, err)
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
		Name:            	name,
		Socket:          	sock,
		Config:          	c,
		ConnWg:          	wg,
		ModuleWg:        	module_wg,
		Buffer:          	b,
		GraphiteMutator: 	&mut,
		Logger:          	logger,
		ExitChan:					exit_chan,
	}
}

func (l *Listener) Start() {
	l.ModuleWg.Add(1)
	defer l.ModuleWg.Done()

	l.Logger.Infof("[listener:%s] Starting to accept connections", l.Name)

	conn_pipe := make(chan net.Conn)

	go func() {
		for {
			conn, err := l.Socket.Accept()
			switch {
			case err == nil && conn != nil: // connection
				l.Logger.Debugf("[listener:%s] Accepted connection from %s", l.Name, conn.RemoteAddr().String())
				conn_pipe <- conn
			case err != nil && conn != nil: // other error
				l.Logger.Errorf("[listener:%s] Can't accept connectionfrom %s: %v", l.Name, conn.RemoteAddr().String(), err)
			case err != nil && conn == nil: // exiting
				return
			}
		}
	}()

	for {
		select {
		case <-l.ExitChan:
			l.Logger.Debugf("[listener:%s] Received exit signal", l.Name)
			l.Stop()
			return
		case conn := <-conn_pipe:
			go l.handleConnection(conn)
		}
	}
}

func (l *Listener) Stop() {
	l.Logger.Infof("[listener:%s] Stopping...", l.Name)
	l.Logger.Debugf("[listener:%s] Closing LISTEN socket", l.Name)
	l.Socket.Close()
	l.Logger.Debugf("[listener:%s] Socket closed", l.Name)
	l.Logger.Debugf("[listener:%s] Processing remaining metrics", l.Name)
	l.ConnWg.Wait()
	l.Logger.Debugf("[listener:%s] Remaining metrics processed", l.Name)
	l.Logger.Infof("[listener:%s] Stopped", l.Name)
}

func (l *Listener) handleConnection(conn net.Conn) {
	l.ConnWg.Add(1)
	scn := bufio.NewScanner(conn)
	for scn.Scan() {
		line := scn.Text()
		metric, err := NewMetricFromLine(line, l.Config.Codec, l.GraphiteMutator)
		if err == nil {
			if metric.OK {
				err = l.Buffer.Push(&metric)
				if err != nil {
					l.Logger.Errorf("[listener:%s] Can't push metric into Redis buffer: %v", l.Name, err)
				}
			} else {
				l.Logger.Debugf("[listener:%s] Empty line, skipping", l.Name)
			}
		} else {
			l.Logger.Errorf("[listener:%s] %v", l.Name, err)
		}
	}
	conn.Close()
	l.Logger.Debugf("[listener:%s] Closed connection from %s", l.Name, conn.RemoteAddr().String())
	l.ConnWg.Done()
}

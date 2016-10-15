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
	ConnWg    sync.WaitGroup
	DataWg    sync.WaitGroup
	ModuleWg  *sync.WaitGroup
	Transport Transport
	Codec     Codec
	Logger    *Logger
	Stats     *ListenerStats
	ExitFlag  *Flag
}

func NewListener(
	name string,
	c ListenerConfig,
	t Transport,
	moduleWg *sync.WaitGroup,
	logger *Logger,
	exitFlag *Flag,
) (Listener, error) {
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
		ConnWg:    sync.WaitGroup{},
		DataWg:    sync.WaitGroup{},
		ModuleWg:  moduleWg,
		Transport: t,
		Codec:     codec,
		Logger:    logger,
		ExitFlag:  exitFlag,
		Stats:     NewListenerStats(),
	}, nil
}

func (l *Listener) Start() {
	l.ModuleWg.Add(1)
	defer l.ModuleWg.Done()

	l.Logger.Infof("[listener:%s] Starting to accept connections", l.Name)

	connPipe := make(chan *net.Conn, 1000)
	dataPipe := make(chan *bytes.Buffer, 100000)
	exitMux := make(chan struct{}, 1)
	exitDecoders := make(chan struct{})
	exitFinished := make(chan struct{}, 1)
	decoderWg := sync.WaitGroup{}

	// connection acceptor
	go func() {
		for {
			conn, err := l.Socket.Accept()
			if err != nil {
				l.Logger.Errorf("[listener:%s] Can't accept connection: %v", l.Name, err)
				return
			}
			l.ConnWg.Add(1)
			l.Stats.ConnOpen.Increment(1)
			connPipe <- &conn
		}
	}()

	// decoder multiplexer
	go func() {
		for {
			select {
			case <-exitMux:
				l.Logger.Debugf("[listener:%s] Closing LISTEN socket", l.Name)
				l.Socket.Close()
				l.Logger.Infof("[listener:%s] LISTEN socket closed", l.Name)
				l.Logger.Infof("[listener:%s] Waiting for connections to close", l.Name)
				l.ConnWg.Wait()
				l.Logger.Infof("[listener:%s] All connections closed", l.Name)
				time.Sleep(10 * time.Millisecond)
				l.Logger.Debugf("[listener:%s] Waiting for decoders to finish", l.Name)
				l.DataWg.Wait()
				for _n := 0; _n < l.Config.Decoders; _n++ {
					exitDecoders <- struct{}{}
				}
				close(dataPipe)
				decoderWg.Wait()
				l.Logger.Infof("[listener:%s] Decoders finished", l.Name)
				exitFinished <- struct{}{}
				return
			case conn := <-connPipe:
				go l.read(*conn, &dataPipe, time.Now())
			}
		}
	}()

	// decoders
	for _n := 0; _n < l.Config.Decoders; _n++ {
		go func() {
			defer decoderWg.Done()
			decoderWg.Add(1)
			for {
				select {
				case data, ok := <-dataPipe:
					if ok {
						l.decode(data)
					}
				case <-exitDecoders:
					func() {
						for {
							data, ok := <-dataPipe
							if !ok {
								return
							}
							l.decode(data)
						}
					}()
					return
				}
			}
		}()
	}

	// update dataPipe statistic
	go func() {
		for {
			l.Stats.CodecToProcess.Set(int64(len(dataPipe)))
			time.Sleep(1 * time.Second)
		}
	}()

	// shutdown handler
	for {
		if l.ExitFlag.Get() {
			l.Logger.Infof("[listener:%s] Stopping...", l.Name)
			exitMux <- struct{}{}
			<-exitFinished
			l.Logger.Infof("[listener:%s] Stopped", l.Name)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func (l *Listener) LogReport() {
	l.Logger.Infof("[listener:%s] connections: %d/%d/%d/%.3f (open/total/total_failed/rate_per_sec), connection_time: %s/%s (avg/max)",
		l.Name,
		l.Stats.ConnOpen.Get(),
		l.Stats.ConnProcessed.Total(),
		l.Stats.ConnFailed.Total(),
		l.Stats.ConnProcessed.Rate(time.Second),
		l.Stats.ConnTime.Avg(),
		l.Stats.ConnTime.Max(),
	)
	l.Logger.Infof("[listener:%s] decoders: %d/%d/%d (processing/to_process/total_processed), metrics: %d/%.3f (total_decoded/rate_per_sec), decoding_time: %s/%s (avg/max)",
		l.Name,
		l.Stats.CodecProcessing.Get(),
		l.Stats.CodecToProcess.Get(),
		l.Stats.CodecProcessed.Total(),
		l.Stats.CodecDecodedMetrics.Total(),
		l.Stats.CodecDecodedMetrics.Rate(time.Second),
		l.Stats.CodecTime.Avg(),
		l.Stats.CodecTime.Max(),
	)

}

func (l *Listener) read(conn net.Conn, pipe *chan *bytes.Buffer, tStart time.Time) {
	defer l.Stats.ConnProcessed.Increment(1)
	defer l.ConnWg.Done()
	l.Logger.Debugf("[listener:%s] Accepted connection from %s", l.Name, conn.RemoteAddr().String())
	iBuf := bufio.NewReader(conn)
	var oBuf bytes.Buffer
	_, err := io.Copy(&oBuf, iBuf)
	conn.Close()
	dur := time.Since(tStart)
	l.Stats.ConnOpen.Decrement(1)
	if err != nil {
		l.Stats.ConnFailed.Increment(1)
		l.Logger.Errorf("[listener:%s] Error reading connection data from %s: %v", l.Name, conn.RemoteAddr().String(), err)
		return
	}
	l.Logger.Debugf("[listener:%s] Handled connection from %s, %d bytes, took %v", l.Name, conn.RemoteAddr().String(), oBuf.Len(), dur)
	l.Stats.ConnTime.Add(dur)
	l.DataWg.Add(1)
	*pipe <- &oBuf

}

func (l *Listener) decode(data *bytes.Buffer) {
	t0 := time.Now()
	defer l.Stats.CodecProcessed.Increment(1)
	defer l.Stats.CodecProcessing.Decrement(1)
	defer l.DataWg.Done()
	l.Stats.CodecProcessing.Increment(1)
	metrics, errs := l.Codec.Decode(bytes.NewReader(data.Bytes()))
	for metric := range metrics {
		l.Transport.InputChan() <- metric
		l.Stats.CodecDecodedMetrics.Increment(1)
	}
	if len(errs) > 0 {
		l.Logger.Errorf("[listener:%s] Failed to decode %d metrics!", l.Name, len(errs))
		// log the metric raw data?
	}
	l.Stats.CodecTime.Add(time.Since(t0))
}

type ListenerStats struct {
	ConnProcessed       *StatsCounter
	ConnFailed          *StatsCounter
	ConnOpen            *StatsGauge
	ConnTime            *StatsTimer
	CodecProcessed      *StatsCounter
	CodecProcessing     *StatsGauge
	CodecToProcess      *StatsGauge
	CodecDecodedMetrics *StatsCounter
	CodecTime           *StatsTimer
}

func NewListenerStats() *ListenerStats {
	now := time.Now()
	return &ListenerStats{
		ConnProcessed:       NewStatsCounter(now),
		ConnFailed:          NewStatsCounter(now),
		ConnOpen:            NewStatsGauge(),
		ConnTime:            NewStatsTimer(1000),
		CodecProcessed:      NewStatsCounter(now),
		CodecProcessing:     NewStatsGauge(),
		CodecToProcess:      NewStatsGauge(),
		CodecDecodedMetrics: NewStatsCounter(now),
		CodecTime:           NewStatsTimer(1000),
	}
}

func (s *ListenerStats) Reset() {
	s.ConnProcessed.Reset()
	s.ConnFailed.Reset()
	s.CodecProcessed.Reset()
	s.CodecDecodedMetrics.Reset()
}

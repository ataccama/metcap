package metcap

import (
	"regexp"
	"strconv"
	"sync"
	"time"

	"gopkg.in/redis.v4"
)

type RedisTransport struct {
	Redis           *redis.Client
	Size            int
	Wait            int
	Queue           string
	ListenerEnabled bool
	WriterEnabled   bool
	Input           chan *Metric
	Output          chan *Metric
	ExitChan        chan bool
	ExitFlag        *Flag
	Wg              *sync.WaitGroup
	Stats           *RedisTransportStats
	Logger          *Logger
}

// NewRedisTransport
func NewRedisTransport(c *TransportConfig, listenerEnabled bool, writerEnabled bool, exitFlag *Flag, logger *Logger) (*RedisTransport, error) {
	connRe := regexp.MustCompile(`^(?P<network>(tcp|unix)):/{2,3}(?P<addr>[0-9a-zA-Z\._]+:[0-9]+)|(?P<db>1?[0-9])?$`)
	connMatch := connRe.FindStringSubmatch(c.RedisURL)
	connData := map[string]string{}
	for i, n := range connRe.SubexpNames() {
		connData[n] = connMatch[i]
	}

	if c.BufferSize == 0 {
		c.BufferSize = 1000
	}

	if connData["db"] == "" {
		connData["db"] = "0"
	}
	dbNum, err := strconv.Atoi(connData["db"])
	if err != nil {
		return nil, &TransportError{"redis", err}
	}

	if c.RedisQueue == "" {
		c.RedisQueue = "default"
	}

	conn := redis.NewClient(&redis.Options{
		Network:     connData["network"],
		Addr:        connData["addr"],
		DB:          dbNum,
		MaxRetries:  c.RedisRetries,
		PoolSize:    c.RedisConnections,
		PoolTimeout: time.Duration(c.RedisTimeout) * time.Second},
	)

	_, err = conn.Ping().Result()
	if err != nil {
		return nil, &TransportError{"redis", err}
	}

	return &RedisTransport{
		Redis:           conn,
		Size:            c.BufferSize,
		Queue:           "metcap:" + c.RedisQueue,
		Wait:            c.RedisWait,
		ListenerEnabled: listenerEnabled,
		WriterEnabled:   writerEnabled,
		Input:           make(chan *Metric, c.BufferSize),
		Output:          make(chan *Metric, c.BufferSize),
		ExitChan:        make(chan bool, 1),
		ExitFlag:        exitFlag,
		Wg:              &sync.WaitGroup{},
		Stats:           NewRedisTransportStats(),
		Logger:          logger,
	}, nil
}

func (t *RedisTransport) Start() {

	if t.ListenerEnabled {
		go func() {
			t.Wg.Add(1)
			defer t.Wg.Done()
			for {
				select {
				case m := <-t.Input:
					err := t.Redis.RPush(t.Queue, m.Serialize()).Err()
					if err != nil {
						t.Logger.Error("[redis] Failed to push metric: %v - %v", err, err.Error())
						continue
					}
				case <-t.ExitChan:
					for m := range t.Input {
						err := t.Redis.RPush(t.Queue, m.Serialize()).Err()
						if err != nil {
							t.Logger.Error("[redis] Failed to push metric: %v - %v", err, err.Error())
							continue
						}
					}
					return
				}
			}
		}()
	}

	if t.WriterEnabled {
		go func() {
			t.Wg.Add(1)
			defer t.Wg.Done()
			for {
				if t.ExitFlag.Get() {
					t.ExitChan <- true
					return
				}
				m, err := t.Redis.BLPop(time.Duration(t.Wait)*time.Second, t.Queue).Result()
				if err != nil {
					t.Logger.Error("[redis] Failed to get metric: %v - %v", err, err.Error())
				}
				if m != nil {
					metric, err := DeserializeMetric(m[1])
					if err == nil {
						t.Output <- &metric
					} else {
						t.Logger.Error("[redis] failed to DeserializeMetric(): %v - %v", err, err.Error())
					}
				}
			}
		}()
	}

	// ticker
	tick := make(chan struct{}, 1)
	go func() {
		for {
			tick <- struct{}{}
			time.Sleep(100)
		}
	}()

	go func() {
		n := 0
		for {
			select {
			case <-tick:
				if n <= 9 {
					qSize, err := t.Redis.LLen(t.Queue).Result()
					if err == nil {
						t.Stats.QueueSize.Set(qSize)
					}
					n = 0
				}
			case <-t.ExitChan:
				return
			}
		}
	}()
}

func (t *RedisTransport) Stop() {
	t.Wg.Wait()
	t.Redis.Close()
}

func (t *RedisTransport) CloseOutput() {
	return
}

func (t *RedisTransport) CloseInput() {
	return
}

func (t *RedisTransport) InputChan() chan<- *Metric {
	return t.Input
}

func (t *RedisTransport) OutputChan() <-chan *Metric {
	return t.Output
}

func (t *RedisTransport) InputChanLen() int {
	return len(t.Input)
}

func (t *RedisTransport) OutputChanLen() int {
	return len(t.Output)
}

func (t *RedisTransport) LogReport() {

}

type RedisTransportStats struct {
	QueueSize     *StatsGauge
	InputChannel  *StatsGauge
	OutputChannel *StatsGauge
}

func NewRedisTransportStats() *RedisTransportStats {
	return &RedisTransportStats{
		QueueSize:     NewStatsGauge(),
		InputChannel:  NewStatsGauge(),
		OutputChannel: NewStatsGauge(),
	}
}

func (s *RedisTransportStats) Reset() {}

func (s *RedisTransportStats) Report() {}

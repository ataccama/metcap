package metcap

import (
	"gopkg.in/redis.v4"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type RedisTransport struct {
	Redis           *redis.Client
	Size            int
	Wait            int
	Queue           string
	ListenerEnabled bool
	WriterEnabled   bool
	Listener        chan *Metric
	Writer          chan *Metric
	ExitChan        chan bool
	ExitFlag        *Flag
	Wg              *sync.WaitGroup
}

// NewRedisTransport
func NewRedisTransport(c *TransportConfig, listenerEnabled bool, writerEnabled bool, exitFlag *Flag) *RedisTransport {
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
	dbNum, _ := strconv.Atoi(connData["db"])

	if c.RedisQueue == "" {
		c.RedisQueue = "default"
	}

	return &RedisTransport{
		Redis: redis.NewClient(&redis.Options{
			Network:     connData["network"],
			Addr:        connData["addr"],
			DB:          dbNum,
			PoolSize:    c.RedisConnections,
			PoolTimeout: time.Duration(c.RedisTimeout) * time.Second}),
		Size:            c.BufferSize,
		Queue:           "metcap:" + c.RedisQueue,
		Wait:            c.RedisWait,
		ListenerEnabled: listenerEnabled,
		WriterEnabled:   writerEnabled,
		Listener:        make(chan *Metric, c.BufferSize),
		Writer:          make(chan *Metric, c.BufferSize),
		ExitChan:        make(chan bool, 1),
		ExitFlag:        exitFlag,
		Wg:              &sync.WaitGroup{},
	}
}

func (t *RedisTransport) Start() {

	if t.ListenerEnabled {
		go func() {
			t.Wg.Add(1)
			defer t.Wg.Done()
			for {
				select {
				case m := <-t.Listener:
					t.Redis.RPush(t.Queue, m.Serialize()).Err()
				case <-t.ExitChan:
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
				switch {
				case t.ExitFlag.Get():
					t.ExitChan <- true
					return
				default:
					m := t.Redis.BLPop(time.Duration(t.Wait)*time.Second, t.Queue).Val()
					if m != nil {
						metric, err := DeserializeMetric(m[1])
						if err == nil {
							t.Writer <- &metric
						}
					}
				}
			}
		}()
	}
}

func (t *RedisTransport) Stop() {
	t.Wg.Wait()
	t.Redis.Close()
}

func (t *RedisTransport) ListenerChan() chan<- *Metric {
	return t.Listener
}

func (t *RedisTransport) WriterChan() <-chan *Metric {
	return t.Writer
}

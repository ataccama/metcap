package metcap

import (
  "time"
  "gopkg.in/redis.v4"
)

type Buffer struct {
  Redis     *redis.Client
  Queue     string
  Wait      int
  ExitChan  chan bool
  Logger    *Logger
}

// initialize buffer
func NewBuffer(c *BufferConfig, logger *Logger) *Buffer {
  logger.Infof("Initializing Redis buffer [tcp://%s/%d]", c.Address, c.DB)
  return &Buffer{
    Redis: redis.NewClient(&redis.Options{
      Network: c.Socket,
      Addr: c.Address,
      DB: c.DB,
      PoolSize: c.Connections,
      PoolTimeout: time.Duration(c.Timeout) * time.Second}),
    Queue: "mc:" + c.Queue,
    Wait: c.Wait,
    ExitChan: make(chan bool)}
}

// send metric to buffer
func (b *Buffer) Push(m *Metric) error {
  return b.Redis.RPush(b.Queue, m.Serialize()).Err()
}

// retrieve metric from buffer
func (b *Buffer) Pop() (Metric, error) {
  m, err := b.Redis.BLPop(time.Duration(b.Wait) * time.Second, b.Queue).Result()
  if err != nil {
    return Metric{}, err
  }
  metric, err := DeserializeMetric(m[1])
  return metric, err
}

func (b *Buffer) Close() {
  b.Redis.Close()
}

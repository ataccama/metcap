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
}

func NewBuffer(c *BufferConfig) *Buffer {
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

func (b *Buffer) Push(m *Metric) error {
  return b.Redis.RPush(b.Queue, m.JSON()).Err()
}

func (b *Buffer) Pop() (Metric, error) {
  m, err := b.Redis.BLPop(time.Duration(b.Wait) * time.Second, b.Queue).Result()
  if err != nil {
    return Metric{}, err
  }
  metric, err := NewMetricFromJSON([]byte(m[1]))
  return metric, err
}

func (b *Buffer) Close() {
  b.Redis.Close()
}

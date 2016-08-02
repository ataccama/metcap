package metcap

import (
  "time"

  "gopkg.in/redis.v4"
)

type Buffer struct {
  Redis     *redis.Client
  Queue     string
  ExitChan  chan bool
}

func NewBuffer(c *RedisConfig) *Buffer {
  return &Buffer{
    Redis: redis.NewClient(&redis.Options{
      Network: c.Socket,
      Addr: c.Address,
      DB: c.DB,
      PoolSize: c.Connections,
      PoolTimeout: time.Duration(c.Timeout) * time.Second}),
    ExitChan: make(chan bool),
    Queue: "mc:" + c.Queue}
}

func (b *Buffer) Push(m *Metric) error {
  return b.Redis.RPush(b.Queue, m.Bufferize).Err()
}

func (b *Buffer) Pop() (Metric, error) {
  m := b.Redis.BLPop(0, b.Queue)
  if m.Err() != nil {
    return Metric{}, m.Err()
  }
  return Unbufferize(m.String())
}

func (b *Buffer) Close() {
  b.Redis.Close()
}

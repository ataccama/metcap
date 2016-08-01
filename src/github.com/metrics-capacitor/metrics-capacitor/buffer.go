package metcap

import (
  "time"

  "github.com/garyburd/redigo/redis"
)

type Buffer struct {
  Redis     *redis.Pool
  ExitChan  chan bool
}

func NewBuffer(c *Config) *Buffer {
  return &Buffer{
    Redis: &redis.Pool{
      MaxIdle: c.Redis.MaxIdle,
      MaxActive: c.Redis.MaxActive,
      IdleTimeout: time.Duration(c.Redis.Timeout) * time.Second,
      Dial: func () (redis.Conn, error) {
        r, err := redis.Dial("tcp", c.Redis.Url)
        if err != nil {
          return nil, err
        }
        return r, err
      },
      TestOnBorrow: func (r redis.Conn, t time.Time) (error) {
        _, err := r.Do("PING")
        return err
      },
    },
    ExitChan: make(chan bool)}
}

func (b *Buffer) Push(m *Metrics) {
  return
}

func (b *Buffer) Pop(count int) Metrics {
  return nil
}

func (b *Buffer) Close() {
  return
}

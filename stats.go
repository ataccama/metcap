package metcap

import (
	"sync"
	"time"
)

// ----- Stats Gauge ------
type StatsGauge struct {
	*sync.Mutex
	val int64
}

func NewStatsGauge() *StatsGauge {
	return &StatsGauge{&sync.Mutex{}, 0}
}

func (g *StatsGauge) Set(n int64) {
	g.Lock()
	defer g.Unlock()
	g.val = n
}

func (g *StatsGauge) Get() int64 {
	g.Lock()
	defer g.Unlock()
	return g.val
}

func (g *StatsGauge) Increment(n int) {
	g.Lock()
	defer g.Unlock()
	g.val = g.val + int64(n)
}

func (g *StatsGauge) Decrement(n int) {
	g.Lock()
	defer g.Unlock()
	g.val = g.val - int64(n)
}

// ----- Stats Counter ------
type StatsCounter struct {
	*sync.Mutex
	val   uint64
	count uint64
	since time.Time
}

func NewStatsCounter(t time.Time) *StatsCounter {
	return &StatsCounter{&sync.Mutex{}, 0, 0, t}
}

func (c *StatsCounter) Increment(n int) {
	c.Lock()
	defer c.Unlock()
	c.val = c.val + uint64(n)
	c.count = c.count + 1
}

func (c *StatsCounter) Reset() {
	c.Lock()
	defer c.Unlock()
	c.val, c.count = 0, 0
	c.since = time.Now()
}

func (c *StatsCounter) Total() uint64 {
	c.Lock()
	defer c.Unlock()
	return c.val
}

func (c *StatsCounter) Count() uint64 {
	c.Lock()
	defer c.Unlock()
	return c.count
}

func (c *StatsCounter) Avg() float64 {
	c.Lock()
	defer c.Unlock()
	return float64(c.val) / float64(c.count)
}

func (c *StatsCounter) Rate(div ...time.Duration) float64 {
	c.Lock()
	total := float64(c.val)
	dur := time.Since(c.since)
	c.Unlock()
	switch div[0] {
	case time.Second:
		return total / float64(dur.Seconds())
	case time.Minute:
		return total / float64(dur.Minutes())
	case time.Hour:
		return total / float64(dur.Hours())
	default:
		return total / float64(dur.Seconds())
	}
}

func (c *StatsCounter) Since() time.Time {
	c.Lock()
	defer c.Unlock()
	return c.since
}

// ----- Stats Timer ------
type StatsTimer struct {
	*sync.Mutex
	vals []time.Duration
}

func NewStatsTimer(size int) *StatsTimer {
	return &StatsTimer{&sync.Mutex{}, make([]time.Duration, size)}
}

func (s *StatsTimer) Add(t time.Duration) {
	s.Lock()
	defer s.Unlock()
	_, s.vals = s.vals[0], append(s.vals[1:], t)
}

func (s *StatsTimer) Avg() time.Duration {
	s.Lock()
	defer s.Unlock()
	var total uint64
	var count uint64
	for _, t := range s.vals {
		if t > 0 {
			total = total + uint64(t.Nanoseconds())
			count = count + 1
		}
	}
	if count == 0 {
		return time.Duration(0)
	}
	return time.Duration(int64(total / count))
}

func (s *StatsTimer) Max() time.Duration {
	s.Lock()
	defer s.Unlock()
	var max time.Duration
	for _, t := range s.vals {
		if max < t {
			max = t
		}
	}
	return max
}

// ------ Stats Manual Timer (with reset)
type StatsManualTimer struct {
	*sync.Mutex
}

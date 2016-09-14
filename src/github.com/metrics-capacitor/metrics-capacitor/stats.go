package metcap

import (
	"sync"
	"time"
)

type Stats struct {
	Transport *TransportStats
	Listeners map[string]*ListenerStats
	Writer    *WriterStats
	Engine    *EngineStats
}

func (s *Stats) JSON() string {
	return ""
}

func (s *Stats) Reset() {
	s.Engine.Reset()
	s.Transport.Reset()
	for _, l := range s.Listeners {
		l.Reset()
	}
	s.Writer.Reset()
}

type TransportStats struct {
	QueueSize     *StatsGauge
	InputChannel  *StatsGauge
	OutputChannel *StatsGauge
}

func NewTransportStats() *TransportStats {
	return &TransportStats{
		QueueSize:     NewStatsGauge(),
		InputChannel:  NewStatsGauge(),
		OutputChannel: NewStatsGauge(),
	}
}

func (s *TransportStats) Reset() {}

type ListenerStats struct {
	Conns           *StatsCounter
	ConnFail        *StatsCounter
	ConnOpen        *StatsGauge
	ConnTime        *StatsTimer
	CodecProcessed  *StatsCounter
	CodecProcessing *StatsGauge
	CodecTime       *StatsTimer
}

func NewListenerStats() *ListenerStats {
	now := time.Now()
	return &ListenerStats{
		Conns:           NewStatsCounter(now),
		ConnFail:        NewStatsCounter(now),
		ConnOpen:        NewStatsGauge(),
		ConnTime:        NewStatsTimer(100000),
		CodecProcessed:  NewStatsCounter(now),
		CodecProcessing: NewStatsGauge(),
		CodecTime:       NewStatsTimer(100000),
	}
}

func (s *ListenerStats) Reset() {
	s.Conns.Reset()
	s.ConnFail.Reset()
	s.CodecProcessed.Reset()
}

type WriterStats struct {
	Committed *StatsCounter
	Failed    *StatsCounter
}

func NewWriterStats() *WriterStats {
	now := time.Now()
	return &WriterStats{
		Committed: NewStatsCounter(now),
		Failed:    NewStatsCounter(now),
	}
}

func (s *WriterStats) Reset() {
	s.Committed.Reset()
	s.Failed.Reset()
}

type EngineStats struct {
}

func NewEngineStats() *EngineStats {
	return &EngineStats{}
}

func (s *EngineStats) Reset() {

}

// ----- Stats Gauge ------
type StatsGauge struct {
	*sync.Mutex
	val int64
}

func NewStatsGauge() *StatsGauge {
	return &StatsGauge{&sync.Mutex{}, 0}
}

func (g *StatsGauge) Set() {
	g.Lock()
	defer g.Unlock()
}

func (g *StatsGauge) Get() int64 {
	g.Lock()
	defer g.Unlock()
	return g.val
}

func (g *StatsGauge) Increment(n int64) {
	g.Lock()
	defer g.Unlock()
	g.val = g.val + n
}

func (g *StatsGauge) Decrement(n int64) {
	g.Lock()
	defer g.Unlock()
	g.val = g.val - n
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

func (c *StatsCounter) Increment(n uint64) {
	c.Lock()
	defer c.Unlock()
	c.val = c.val + n
	c.count = c.count + 1
}

func (c *StatsCounter) Reset() {
	c.Lock()
	defer c.Unlock()
	c.val, c.count = 0, 0
	c.since = time.Now()
}

func (c *StatsCounter) Count() uint64 {
	c.Lock()
	defer c.Unlock()
	return c.val
}

func (c *StatsCounter) Avg() float64 {
	c.Lock()
	defer c.Unlock()
	return float64(c.val) / float64(c.count)
}

func (c *StatsCounter) Rate(div ...time.Duration) float64 {
	c.Lock()
	defer c.Unlock()
	if len(div) == 0 {
		div[0] = time.Second
	}
	return float64(c.val) / float64((time.Since(c.since) * div[0]))
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

func (s *StatsTimer) Avg(div ...time.Duration) time.Duration {
	s.Lock()
	defer s.Unlock()
	var total uint64
	var count uint64
	var avg time.Duration
	for _, t := range s.vals {
		if t > 0 {
			total = total + uint64(t.Nanoseconds())
			count = count + 1
		}
	}
	if len(div) == 0 {
		div[0] = time.Second
	}
	avg = time.Duration(int64(total/count)) * time.Nanosecond
	return time.Duration(avg) * div[0]
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

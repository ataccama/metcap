package metcap

type ChannelTransport struct {
	Size int
	Chan chan *Metric
}

func NewChannelTransport(c *TransportConfig) *ChannelTransport {
	return &ChannelTransport{
		Size: c.BufferSize,
		Chan: make(chan *Metric, c.BufferSize),
	}
}

func (t *ChannelTransport) Start() { return }

func (t *ChannelTransport) Stop() { return }

func (t *ChannelTransport) ListenerChan() chan<- *Metric {
	return t.Chan
}

func (t *ChannelTransport) WriterChan() <-chan *Metric {
	return t.Chan
}

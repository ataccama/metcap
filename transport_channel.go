package metcap

type ChannelTransport struct {
	Size   int
	Chan   chan *Metric
	Logger *Logger
}

func NewChannelTransport(c *TransportConfig, logger *Logger) *ChannelTransport {
	return &ChannelTransport{
		Size:   c.BufferSize,
		Chan:   make(chan *Metric, c.BufferSize),
		Logger: logger,
	}
}

func (t *ChannelTransport) Start() { return }

func (t *ChannelTransport) Stop() { return }

func (t *ChannelTransport) CloseOutput() {
	return
}

func (t *ChannelTransport) CloseInput() {
	return
}

func (t *ChannelTransport) InputChan() chan<- *Metric {
	return t.Chan
}

func (t *ChannelTransport) OutputChan() <-chan *Metric {
	return t.Chan
}

func (t *ChannelTransport) InputChanLen() int {
	return len(t.Chan)
}

func (t *ChannelTransport) OutputChanLen() int {
	return len(t.Chan)
}

func (t *ChannelTransport) LogReport() {
	t.Logger.Infof("[transport] channel: %d/%d (length/capacity)", len(t.Chan), t.Size)
}

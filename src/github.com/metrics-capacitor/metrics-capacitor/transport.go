package metcap

type Transport interface {
	Start()
	Stop()
	// Status() TransportStatus
	ListenerChan() chan<- *Metric
	WriterChan() <-chan *Metric
}

type TransportStatus struct {
	InputCount     int64
	OutputCount    int64
	InputErrCount  int64
	OutputErrCount int64
	InputRate      float64
	OutputRate     float64
	InputErrRate   float64
	OutputErrRate  float64
}

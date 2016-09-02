package metcap

type Transport interface {
	Start()
	Stop()
	ListenerChan() chan<- *Metric
	WriterChan() <-chan *Metric
}

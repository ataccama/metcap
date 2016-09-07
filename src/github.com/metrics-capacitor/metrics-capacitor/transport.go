package metcap

import "fmt"

type Transport interface {
	Start()
	Stop()
	ListenerChan() chan<- *Metric
	WriterChan() <-chan *Metric
}

type TransportError struct {
	provider string
	err      error
}

func (e *TransportError) Error() string {
	return fmt.Sprintf("[%s] Error: %s", e.provider, e.err.Error())
}

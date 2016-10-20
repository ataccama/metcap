package metcap

import "fmt"

type Transport interface {
	Start()
	Stop()
	CloseInput()
	CloseOutput()
	LogReport()
	InputChan() chan<- *Metric
	InputChanLen() int
	OutputChan() <-chan *Metric
	OutputChanLen() int
}

type TransportError struct {
	provider string
	err      error
}

func (e *TransportError) Error() string {
	return fmt.Sprintf("[%s] Error: %v", e.provider, e.err)
}

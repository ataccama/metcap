package metcap

type Transport interface {
	Start()
	Stop()
	StopInput()
	StopOutput()
	LogReport()
	InputChan() chan<- *Metric
	InputChanLen() int
	OutputChan() <-chan *Metric
	OutputChanLen() int
}

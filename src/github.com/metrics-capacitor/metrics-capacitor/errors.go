package metcap

import "fmt"

type MetricFromLineError struct {
	msg  string
	line string
}

func (e *MetricFromLineError) Error() string {
	return fmt.Sprintf("%s (LINE: %s)", e.msg, e.line)
}

type TransportError struct {
	provider string
	err      error
}

func (e *TransportError) Error() string {
	return fmt.Sprintf("[%s] Error: %s", e.provider, e.err.Error())
}

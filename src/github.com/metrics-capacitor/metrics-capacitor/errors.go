package metcap

import "fmt"

type TransportError struct {
	provider string
	err      error
}

func (e *TransportError) Error() string {
	return fmt.Sprintf("[%s] Error: %s", e.provider, e.err.Error())
}

type CodecError struct {
	msg string
	err error
	src interface{}
}

func (e *CodecError) Error() string {
	return fmt.Sprintf("%s - &v [%v]", e.msg, e.err, e.src)
}

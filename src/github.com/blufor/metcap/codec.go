package metcap

import (
	"fmt"
	"io"
)

type Codec interface {
	Decode(io.Reader) (<-chan *Metric, <-chan error)
}

type CodecError struct {
	msg string
	err error
	src interface{}
}

func (e *CodecError) Error() string {
	return fmt.Sprintf("%s - %v [%v]", e.msg, e.err, e.src)
}

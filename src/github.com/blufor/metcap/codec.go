package metcap

import "io"

type Codec interface {
	Decode(io.Reader) (<-chan *Metric, <-chan error)
}

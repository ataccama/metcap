package metcap

import (
	"io"
	"time"
)

type Codec interface {
	Decode(io.Reader) ([]Metric, time.Duration, []error)
}

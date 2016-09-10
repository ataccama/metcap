package metcap

import (
	"io"
	"time"
)

type Codec interface {
	Decode(io.ReadWriter) ([]Metric, time.Duration, []error)
}

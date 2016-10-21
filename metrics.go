package metcap

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
)

// Metric struct
//
type Metric struct {
	Name      string            `json:"name"`
	Timestamp time.Time         `json:"@timestamp"`
	Value     float64           `json:"value"`
	Fields    map[string]string `json:"fields"`
	OK        bool              `json:"ok"`
}

type Metrics []Metric

func (m *Metric) JSON() []byte {
	out, err := json.Marshal(m)
	if err != nil {
		panic(err) // REFACTOR: throw error and do checking
	}
	return out
}

func (m *Metric) Serialize() []byte {
	out, err := msgpack.Marshal(m)
	if err != nil {
		panic(err) // REFACTOR: throw error and do checking
	}
	return out
}

func (m *Metric) Index(name string) string {
	t := m.Timestamp.UTC()
	return fmt.Sprintf("%s-%d.%02d.%02d", name, t.Year(), int(t.Month()), t.Day())
}

func DeserializeMetric(data string) (Metric, error) {
	var m Metric
	err := msgpack.Unmarshal([]byte(data), &m)
	if err != nil {
		return Metric{}, err
	}
	return m, nil
}

/// generate Metric from JSON
/// TODO: will be implemented within JSON codec
// func NewMetricFromJSON(j []byte) (Metric, error) {
// 	var m Metric
// 	err := json.Unmarshal(j, &m)
// 	if err != nil {
// 		return Metric{}, err
// 	}
// 	return m, nil
// }

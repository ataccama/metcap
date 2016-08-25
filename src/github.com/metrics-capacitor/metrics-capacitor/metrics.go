package metcap

import (
	"encoding/json"
	"fmt"
	"regexp"
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
	OK				bool							`json:"ok"`
}

type Metrics []Metric

func (m *Metric) JSON() []byte {
	out, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return out
}

func (m *Metric) Serialize() []byte {
	out, err := msgpack.Marshal(m)
	if err != nil {
		panic(err)
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

// generate Metric from JSON
func NewMetricFromJSON(j []byte) (Metric, error) {
	var m Metric
	err := json.Unmarshal(j, &m)
	if err != nil {
		return Metric{}, err
	}
	return m, nil
}

// generate Metric from Graphite data listener
func NewMetricFromLine(line string, codec string, mut *[]string) (Metric, error) {
	var pat string

	switch codec {
	case "graphite":
		pat = `^(?P<path>[a-zA-Z0-9_\-\.]+) (?P<value>[0-9\.]+)(\ (?P<timestamp>[0-9]{10,13}))?$`
	case "influx":
		pat = `^(?P<name>[a-zA-Z0-9_\-\.]+) (?P<fields>[a-zA-Z0-9,_\-\.\=]+) (?P<value>[0-9\.]+)( (?P<timestamp>\d{10,13}))?\s*$`
	}

	re, err := regexp.Compile(pat)

	if err != nil {
		return Metric{OK: false}, err
	}

	if re_empty_line := regexp.MustCompile(`^$`); re_empty_line.Match([]byte(line)) {
		return Metric{OK: false}, nil
	}

	if re.Match([]byte(line)) {
		match := re.FindStringSubmatch(line)
		dissected := map[string]string{}

		for i, n := range re.SubexpNames() {
			dissected[n] = match[i]
		}

		timestamp := parseTimestamp(dissected)

		name, fields, err := parseFields(dissected, mut)
		if err != nil {
			return Metric{OK: false}, err
		}

		value, err := parseValue(dissected)
		if err != nil {
			return Metric{OK: false}, err
		}

		return Metric{
			OK:		 		 true,
			Name:      name,
			Timestamp: timestamp,
			Value:     value,
			Fields:    fields}, nil
	} else {
		return Metric{OK: false}, &NewMetricFromLineError{"Failed to ingest metric", line}
	}
}

// Errors

type NewMetricFromLineError struct {
	msg  string
	line string
}

func (e *NewMetricFromLineError) Error() string {
	return fmt.Sprintf("%s (LINE: %s)", e.msg, e.line)
}

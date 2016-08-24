package metcap

import (
	"encoding/json"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	date := []string{strconv.Itoa(t.Year()), strconv.Itoa(int(t.Month())), strconv.Itoa(t.Day())}
	return name + "_" + strings.Join(date, "-")
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

// helper function to parse timestamp into time.Time
func parseTimestamp(d map[string]string) time.Time {
	var (
		t_now       time.Time
		t_str       string
		t_byte      []byte
		t_len       int
		t_unix_sec  int64
		t_unix_nsec int64
		err         error
	)

	t_now = time.Now()
	t_str = d["timestamp"]
	t_byte = []byte(t_str)
	t_len = len(t_byte)

	switch {
	// time not specified
	case t_len == 0:
		return t_now
	// time is in Unix timestamp
	case t_len <= 11:
		t_int, err := strconv.ParseInt(t_str, 10, 64)
		if err != nil {
			return t_now
		}
		return time.Unix(t_int, 0)
	// time is in Unix timestamp with second fractions
	case t_len > 11:
		t_unix_sec, err = strconv.ParseInt(string(t_byte[:11]), 10, 64)
		if err != nil {
			return t_now
		} else {
			t_unix_nsec, err = strconv.ParseInt(string(t_byte[11:t_len])+strings.Repeat("0", len(t_byte[11:t_len])), 10, 64)
			if err != nil {
				return t_now
			}
		}
		return time.Unix(t_unix_sec, t_unix_nsec*int64(time.Millisecond))
	default:
		return t_now
	}
}

// helper function to parse value as float64
func parseValue(d map[string]string) (float64, error) {
	var (
		value float64
		err   error
	)
	if value, err = strconv.ParseFloat(d["value"], 64); err != nil {
		return float64(0), &ParserError{"Failed to parse value", d["value"]}
	}
	return value, nil
}

// helper function to parse metric name and fields
func parseFields(d map[string]string, mut *[]string) (string, map[string]string, error) {
	name := []string{}
	fields := make(map[string]string)

	// check if we have graphite path
	if _, ok := d["path"]; ok {
		// iterate through mutator rules
		for _, lineRule := range *mut {
			mut_rule := strings.Split(lineRule, "|||")
			mut_re, err := regexp.Compile(mut_rule[0])
			if err != nil {
				continue
			}
			// try to match metric path with a mutator rule
			if mut_re.Match([]byte(d["path"])) {
				field_values := strings.Split(d["path"], ".")
				field_names := strings.Split(mut_rule[1], ".")
				if len(field_values) != len(field_names) {
					continue
				}
				// iterate thru fields
				for i, field := range field_values {
					switch {
					case regexp.MustCompile(`^[0-9]+$`).Match([]byte(field_names[i])):
						name = append(name, field)
					case regexp.MustCompile(`^[a-zA-Z0-9_]+$`).Match([]byte(field_names[i])):
						fields[field_names[i]] = field
					case regexp.MustCompile(`^-$`).Match([]byte(field_names[i])):
						continue
					default:
						continue
					}
				}
			}
		}
		// not Graphite? then it must be only Influx (for now :))
	} else {
		name = append(name, d["name"])
		// iterate thru fields
		for _, field := range strings.Split(d["fields"], ",") {
			kv := strings.Split(field, "=")
			if kv[0] != "" {
				fields[kv[0]] = kv[1]
			}
		}
	}
	if len(name) == 0 {
		return "", make(map[string]string), &ParserError{"Failed to parse metric name", name}
	}
	if len(fields) == 0 {
		return "", make(map[string]string), &ParserError{"Failed to parse metric fields", fields}
	}
	return strings.Join(name, ":"), fields, nil
}

// Errors

type NewMetricFromLineError struct {
	msg  string
	line string
}

func (e *NewMetricFromLineError) Error() string {
	return fmt.Sprintf("%s (LINE: %s)", e.msg, e.line)
}

type ParserError struct {
	msg string
	src interface{}
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("%s - %v", e.msg, e.src)
}

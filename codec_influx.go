package metcap

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InfluxCodec struct {
	lineRegex *regexp.Regexp
	fields    [][2]string
}

func NewInfluxCodec() (InfluxCodec, error) {
	re := regexp.MustCompile(`^(?P<name>[a-zA-Z0-9_\-\.]+) ((?P<fields>[a-zA-Z0-9,_\-\.\=]+)\ )?value=(?P<value>-?[0-9\.]+)(\ (?P<timestamp>\d{10,13}))?$`)

	return InfluxCodec{
		lineRegex: re,
	}, nil
}

func (c InfluxCodec) Decode(input io.Reader) (<-chan *Metric, <-chan error) {
	scn := bufio.NewScanner(input)
	wg := &sync.WaitGroup{}
	metrics := make(chan *Metric)
	errs := make(chan error)

	for scn.Scan() {
		go func(line string) {
			defer wg.Done()
			wg.Add(1)
			if regexp.MustCompile(`^$`).Match([]byte(line)) {
				return
			}
			if !c.lineRegex.Match([]byte(line)) {
				return
			}
			// read name, fields, value and optional timestamp into hash map `dissected`
			match := c.lineRegex.FindStringSubmatch(line)
			dissected := map[string]string{}
			for i, n := range c.lineRegex.SubexpNames() {
				dissected[n] = match[i]
			}
			mTimestamp := c.readTimestamp(dissected)
			mValue, err := c.readValue(dissected)
			if err != nil {
				errs <- &CodecError{"Failed to read value", err, dissected}
				return
			}
			mName, err := c.readName(dissected)
			if err != nil {
				errs <- &CodecError{"Failed to read name", err, dissected}
				return
			}
			mFields, err := c.readFields(dissected)
			if err != nil {
				errs <- &CodecError{"Failed to read fields", err, dissected}
				return
			}
			metrics <- &Metric{Name: mName, Timestamp: mTimestamp, Value: mValue, Fields: mFields}
		}(scn.Text())
	}

	go func() {
		wg.Wait()
		close(metrics)
		close(errs)
	}()

	return metrics, errs
}

func (c InfluxCodec) readTimestamp(d map[string]string) time.Time {
	var (
		tNow      time.Time
		tByte     []byte
		tLen      int
		tUnixSec  int64
		tUnixNsec int64
		err       error
	)

	tNow = time.Now()
	tByte = []byte(d["timestamp"])
	tLen = len(tByte)

	switch {
	// time not specified
	case tLen == 0:
		return tNow
	// time is in Unix timestamp
	case tLen <= 10:
		tInt, err := strconv.ParseInt(string(tByte), 10, 64)
		if err != nil {
			return tNow
		}
		return time.Unix(tInt, 0)
	// time is in Unix timestamp with second fractions
	case tLen > 10:
		tUnixSec, err = strconv.ParseInt(string(tByte[:10]), 10, 64)
		if err != nil {
			return tNow
		}
		tUnixNsec, err = strconv.ParseInt(string(tByte[10:tLen])+strings.Repeat("0", len(tByte[10:tLen])), 10, 64)
		if err != nil {
			return tNow
		}
		return time.Unix(tUnixSec, tUnixNsec*int64(time.Millisecond))
	default:
		return tNow
	}
}

// helper function to parse value as float64
func (c InfluxCodec) readValue(d map[string]string) (float64, error) {
	var (
		value float64
		err   error
	)
	if value, err = strconv.ParseFloat(d["value"], 64); err != nil {
		return float64(0), &CodecError{"Failed to parse value", err, d}
	}
	return value, nil
}

// helper function to parse metric name
func (c InfluxCodec) readName(d map[string]string) (string, error) {
	if name, ok := d["name"]; ok {
		return name, nil
	} else {
		return "", &CodecError{"Failed to parse name", nil, d}
	}
}

// helper function to parse metric fields
func (c InfluxCodec) readFields(d map[string]string) (map[string]string, error) {
	fields := make(map[string]string)
	if _, ok := d["fields"]; ok {
		for _, field := range strings.Split(d["fields"], ",") {
			kv := strings.Split(field, "=")
			if kv[0] != "" {
				fields[kv[0]] = kv[1]
			}
		}
	}
	if len(fields) == 0 {
		return make(map[string]string), &CodecError{"Failed to parse fields", nil, d}
	}
	return fields, nil
}

package metcap

import (
  "encoding/json"
)

// Metric struct
//
type Metric struct {
  Name      string
  Timestamp float64
  Value     float64
  Fields    []MetricField
}

// Metric Document
type MetricDocument struct {
  Name      string
  Timestamp float64
  Value     float64
  Fields    []MetricField
}

// Metric fields struct
//
type MetricField struct {
  Name    string
  Value   string
}

// Multiple metrics is just slice of Metric objects
//
type Metrics []Metric

// Create ElasticSearch Document Bulk
//
func (m *Metric) ToElastic() []byte {
  return []byte{}
}

// Encode for Redis buffer
//
func (m *Metric) Bufferize() []byte {
  return []byte{}
}


// Decode from Redis buffer
//
func Unbufferize(j string) (Metric, error) {
  var m Metric
  err := json.Unmarshal([]byte(j), &m)
  if err != nil {
    return Metric{}, err
  }
  return m, err
}

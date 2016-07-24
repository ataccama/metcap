package main

// Metric struct
//
type Metric struct {
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

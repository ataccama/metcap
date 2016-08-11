package metcap

import (
  "fmt"
  "sync"
  "time"

  "gopkg.in/olivere/elastic.v3"
)

type Writer struct {
  Config    *WriterConfig
  Wg        *sync.WaitGroup
  Buffer    *Buffer
  Elastic   *elastic.Client
  Processor *elastic.BulkProcessor
}

func NewWriter(c *WriterConfig, b *Buffer, wg *sync.WaitGroup) *Writer {
  wg.Add(1)

  es, err := elastic.NewClient(elastic.SetURL(c.Urls...))
  if err != nil {
    fmt.Println("ERROR: Can't connect to Elasticsearch:", err)
  }

  processor, err := elastic.NewBulkProcessorService(es).
    BulkActions(c.BulkMax).
    BulkSize(-1).
    FlushInterval(time.Duration(c.BulkWait) * time.Second).
    Name("metrics-capacitor").
    Stats(true).
    Workers(c.Concurrency).Do()

  if err != nil {
    fmt.Println("ERROR: Can't connect to Elasticsearch:", err)
  }

  return &Writer{
    Config: c,
    Wg: wg,
    Buffer: b,
    Elastic: es,
    Processor: processor}

}

func (w *Writer) Run() {
  pipe_limit := w.Config.BulkMax * w.Config.Concurrency * 1000
  pipe := make(chan Metric, pipe_limit)

  for r := 0; r < w.Config.Concurrency; r++ {
    go w.readFromBuffer(pipe)
  }

  for {
    metric := <-pipe
    req := elastic.NewBulkIndexRequest().Index(metric.Index(w.Config.Index)).Type(w.Config.DocType).Doc(metric.JSON())
    w.Processor.Add(req)
  }

}

func (w *Writer) Stop() {
  w.Processor.Flush()
  w.Processor.Close()
  w.Wg.Done()
}

func (w *Writer) readFromBuffer(p chan Metric)  {
  for {
    metric, err := w.Buffer.Pop()
    if err != nil {
      fmt.Println("ERROR: ", err)
    }
    p <- metric
  }

}

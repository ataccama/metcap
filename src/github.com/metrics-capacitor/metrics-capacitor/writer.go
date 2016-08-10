package metcap

import (
  "fmt"
  "sync"
  // "os"

  "gopkg.in/olivere/elastic.v3"
)

type Writer struct {
  Config  *WriterConfig
  Wg      *sync.WaitGroup
  Buffer  *Buffer
  Elastic *elastic.Client
}

func NewWriter(c *WriterConfig, b *Buffer, wg *sync.WaitGroup) *Writer {
  wg.Add(1)

  es, err := elastic.NewClient(elastic.SetURL(c.Urls...))
  if err != nil {
    fmt.Println("ERROR: Can't connect to Elasticsearch:", err)
    // os.Exit(1)
  }

  return &Writer{
    Elastic: es,
    Config: c,
    Wg: wg,
    Buffer: b}
}

func (w *Writer) Run() {
  //
  // TODO
  //
}

func (w *Writer) Stop() {
  w.Wg.Done()
}

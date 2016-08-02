package metcap

import (
  // "fmt"
  "sync"

  // "gopkg.in/olivere/elastic.v3"
)

type Writer struct {
  Config  *WriterConfig
  Wg      *sync.WaitGroup
  Buffer  *Buffer
}

func NewWriter(c *WriterConfig, b *Buffer, wg *sync.WaitGroup) *Writer {
  wg.Add(1)

  return &Writer{
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
  //
  // TODO
  //
  w.Wg.Done()
}

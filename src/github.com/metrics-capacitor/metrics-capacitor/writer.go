package metcap

import (
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
  Logger    *Logger
}

func NewWriter(c *WriterConfig, b *Buffer, wg *sync.WaitGroup, logger *Logger) *Writer {
  logger.Info("Initializing writer module")
  wg.Add(1)

  logger.Debugf("Connecting to ElasticSearch [%v]", c.Urls)
  es, err := elastic.NewClient(elastic.SetURL(c.Urls...))
  if err != nil {
    logger.Alertf("Can't connect to ElasticSearch: %v", err)
  }
  logger.Debug("Successfully connected to ElasticSearch")

  logger.Debug("Setting up buffer-readers")
  processor, err := elastic.NewBulkProcessorService(es).
    BulkActions(c.BulkMax).
    BulkSize(-1).
    FlushInterval(time.Duration(c.BulkWait) * time.Second).
    Name("metrics-capacitor").
    Stats(true).
    Workers(c.Concurrency).Do()

  if err != nil {
    logger.Alertf("Failed to setup bulk-processor: %v", err)
  }

  return &Writer{
    Config: c,
    Wg: wg,
    Buffer: b,
    Elastic: es,
    Processor: processor,
    Logger: logger}
}

func (w *Writer) Run() {
  w.Logger.Info("Starting writer module")

  pipe_limit := w.Config.BulkMax * w.Config.Concurrency * 100
  pipe := make(chan Metric, pipe_limit)

  for r := 0; r < w.Config.Concurrency; r++ {
    w.Logger.Debugf("Starting writer buffer-reader %2d", r+1)
    go w.readFromBuffer(pipe)
  }
  w.Logger.Info("Writer module started")

  for {
    metric := <-pipe
    w.Logger.Debug("Adding metric to bulk")
    req := elastic.NewBulkIndexRequest().
      Index(metric.Index(w.Config.Index)).
      Type(w.Config.DocType).
      Doc(string(metric.JSON()))
    w.Processor.Add(req)
  }
}

func (w *Writer) Stop() {
  w.Logger.Info("Stopping writer module")
  w.Processor.Flush()
  w.Processor.Close()
  w.Logger.Info("Writer module stopped")
  w.Wg.Done()
}

func (w *Writer) readFromBuffer(p chan Metric)  {
  for {
    metric, err := w.Buffer.Pop()
    if err != nil {
      w.Logger.Error("Failed to BLPOP metric from buffer: " + err.Error())
    } else {
      p <- metric
      w.Logger.Debug("Popped metric from buffer")
    }
  }

}

package metcap

import (
	"sync"
	"time"

	"gopkg.in/olivere/elastic.v3"
)

type Writer struct {
	Config    *WriterConfig
	ModuleWg  *sync.WaitGroup
	Transport Transport
	Elastic   *elastic.Client
	Processor *elastic.BulkProcessor
	Logger    *Logger
	ExitFlag  *Flag
	Stats     *WriterStats
}

func NewWriter(c *WriterConfig, t Transport, module_wg *sync.WaitGroup, logger *Logger, exitFlag *Flag) (Writer, error) {
	logger.Info("[writer] Initializing module")

	logger.Debugf("[writer] Connecting to ElasticSearch %v", c.URLs)
	es, err := elastic.NewClient(elastic.SetURL(c.URLs...))
	if err != nil {
		logger.Alertf("[writer] Can't connect to ElasticSearch: %v", err)
		return Writer{}, err
	}
	logger.Debug("[writer] Successfully connected to ElasticSearch")

	ESTemplate := `{"template":"` + c.Index + `*","mappings":{"raw":{"_source":{"enabled":false},"dynamic_templates":[{"fields":{"mapping":{"index":"not_analyzed","type":"string","copy_to":"@uniq"},"path_match":"fields.*"}}],"properties":{"@timestamp":{"type":"date","format":"strict_date_optional_time||epoch_millis"},"@uniq":{"type":"string","index":"not_analyzed"},"name":{"type":"string","index":"not_analyzed"},"value":{"type":"double","index":"not_analyzed"}}}}}`

	tmplExists, err := es.IndexTemplateExists(c.Index).Do()
	if err != nil {
		logger.Alertf("[writer] Error checking index mapping template existence: %v", err)
		return Writer{}, err
	}
	if !tmplExists {
		logger.Infof("[writer] Index mapping template doesn't exits, creating '%s'", c.Index)
		tmpl := es.IndexPutTemplate(c.Index).
			Create(true).
			BodyString(ESTemplate).
			Order(0)
		err := tmpl.Validate()
		if err != nil {
			logger.Alertf("[writer] Failed to validate the index mapping template: %v", err)
			return Writer{}, err
		}
		res, err := tmpl.Do()
		if err != nil {
			logger.Alertf("[writer] Failed to put the index mapping template: %v", err)
			return Writer{}, err
		}
		if !res.Acknowledged {
			logger.Error("[writer] Failed to acknowledge the new index mapping template")
			return Writer{}, err
		}
		logger.Info("[writer] New index mapping template acknowledged")
	}

	return Writer{
		Config:    c,
		ModuleWg:  module_wg,
		Transport: t,
		Elastic:   es,
		Logger:    logger,
		ExitFlag:  exitFlag,
		Stats:     NewWriterStats(),
	}, nil
}

func (w *Writer) Start() {
	w.ModuleWg.Add(1)
	defer w.ModuleWg.Done()
	w.Logger.Info("[writer] Starting writer module")

	exitTrigger := make(chan struct{}, 1)
	exitFinished := make(chan struct{}, 1)

	w.Logger.Debug("[writer] Setting up bulk-processor")
	var err interface{}
	w.Processor, err = elastic.NewBulkProcessorService(w.Elastic).
		Name("metcap").
		Workers(w.Config.Concurrency).
		BulkActions(w.Config.BulkMax).
		BulkSize(-1).
		Before(w.hookBeforeCommit).
		After(w.hookAfterCommit).
		FlushInterval(w.Config.BulkWait.Duration).
		Do()

	if err != nil {
		w.Logger.Alertf("[writer] Failed to setup bulk-processor: %v", err)
		return
	}

	w.Logger.Info("[writer] Writer module started")

	go func() {
		for {
			select {
			case metric, ok := <-w.Transport.OutputChan():
				if ok {
					w.add(metric)
				}
			case <-exitTrigger:
				w.Logger.Debug("[writer] Calling transport to stop retrieve loop...") // doesn't apply to channel transport
				w.Transport.CloseOutput()

				w.Logger.Info("[writer] Draining buffer...")
				drainingDone := make(chan struct{}, 1)

				// chan length checker (triggers)
				go func() {
					count, retries := 0, 10
					for count < retries {
						time.Sleep(500 * time.Millisecond)
						if w.Transport.OutputChanLen() == 0 {
							count++
							if count == retries {
								w.Logger.Debug("[writer] Buffer is empty")
							} else {
								w.Logger.Debug("[writer] Buffer seems empty, rechecking")
							}
						} else {
							w.Logger.Debug("[writer] Buffer isn't empty yet")
							count = 0
						}
					}
					drainingDone <- struct{}{}
				}()

				for {
					select {
					case <-drainingDone:
						w.Logger.Info("[writer] Draining done")
						w.Logger.Info("[writer] Flushing bulk-processors...")
						w.Processor.Close()
						exitFinished <- struct{}{}
						return
					case metric, ok := <-w.Transport.OutputChan():
						if ok {
							w.add(metric)
						}
					}
				}
			}
		}
	}()

	// shutdown handler
	for {
		if w.ExitFlag.Get() {
			w.Logger.Info("[writer] Stopping...")
			exitTrigger <- struct{}{}
			<-exitFinished
			w.Logger.Info("[writer] Stopped")
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func (w *Writer) add(m *Metric) {
	w.Stats.Queued.Increment(1)
	w.Processor.Add(elastic.NewBulkIndexRequest().
		Index(m.Index(w.Config.Index)).
		Type(w.Config.DocType).
		Doc(string(m.JSON())))
}

func (w *Writer) hookBeforeCommit(id int64, reqs []elastic.BulkableRequest) {
	w.Stats.Committed.Increment(len(reqs))
	w.Logger.Debugf("[writer] Committing %d metrics", len(reqs))
	w.Stats.Running.Increment(1)
	w.Stats.Queued.Reset()
}

func (w *Writer) hookAfterCommit(id int64, reqs []elastic.BulkableRequest, res *elastic.BulkResponse, err error) {
	w.Stats.Running.Decrement(1)
	w.Stats.Succeeded.Increment(len(res.Succeeded()))
	w.Stats.Duration.Add(time.Duration(res.Took) * time.Millisecond)
	w.Logger.Debugf("[writer] Successfully indexed %d metrics", len(res.Succeeded()))
	if len(res.Failed()) > 0 {
		w.Stats.Failed.Increment(len(res.Failed()))
		w.Logger.Errorf("[writer] Failed to index %d metrics", len(res.Failed()))
	}
	if err != nil {
		w.Logger.Errorf("[writer] %v", err.Error())
	}
	w.Stats.Flushed.Increment(1)
}

func (w *Writer) LogReport() {
	w.Logger.Infof("[writer] flushes: %d/%d/%.3f (running/total/rate_per_m), metrics: %d/%d/%d/%.3f (committed/succeeded/failed/rate_per_sec), duration: %s/%s (avg/max)",
		w.Stats.Running.Get(),
		w.Stats.Flushed.Total(),
		w.Stats.Flushed.Rate(time.Minute),
		w.Stats.Committed.Total(),
		w.Stats.Succeeded.Total(),
		w.Stats.Failed.Total(),
		w.Stats.Committed.Rate(time.Second),
		w.Stats.Duration.Avg(),
		w.Stats.Duration.Max(),
	)
}

type WriterStats struct {
	Running   *StatsGauge
	Flushed   *StatsCounter
	Committed *StatsCounter
	Succeeded *StatsCounter
	Failed    *StatsCounter
	Queued    *StatsCounter
	Duration  *StatsTimer
}

func NewWriterStats() *WriterStats {
	now := time.Now()
	return &WriterStats{
		Running:   NewStatsGauge(),
		Flushed:   NewStatsCounter(now),
		Committed: NewStatsCounter(now),
		Succeeded: NewStatsCounter(now),
		Failed:    NewStatsCounter(now),
		Queued:    NewStatsCounter(now),
		Duration:  NewStatsTimer(1000),
	}
}

func (s *WriterStats) Reset() {}

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
}

func NewWriter(c *WriterConfig, t Transport, module_wg *sync.WaitGroup, logger *Logger, exitFlag *Flag) (*Writer, error) {
	logger.Info("[writer] Initializing module")

	logger.Debugf("[writer] Connecting to ElasticSearch %v", c.URLs)
	es, err := elastic.NewClient(elastic.SetURL(c.URLs...))
	if err != nil {
		logger.Alertf("[writer] Can't connect to ElasticSearch: %v", err)
		return &Writer{}, err
	}
	logger.Debug("[writer] Successfully connected to ElasticSearch")

	ESTemplate := `{"template":"` + c.Index + `*","mappings":{"raw":{"_source":{"enabled":false},"dynamic_templates":[{"fields":{"mapping":{"index":"not_analyzed","type":"string","copy_to":"@uniq"},"path_match":"fields.*"}}],"properties":{"@timestamp":{"type":"date","format":"strict_date_optional_time||epoch_millis"},"@uniq":{"type":"string","index":"not_analyzed"},"name":{"type":"string","index":"not_analyzed"},"value":{"type":"double","index":"not_analyzed"}}}}}`

	tmplExists, err := es.IndexTemplateExists(c.Index).Do()
	if err != nil {
		logger.Alertf("[writer] Error checking index mapping template existence: %v", err)
		return &Writer{}, err
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
			return &Writer{}, err
		}
		res, err := tmpl.Do()
		if err != nil {
			logger.Alertf("[writer] Failed to put the index mapping template: %v", err)
			return &Writer{}, err
		}
		if !res.Acknowledged {
			logger.Error("[writer] Failed to acknowledge the new index mapping template")
		} else {
			logger.Info("[writer] New index mapping template acknowledged")
		}
	}

	return &Writer{
		Config:    c,
		ModuleWg:  module_wg,
		Transport: t,
		Elastic:   es,
		Logger:    logger,
		ExitFlag:  exitFlag,
	}, nil
}

func (w *Writer) Start() {
	w.ModuleWg.Add(1)
	w.Logger.Info("[writer] Starting writer module")
	defer w.ModuleWg.Done()

	exitChan := make(chan bool, 1)

	w.Logger.Debug("[writer] Setting up bulk-processor")
	var err interface{}
	w.Processor, err = w.Elastic.BulkProcessor().
		BulkActions(w.Config.BulkMax).
		BulkSize(-1).
		Before(w.hookBeforeCommit).
		After(w.hookAfterCommit).
		FlushInterval(time.Duration(w.Config.BulkWait) * time.Second).
		Name("metrics-capacitor").
		Stats(true).
		Workers(w.Config.Concurrency).Do()

	if err != nil {
		w.Logger.Alertf("[writer] Failed to setup bulk-processor: %v", err)
		return
	}

	w.Logger.Info("[writer] Writer module started")

	// shutdown handler
	go func() {
		for {
			switch {
			case w.ExitFlag.Get():
				w.Logger.Info("[writer] Stopping...")
				exitChan <- true
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	for {
		select {
		case metric := <-w.Transport.WriterChan():
			req := elastic.NewBulkIndexRequest().
				Index(metric.Index(w.Config.Index)).
				Type(w.Config.DocType).
				Doc(string(metric.JSON()))
			w.Processor.Add(req)
		case <-exitChan:
			w.Logger.Info("[writer] Flushing unwritten data...")
			w.Processor.Flush()
			w.Logger.Debug("[writer] Closing bulk-processor")
			w.Processor.Close()
			w.Logger.Info("[writer] Stopped")
			return
		}
	}
}

func (w *Writer) hookBeforeCommit(id int64, reqs []elastic.BulkableRequest) {
	w.Logger.Debugf("[writer] Committing %d metrics", len(reqs))
}

func (w *Writer) hookAfterCommit(id int64, reqs []elastic.BulkableRequest, res *elastic.BulkResponse, err error) {
	w.Logger.Infof("[writer] Successfully commited %d metrics", len(res.Succeeded()))
	if len(res.Failed()) > 0 {
		w.Logger.Errorf("[writer] Failed to commit %d metrics", len(res.Failed()))
	}
	if err != nil {
		w.Logger.Errorf("[writer] %v", err.Error())
	}
}

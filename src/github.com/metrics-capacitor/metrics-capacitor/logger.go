package metcap

import (
	"fmt"
	syslog "github.com/RackSec/srslog"
	"log"
	"os"
)

type Logger struct {
	cDebug chan string
	cInfo  chan string
	cErr   chan string
	cAlert chan string
	debug  *Flag
	syslog *bool
	logger *log.Logger
}

func NewLogger(syslog_enabled *bool, debugFlag *Flag) *Logger {
	var flags int = log.Ldate | log.Ltime | log.Lmicroseconds
	var logger *log.Logger
	var err error

	if *syslog_enabled {
		logger, err = syslog.NewLogger(syslog.LOG_INFO, flags)
		if err != nil {
			logger = log.New(os.Stdout, "", flags)
		}
	} else {
		logger = log.New(os.Stdout, "", flags)
	}
	return &Logger{
		cDebug: make(chan string),
		cInfo:  make(chan string),
		cErr:   make(chan string),
		cAlert: make(chan string),
		debug:  debugFlag,
		syslog: syslog_enabled,
		logger: logger}
}

func (l *Logger) Run() error {
	for {
		select {
		case line := <-l.cAlert:
			l.Log(line, syslog.LOG_ALERT)
		case line := <-l.cErr:
			l.Log(line, syslog.LOG_ERR)
		case line := <-l.cInfo:
			l.Log(line, syslog.LOG_INFO)
		case line := <-l.cDebug:
			l.Log(line, syslog.LOG_DEBUG)
		}
	}
}

func (l *Logger) Log(message string, level syslog.Priority) {
	if *l.syslog {
		// TODO
	} else {
		var txt_lvl string

		switch level {
		case syslog.LOG_DEBUG:
			txt_lvl = "DEBUG"
		case syslog.LOG_INFO:
			txt_lvl = " INFO"
		case syslog.LOG_ERR:
			txt_lvl = "ERROR"
		case syslog.LOG_ALERT:
			txt_lvl = "ALERT"
		}

		l.logger.Println(txt_lvl + ": " + message)
	}
}

func (l *Logger) Debug(m string) {
	if l.debug.Get() {
		l.cDebug <- m
	}
}
func (l *Logger) Debugf(f string, v ...interface{}) {
	if l.debug.Get() {
		l.cDebug <- fmt.Sprintf(f, v...)
	}
}
func (l *Logger) Info(m string)                     { l.cInfo <- m }
func (l *Logger) Infof(f string, v ...interface{})  { l.cInfo <- fmt.Sprintf(f, v...) }
func (l *Logger) Error(m string)                    { l.cErr <- m }
func (l *Logger) Errorf(f string, v ...interface{}) { l.cErr <- fmt.Sprintf(f, v...) }
func (l *Logger) Alert(m string)                    { l.cAlert <- m }
func (l *Logger) Alertf(f string, v ...interface{}) { l.cAlert <- fmt.Sprintf(f, v...) }

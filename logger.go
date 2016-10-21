package metcap

import (
	"fmt"
	"os"
	"time"
	"log"

	syslog "github.com/RackSec/srslog"
)

type Logger struct {
	chanDebug chan string
	chanInfo  chan string
	chanErr   chan string
	chanAlert chan string
	debug     *Flag
	syslog    bool
	syslogger *syslog.Writer
	logger    *log.Logger
}

func NewLogger(syslog_enabled *bool, debugFlag *Flag) *Logger {
	var (
		syslogger *syslog.Writer
		err       error
	)

	if *syslog_enabled {
		syslogger, err = syslog.Dial("", "", syslog.LOG_USER, "metcap")
		if err != nil {
			syslogger = nil
			*syslog_enabled = false
		}
	}
	return &Logger{
		chanDebug: make(chan string),
		chanInfo:  make(chan string),
		chanErr:   make(chan string),
		chanAlert: make(chan string),
		debug:     debugFlag,
		syslog:    *syslog_enabled,
		syslogger: syslogger,
		logger:		 log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) Run() error {
	for {
		select {
		case line := <-l.chanAlert:
			l.log(line, syslog.LOG_ALERT)
		case line := <-l.chanErr:
			l.log(line, syslog.LOG_ERR)
		case line := <-l.chanInfo:
			l.log(line, syslog.LOG_INFO)
		case line := <-l.chanDebug:
			l.log(line, syslog.LOG_DEBUG)
		}
	}
}

func (l *Logger) log(message string, severity syslog.Priority) {
	var txtSeverity string
	if l.syslog {
		l.syslogger.WriteWithPriority(severity, []byte(message + "\n"))
	} else {
		switch severity {
		case syslog.LOG_DEBUG:
			txtSeverity = " DEBUG: "
		case syslog.LOG_INFO:
			txtSeverity = "  INFO: "
		case syslog.LOG_ERR:
			txtSeverity = " ERROR: "
		case syslog.LOG_ALERT:
			txtSeverity = " ALERT: "
		}
		l.logger.Print(time.Now().Format(time.RFC3339) + txtSeverity + message + "\n")
	}
}

func (l *Logger) Debug(f string, v ...interface{}) {
	if l.debug.Get() {
		l.chanDebug <- fmt.Sprintf(f, v...)
	}
}
func (l *Logger) Info(f string, v ...interface{})  { l.chanInfo <- fmt.Sprintf(f, v...) }
func (l *Logger) Error(f string, v ...interface{}) { l.chanErr <- fmt.Sprintf(f, v...) }
func (l *Logger) Alert(f string, v ...interface{}) { l.chanAlert <- fmt.Sprintf(f, v...) }

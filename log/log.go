package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

// Logger generalizes the Logger types
type Logger interface {
	logrus.FieldLogger
}

// StructuredLogger provides the logger backend using Sirupsen/logrus
type StructuredLogger struct {
	*logrus.Logger
}

// New creates StructuredLogger and configure it
func New(format, level string, output io.Writer) *StructuredLogger {
	logger := logrus.New()

	logger.Out = output

	f := strings.ToLower(format)
	switch f {
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
	case "text":
		logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	default:
		logger.Warnf("log: invalid formatter: %v, continue with default", f)
	}

	l := strings.ToLower(level)
	sev, err := logrus.ParseLevel(l)
	if err != nil {
		logger.Warnf("log: invalid level: %v, continue with info", l)
		sev = logrus.InfoLevel
	}
	logger.Level = sev

	return &StructuredLogger{logger}
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
	logFields := logrus.Fields{}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.Logger = entry.Logger.WithFields(logFields)
	return entry
}

type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elasped_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
	})

	l.Logger.Infoln("request complete")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}

// Helper methods used by the application to get the request-scoped
// logger entry and set additional fields between handlers.
//
// This is a useful pattern to use to set state on the entry as it
// passes through the handler chain, which at any point can be logged
// with a call to .Print(), .Info(), etc.

func GetLogEntry(r *http.Request) Logger {
	l, ok := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	if ok {
		return l.Logger
	}
	// Discard log in case if logger is not attached.
	return New("text", "error", ioutil.Discard)
}

func LogEntrySetField(r *http.Request, key string, value interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.WithField(key, value)
	}
}

func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.WithFields(fields)
	}
}

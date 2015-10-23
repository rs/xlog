// Package xlog is a logger coupled with HTTP net/context aware middleware.
//
// Features:
//
//   - Per request log context
//   - Per request and/or per message key/value fields
//   - Log levels (Debug, Info, Warn, Error)
//   - Custom output
//   - Automatically gathers request context like User-Agent, IP etc.
//   - Drops message rather than blocking execution
//
// It works best in combinason with github.com/rs/xhandler.
package xlog // import "github.com/rs/xlog"

import (
	"fmt"
	"time"
)

// Logger is per request logger interface
type Logger interface {
	// SetField sets a field on the logger's context. All future messages on this logger
	// will have this field set.
	SetField(name string, value interface{})
	// Debug logs a debug message. If last parameter is a map[string]string, it's content
	// is added as fields to the message.
	Debug(v ...interface{})
	// Debug logs a debug message with format. If last parameter is a map[string]string,
	// it's content is added as fields to the message.
	Debugf(format string, v ...interface{})
	// Info logs a info message. If last parameter is a map[string]string, it's content
	// is added as fields to the message.
	Info(v ...interface{})
	// Info logs a info message with format. If last parameter is a map[string]string,
	// it's content is added as fields to the message.
	Infof(format string, v ...interface{})
	// Warn logs a warning message. If last parameter is a map[string]string, it's content
	// is added as fields to the message.
	Warn(v ...interface{})
	// Warn logs a warning message with format. If last parameter is a map[string]string,
	// it's content is added as fields to the message.
	Warnf(format string, v ...interface{})
	// Error logs an error message. If last parameter is a map[string]string, it's content
	// is added as fields to the message.
	Error(v ...interface{})
	// Error logs an error message with format. If last parameter is a map[string]string,
	// it's content is added as fields to the message.
	Errorf(format string, v ...interface{})
}

// F represents a set of log message fields string -> interface{}
type F map[string]interface{}

type logger struct {
	level  Level
	output chan map[string]interface{}
	fields map[string]interface{}
}

// Common key names for log messages
const (
	KeyTime      = "time"
	KeyMessage   = "message"
	KeyLevel     = "level"
	KeyIP        = "ip"
	KeyUserAgent = "user-agent"
)

var now = time.Now

func (l *logger) send(level Level, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}
	data := map[string]interface{}{
		KeyTime:    now(),
		KeyLevel:   level.String(),
		KeyMessage: msg,
	}
	for k, v := range fields {
		data[k] = v
	}
	for k, v := range l.fields {
		data[k] = v
	}
	select {
	case l.output <- data:
		// Sent with success
	default:
		// Channel is full, message dropped
	}
}

func extractFields(v *[]interface{}) map[string]interface{} {
	if l := len(*v); l > 0 {
		if f, ok := (*v)[l-1].(map[string]interface{}); ok {
			*v = (*v)[:l-1]
			return f
		}
		if f, ok := (*v)[l-1].(F); ok {
			*v = (*v)[:l-1]
			return f
		}
	}
	return nil
}

// SetField implements Logger interface
func (l *logger) SetField(name string, value interface{}) {
	if l.fields == nil {
		l.fields = map[string]interface{}{}
	}
	l.fields[name] = value
}

// Debug implements Logger interface
func (l *logger) Debug(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelDebug, fmt.Sprint(v...), f)
}

// Debugf implements Logger interface
func (l *logger) Debugf(format string, v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelDebug, fmt.Sprintf(format, v...), f)
}

// Info implements Logger interface
func (l *logger) Info(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelInfo, fmt.Sprint(v...), f)
}

// Infof implements Logger interface
func (l *logger) Infof(format string, v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelInfo, fmt.Sprintf(format, v...), f)
}

// Warn implements Logger interface
func (l *logger) Warn(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelWarn, fmt.Sprint(v...), f)
}

// Warnf implements Logger interface
func (l *logger) Warnf(format string, v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelWarn, fmt.Sprintf(format, v...), f)
}

// Error implements Logger interface
func (l *logger) Error(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelError, fmt.Sprint(v...), f)
}

// Errorf implements Logger interface
func (l *logger) Errorf(format string, v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelError, fmt.Sprintf(format, v...), f)
}

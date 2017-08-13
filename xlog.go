// Package xlog is a logger coupled with HTTP net/context aware middleware.
//
// Unlike most loggers, xlog will never block your application because one its
// outputs is lagging. The log commands are connected to their outputs through
// a buffered channel and will prefer to discard messages if the buffer get full.
// All message formatting, serialization and transport happen in a dedicated go
// routine.
//
// Features:
//
//     - Per request log context
//     - Per request and/or per message key/value fields
//     - Log levels (Debug, Info, Warn, Error)
//     - Color output when terminal is detected
//     - Custom output (JSON, logfmt, …)
//     - Automatic gathering of request context like User-Agent, IP etc.
//     - Drops message rather than blocking execution
//     - Easy access logging thru github.com/rs/xaccess
//
// It works best in combination with github.com/rs/xhandler.
package xlog // import "github.com/rs/xlog"

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Logger defines the interface for a xlog compatible logger
type Logger interface {
	// Implements io.Writer so it can be set a output of log.Logger
	io.Writer

	// SetField sets a field on the logger's context. All future messages on this logger
	// will have this field set.
	SetField(name string, value interface{})
	// GetFields returns all the fields set on the logger
	GetFields() F
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
	// Fatal logs an error message followed by a call to os.Exit(1). If last parameter is a
	// map[string]string, it's content is added as fields to the message.
	Fatal(v ...interface{})
	// Fatalf logs an error message with format followed by a call to ox.Exit(1). If last
	// parameter is a map[string]string, it's content is added as fields to the message.
	Fatalf(format string, v ...interface{})
	// Output mimics std logger interface
	Output(calldepth int, s string) error
	// OutputF outputs message with fields.
	OutputF(level Level, calldepth int, msg string, fields map[string]interface{})
}

// LoggerCopier defines a logger with copy support
type LoggerCopier interface {
	// Copy returns a copy of the logger
	Copy() Logger
}

// Config defines logger's configuration
type Config struct {
	// Level is the maximum level to output, logs with lower level are discarded.
	Level Level
	// Fields defines default fields to use with all messages.
	Fields map[string]interface{}
	// Output to use to write log messages to.
	//
	// You should always wrap your output with an OutputChannel otherwise your
	// logger will be connected to its output synchronously.
	Output Output
	// DisablePooling removes the use of a sync.Pool for cases where logger
	// instances are needed beyond the scope of a request handler. This option
	// puts a greater pressure on GC and increases the amount of memory allocated
	// and freed. Use only if persistent loggers are a requirement.
	DisablePooling bool
}

// F represents a set of log message fields
type F map[string]interface{}

type logger struct {
	level          Level
	output         Output
	fields         F
	disablePooling bool
}

// Common field names for log messages.
const (
	KeyTime    = "time"
	KeyMessage = "message"
	KeyLevel   = "level"
	KeyFile    = "file"
)

var now = time.Now
var exit1 = func() { os.Exit(1) }

// critialLogger is a logger to use when xlog is not able to deliver a message
var critialLogger = log.New(os.Stderr, "xlog: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

var loggerPool = &sync.Pool{
	New: func() interface{} {
		return &logger{}
	},
}

// New manually creates a logger.
//
// This function should only be used out of a request. Use FromContext in request.
func New(c Config) Logger {
	var l *logger
	if c.DisablePooling {
		l = &logger{}
	} else {
		l = loggerPool.Get().(*logger)
	}
	l.level = c.Level
	l.output = c.Output
	if l.output == nil {
		l.output = NewOutputChannel(NewConsoleOutput())
	}
	for k, v := range c.Fields {
		l.SetField(k, v)
	}
	l.disablePooling = c.DisablePooling
	return l
}

// Copy returns a copy of the passed logger if the logger implements
// LoggerCopier or the NopLogger otherwise.
func Copy(l Logger) Logger {
	if l, ok := l.(LoggerCopier); ok {
		return l.Copy()
	}
	return NopLogger
}

// Copy returns a copy of the logger
func (l *logger) Copy() Logger {
	l2 := &logger{
		level:          l.level,
		output:         l.output,
		fields:         map[string]interface{}{},
		disablePooling: l.disablePooling,
	}
	for k, v := range l.fields {
		l2.fields[k] = v
	}
	return l2
}

// close returns the logger to the pool for reuse
func (l *logger) close() {
	if !l.disablePooling {
		l.level = 0
		l.output = nil
		l.fields = nil
		loggerPool.Put(l)
	}
}

func (l *logger) send(level Level, calldepth int, msg string, fields map[string]interface{}) {
	if level < l.level || l.output == nil {
		return
	}
	data := make(map[string]interface{}, 4+len(fields)+len(l.fields))
	data[KeyTime] = now()
	data[KeyLevel] = level.String()
	data[KeyMessage] = msg
	if _, file, line, ok := runtime.Caller(calldepth); ok {
		data[KeyFile] = path.Base(file) + ":" + strconv.FormatInt(int64(line), 10)
	}
	for k, v := range fields {
		data[k] = v
	}
	if l.fields != nil {
		for k, v := range l.fields {
			data[k] = v
		}
	}
	if err := l.output.Write(data); err != nil {
		critialLogger.Print("send error: ", err.Error())
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

// GetFields implements Logger interface
func (l *logger) GetFields() F {
	return l.fields
}

// Output implements Logger interface
func (l *logger) OutputF(level Level, calldepth int, msg string, fields map[string]interface{}) {
	l.send(level, calldepth+1, msg, fields)
}

// Debug implements Logger interface
func (l *logger) Debug(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelDebug, 2, fmt.Sprint(v...), f)
}

// Debugf implements Logger interface
func (l *logger) Debugf(format string, v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelDebug, 2, fmt.Sprintf(format, v...), f)
}

// Info implements Logger interface
func (l *logger) Info(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelInfo, 2, fmt.Sprint(v...), f)
}

// Infof implements Logger interface
func (l *logger) Infof(format string, v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelInfo, 2, fmt.Sprintf(format, v...), f)
}

// Warn implements Logger interface
func (l *logger) Warn(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelWarn, 2, fmt.Sprint(v...), f)
}

// Warnf implements Logger interface
func (l *logger) Warnf(format string, v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelWarn, 2, fmt.Sprintf(format, v...), f)
}

// Error implements Logger interface
func (l *logger) Error(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelError, 2, fmt.Sprint(v...), f)
}

// Errorf implements Logger interface
//
// Go vet users: you may append %v at the end of you format when using xlog.F{} as a last
// argument to workaround go vet false alarm.
func (l *logger) Errorf(format string, v ...interface{}) {
	f := extractFields(&v)
	if f != nil {
		// Let user add a %v at the end of the message when fields are passed to satisfy go vet
		l := len(format)
		if l > 2 && format[l-2] == '%' && format[l-1] == 'v' {
			format = format[0 : l-2]
		}
	}
	l.send(LevelError, 2, fmt.Sprintf(format, v...), f)
}

// Fatal implements Logger interface
func (l *logger) Fatal(v ...interface{}) {
	f := extractFields(&v)
	l.send(LevelFatal, 2, fmt.Sprint(v...), f)
	if o, ok := l.output.(*OutputChannel); ok {
		o.Close()
	}
	exit1()
}

// Fatalf implements Logger interface
//
// Go vet users: you may append %v at the end of you format when using xlog.F{} as a last
// argument to workaround go vet false alarm.
func (l *logger) Fatalf(format string, v ...interface{}) {
	f := extractFields(&v)
	if f != nil {
		// Let user add a %v at the end of the message when fields are passed to satisfy go vet
		l := len(format)
		if l > 2 && format[l-2] == '%' && format[l-1] == 'v' {
			format = format[0 : l-2]
		}
	}
	l.send(LevelFatal, 2, fmt.Sprintf(format, v...), f)
	if o, ok := l.output.(*OutputChannel); ok {
		o.Close()
	}
	exit1()
}

// Write implements io.Writer interface
func (l *logger) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	l.send(LevelInfo, 4, msg, nil)
	if o, ok := l.output.(*OutputChannel); ok {
		o.Flush()
	}
	return len(p), nil
}

// Output implements common logger interface
func (l *logger) Output(calldepth int, s string) error {
	l.send(LevelInfo, 2, s, nil)
	return nil
}

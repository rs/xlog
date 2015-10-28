package xlog

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"log/syslog"
	"os"
	"strings"
	"sync"
)

// Output sends a log message fields to its destination
type Output interface {
	Write(fields map[string]interface{}) error
}

type discard struct{}

func (o discard) Write(fields map[string]interface{}) (err error) {
	return nil
}

// Discard discards log output
var Discard = &discard{}

var bufPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// MultiOutput routes the same message to serveral outputs.
// If one or more outputs return error, the last error is returned.
type MultiOutput []Output

func (m MultiOutput) Write(fields map[string]interface{}) (err error) {
	for _, o := range m {
		e := o.Write(fields)
		if e != nil {
			err = e
		}
	}
	return
}

// FilterOutput test a condition on the message and forward it to the child output
// if it returns true
type FilterOutput struct {
	Cond   func(fields map[string]interface{}) bool
	Output Output
}

func (f FilterOutput) Write(fields map[string]interface{}) (err error) {
	if f.Output == nil {
		return
	}
	if f.Cond(fields) {
		return f.Output.Write(fields)
	}
	return
}

// LevelOutput routes messages per level outputs
type LevelOutput struct {
	Debug Output
	Info  Output
	Warn  Output
	Error Output
}

func (l LevelOutput) Write(fields map[string]interface{}) error {
	var o Output
	switch fields[KeyLevel] {
	case "debug":
		o = l.Debug
	case "info":
		o = l.Info
	case "warn":
		o = l.Warn
	case "error":
		o = l.Error
	}
	if o != nil {
		return o.Write(fields)
	}
	return nil
}

// NewSyslogOutput returns JSONOutputs in a LevelOutput with writers set to syslog
// with the proper priority.
// If network and address are empty, Dial will connect to the local syslog server.
func NewSyslogOutput(network, address, tag string) Output {
	var err error
	o := LevelOutput{}
	if o.Debug, err = newJSONSyslogOutput(network, address, syslog.LOG_USER|syslog.LOG_DEBUG, tag); err != nil {
		log.Panicf("xlog: syslog error: %v", err)
	}
	if o.Info, err = newJSONSyslogOutput(network, address, syslog.LOG_USER|syslog.LOG_INFO, tag); err != nil {
		log.Panicf("xlog: syslog error: %v", err)
	}
	if o.Warn, err = newJSONSyslogOutput(network, address, syslog.LOG_USER|syslog.LOG_WARNING, tag); err != nil {
		log.Panicf("xlog: syslog error: %v", err)
	}
	if o.Error, err = newJSONSyslogOutput(network, address, syslog.LOG_USER|syslog.LOG_ERR, tag); err != nil {
		log.Panicf("xlog: syslog error: %v", err)
	}
	return o
}

func newJSONSyslogOutput(network, address string, prio syslog.Priority, tag string) (Output, error) {
	s, err := syslog.Dial(network, address, prio, tag)
	if err != nil {
		return nil, err
	}
	return NewJSONOutput(s), nil
}

// ConsoleOutput writes the message key if present followed by other fields in a
// given io.Writer.
type ConsoleOutput struct {
	w io.Writer
}

// NewConsoleOutput returns ConsoleOutputs in a LevelOutput with error levels on os.Stderr
// and other on os.Stdin
func NewConsoleOutput() Output {
	o := ConsoleOutput{w: os.Stdout}
	e := ConsoleOutput{w: os.Stderr}
	return LevelOutput{
		Debug: o,
		Info:  o,
		Warn:  e,
		Error: e,
	}
}

func (o ConsoleOutput) Write(fields map[string]interface{}) error {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	if msg, ok := fields[KeyMessage].(string); ok {
		delete(fields, KeyMessage)
		msg = strings.Replace(msg, "\n", "\\n", -1)
		buf.Write([]byte(msg + " "))
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	buf.Write(b)
	buf.WriteByte('\n')
	if _, err = o.w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

// JSONOutput marshals message fields and write the result on an io.Writer
type JSONOutput struct {
	w io.Writer
}

// NewJSONOutput returns a new JSONOutput with the given writer
func NewJSONOutput(w io.Writer) Output {
	return JSONOutput{w: w}
}

func (o JSONOutput) Write(fields map[string]interface{}) error {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	b, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	buf.Write(b)
	buf.WriteByte('\n')
	if _, err = o.w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

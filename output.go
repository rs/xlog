package xlog

import (
	"encoding/json"
	"io"
	"log"
	"log/syslog"
	"os"
	"strings"
)

// Output sends a log message fields to its destination
type Output interface {
	Write(fields map[string]interface{}) error
}

// MultiOutput routes the same message to serveral outputs
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

// FilterOutput tests the value of a key and pass the message to the child output
// only if the value of the defined key match one of the allowed values
type FilterOutput struct {
	Key    string
	Values []interface{}
	Invert bool
	Output Output
}

func (f FilterOutput) Write(fields map[string]interface{}) (err error) {
	if f.Output == nil || f.Key == "" || len(f.Values) == 0 {
		return
	}
	if val, found := fields[f.Key]; found {
		for _, allowed := range f.Values {
			if (val == allowed) != f.Invert {
				return f.Output.Write(fields)
			}
		}
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

// ConsoleOutput writes the message key if present followed by other fields in a
// given io.Writer.
type ConsoleOutput struct {
	w io.Writer
}

// NewConsoleOutput returns ConsoleOutputs in a LevelOutput with error levels on os.Stderr
// and other on os.Stdin
func NewConsoleOutput() Output {
	o := &ConsoleOutput{w: os.Stdout}
	e := &ConsoleOutput{w: os.Stderr}
	return &LevelOutput{
		Debug: o,
		Info:  o,
		Warn:  e,
		Error: e,
	}
}

// NewSyslogOutput returns JSONOutputs in a LevelOutput with writers set to syslog
// with the proper priority.
// If network and address are empty, Dial will connect to the local syslog server.
func NewSyslogOutput(network, address, tag string) Output {
	var err error
	o := &LevelOutput{}
	if o.Debug, err = newJSONSyslogOutput(network, address, syslog.LOG_DEBUG, tag); err != nil {
		log.Fatalf("xlog: syslog error: %v", err)
	}
	if o.Info, err = newJSONSyslogOutput(network, address, syslog.LOG_INFO, tag); err != nil {
		log.Fatalf("xlog: syslog error: %v", err)
	}
	if o.Warn, err = newJSONSyslogOutput(network, address, syslog.LOG_WARNING, tag); err != nil {
		log.Fatalf("xlog: syslog error: %v", err)
	}
	if o.Error, err = newJSONSyslogOutput(network, address, syslog.LOG_ERR, tag); err != nil {
		log.Fatalf("xlog: syslog error: %v", err)
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

func (o ConsoleOutput) Write(fields map[string]interface{}) error {
	if msg, ok := fields[KeyMessage].(string); ok {
		delete(fields, KeyMessage)
		msg = strings.Replace(msg, "\n", "\\n", 0)
		o.w.Write([]byte(msg + " "))
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	if _, err = o.w.Write(b); err != nil {
		return err
	}
	o.w.Write([]byte{'\n'})
	return nil
}

// JSONOutput marshals message fields and write the result on an io.Writer
type JSONOutput struct {
	w io.Writer
}

// NewJSONOutput returns a new JSONOutput with the given writer
func NewJSONOutput(w io.Writer) *JSONOutput {
	return &JSONOutput{w: w}
}

func (o JSONOutput) Write(fields map[string]interface{}) error {
	b, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	if _, err = o.w.Write(b); err != nil {
		return err
	}
	o.w.Write([]byte{'\n'})
	return nil
}

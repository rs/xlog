package xlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/xid"
	"github.com/rs/xlog/internal/term"
)

// Output sends a log message fields to its destination
type Output interface {
	Write(fields map[string]interface{}) error
}

// OutputChannel is a send channel between xlog and an Output
type OutputChannel struct {
	input chan map[string]interface{}
	stop  chan struct{}
}

// NewOutputChannel creates a consumer channel for the given output
func NewOutputChannel(o Output) *OutputChannel {
	oc := &OutputChannel{
		input: make(chan map[string]interface{}, 100),
		stop:  make(chan struct{}),
	}

	go func() {
		for {
			select {
			case msg := <-oc.input:
				if err := o.Write(msg); err != nil {
					log.Printf("xlog: cannot write log message: %v", err)
				}
			case <-oc.stop:
				close(oc.stop)
				return
			}
		}
	}()

	return oc
}

// Close closes the output channel
func (oc *OutputChannel) Close() {
	if oc.stop == nil {
		return
	}
	oc.stop <- struct{}{}
	<-oc.stop
	oc.stop = nil
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
// with the proper priority added to a LOG_USER facility.
// If network and address are empty, Dial will connect to the local syslog server.
func NewSyslogOutput(network, address, tag string) Output {
	return NewSyslogOutputFacility(network, address, tag, syslog.LOG_USER)
}

// NewSyslogOutputFacility returns JSONOutputs in a LevelOutput with writers set to syslog
// with the proper priority added to the passed facility.
// If network and address are empty, Dial will connect to the local syslog server.
func NewSyslogOutputFacility(network, address, tag string, facility syslog.Priority) Output {
	var err error
	o := LevelOutput{}
	if o.Debug, err = newJSONSyslogOutput(network, address, facility|syslog.LOG_DEBUG, tag); err != nil {
		log.Panicf("xlog: syslog error: %v", err)
	}
	if o.Info, err = newJSONSyslogOutput(network, address, facility|syslog.LOG_INFO, tag); err != nil {
		log.Panicf("xlog: syslog error: %v", err)
	}
	if o.Warn, err = newJSONSyslogOutput(network, address, facility|syslog.LOG_WARNING, tag); err != nil {
		log.Panicf("xlog: syslog error: %v", err)
	}
	if o.Error, err = newJSONSyslogOutput(network, address, facility|syslog.LOG_ERR, tag); err != nil {
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

const (
	red    = 31
	green  = 32
	yellow = 33
	blue   = 34
	gray   = 37
)

type consoleOutput struct {
	w     io.Writer
	color bool
}

// NewConsoleOutput returns a Output printing message in a human readable form on the
// stdout.
func NewConsoleOutput() Output {
	return consoleOutput{w: os.Stdout, color: term.IsTerminal(os.Stdout)}
}

func (o consoleOutput) Write(fields map[string]interface{}) error {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	if ts, ok := fields[KeyTime].(time.Time); ok {
		buf.Write([]byte(ts.Format("2006/01/02 15:04:05 ")))
	}
	if lvl, ok := fields[KeyLevel].(string); ok {
		levelColor := blue
		switch lvl {
		case "debug":
			levelColor = gray
		case "warn":
			levelColor = yellow
		case "error":
			levelColor = red
		}
		colorPrint(buf, strings.ToUpper(lvl[0:4]), levelColor, o.color)
		buf.WriteByte(' ')
	}
	if msg, ok := fields[KeyMessage].(string); ok {
		msg = strings.Replace(msg, "\n", "\\n", -1)
		buf.Write([]byte(msg))
	}
	for k, v := range fields {
		switch k {
		case KeyTime, KeyLevel, KeyMessage:
			continue
		}
		buf.WriteByte(' ')
		colorPrint(buf, k, green, o.color)
		buf.WriteByte('=')
		if err := writeValue(buf, v); err != nil {
			return err
		}
	}
	buf.WriteByte('\n')
	_, err := o.w.Write(buf.Bytes())
	return err
}

func colorPrint(w io.Writer, s string, color int, enabled bool) {
	if enabled {
		//w.Write([]byte{0x1b, '[', byte('0' + color), 'm'})
		fmt.Fprintf(w, "\x1b[%dm", color)
		w.Write([]byte(s))
		w.Write([]byte("\x1b[0m"))
	} else {
		w.Write([]byte(s))
	}
}

func needsQuotedValueRune(r rune) bool {
	return r <= ' ' || r == '=' || r == '"'
}

func writeValue(w io.Writer, v interface{}) (err error) {
	switch v := v.(type) {
	case string:
		if strings.IndexFunc(v, needsQuotedValueRune) != -1 {
			var b []byte
			b, err = json.Marshal(v)
			if err == nil {
				w.Write(b)
			}
		} else {
			_, err = w.Write([]byte(v))
		}
	case error:
		s := v.Error()
		err = writeValue(w, s)
	default:
		s := fmt.Sprint(v)
		err = writeValue(w, s)
	}
	return
}

// jsonOutput marshals message fields and write the result on an io.Writer
type jsonOutput struct {
	w io.Writer
}

// NewJSONOutput returns a new JSONOutput with the given writer
func NewJSONOutput(w io.Writer) Output {
	return jsonOutput{w: w}
}

func (o jsonOutput) Write(fields map[string]interface{}) error {
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
	_, err = o.w.Write(buf.Bytes())
	return err
}

// uidOutput adds a unique id field to all message transiting thru this output filter.
type uidOutput struct {
	f string
	o Output
}

func (o uidOutput) Write(fields map[string]interface{}) error {
	fields[o.f] = xid.New().String()
	return o.o.Write(fields)
}

// NewUIDOutput returns an output filter adding a globally unique id (using github.com/rs/xid)
// to all message going thru this output. The o parameter defines the next output to pass data
// to.
func NewUIDOutput(field string, o Output) Output {
	return &uidOutput{f: field, o: o}
}

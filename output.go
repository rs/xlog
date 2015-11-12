package xlog

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"log/syslog"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/xid"
	"github.com/rs/xlog/internal/term"
)

// Output sends a log message fields to a destination.
type Output interface {
	Write(fields map[string]interface{}) error
}

// OutputChannel is a send buffered channel between xlog and an Output.
type OutputChannel struct {
	input chan map[string]interface{}
	stop  chan struct{}
}

// NewOutputChannel creates a consumer buffered channel for the given output
// with a default buffer of 100 messages.
func NewOutputChannel(o Output) *OutputChannel {
	return NewOutputChannelBuffer(o, 100)
}

// NewOutputChannelBuffer creates a consumer buffered channel for the given output
// with a customizable buffer size.
func NewOutputChannelBuffer(o Output, bufSize int) *OutputChannel {
	oc := &OutputChannel{
		input: make(chan map[string]interface{}, bufSize),
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

// Write implements the Output interface
func (oc *OutputChannel) Write(fields map[string]interface{}) (err error) {
	select {
	case oc.input <- fields:
		// Sent with success
	default:
		// Channel is full, message dropped
	}
	return nil
}

// Close closes the output channel and release the consumer's go routine.
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

// Discard is an Output that discards all log message going thru it.
var Discard = &discard{}

var bufPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// MultiOutput routes the same message to serveral outputs.
// If one or more outputs return an error, the last error is returned.
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
// if it returns true.
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

// LevelOutput routes messages to different output based on the message's level.
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
	o := LevelOutput{
		Debug: NewJSONOutput(NewSyslogWriter(network, address, facility|syslog.LOG_DEBUG, tag)),
		Info:  NewJSONOutput(NewSyslogWriter(network, address, facility|syslog.LOG_INFO, tag)),
		Warn:  NewJSONOutput(NewSyslogWriter(network, address, facility|syslog.LOG_WARNING, tag)),
		Error: NewJSONOutput(NewSyslogWriter(network, address, facility|syslog.LOG_ERR, tag)),
	}
	return o
}

// NewSyslogWriter returns a writer ready to be used with output modules.
// If network and address are empty, Dial will connect to the local syslog server.
func NewSyslogWriter(network, address string, prio syslog.Priority, tag string) io.Writer {
	s, err := syslog.Dial(network, address, prio, tag)
	if err != nil {
		log.Panicf("xlog: syslog dial error: %v", err)
	}
	return s
}

type consoleOutput struct {
	w io.Writer
}

var isTerminal = term.IsTerminal

// NewConsoleOutput returns a Output printing message in a colored human readable form on the
// stdout.
func NewConsoleOutput() Output {
	if isTerminal(os.Stdout) {
		return consoleOutput{w: os.Stdout}
	}
	return NewLogfmtOutput(os.Stdout)
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
		colorPrint(buf, strings.ToUpper(lvl[0:4]), levelColor)
		buf.WriteByte(' ')
	}
	if msg, ok := fields[KeyMessage].(string); ok {
		msg = strings.Replace(msg, "\n", "\\n", -1)
		buf.Write([]byte(msg))
	}
	// Gather field keys
	keys := []string{}
	for k := range fields {
		switch k {
		case KeyLevel, KeyMessage, KeyTime:
			continue
		}
		keys = append(keys, k)
	}
	// Sort fields by key names
	sort.Strings(keys)
	// Print fields using logfmt format
	for _, k := range keys {
		buf.WriteByte(' ')
		colorPrint(buf, k, green)
		buf.WriteByte('=')
		if err := writeValue(buf, fields[k]); err != nil {
			return err
		}
	}
	buf.WriteByte('\n')
	_, err := o.w.Write(buf.Bytes())
	return err
}

type logfmtOutput struct {
	w io.Writer
}

// NewLogfmtOutput returns a new output using logstash JSON schema v1
func NewLogfmtOutput(w io.Writer) Output {
	return logfmtOutput{w: w}
}

func (o logfmtOutput) Write(fields map[string]interface{}) error {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	// Gather field keys
	keys := []string{}
	for k := range fields {
		switch k {
		case KeyLevel, KeyMessage, KeyTime:
			continue
		}
		keys = append(keys, k)
	}
	// Sort fields by key names
	sort.Strings(keys)
	// Prepend default fields in a specific order
	keys = append([]string{KeyLevel, KeyMessage, KeyTime}, keys...)
	l := len(keys)
	for i, k := range keys {
		buf.Write([]byte(k))
		buf.WriteByte('=')
		if err := writeValue(buf, fields[k]); err != nil {
			return err
		}
		if i+1 < l {
			buf.WriteByte(' ')
		} else {
			buf.WriteByte('\n')
		}
	}
	_, err := o.w.Write(buf.Bytes())
	return err
}

type jsonOutput struct {
	w io.Writer
}

// NewJSONOutput returns a new JSON output with the given writer.
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

type logstashOutput struct {
	w io.Writer
}

// NewLogstashOutput returns an output to generate logstash friendly JSON format.
func NewLogstashOutput(w io.Writer) Output {
	return logstashOutput{w: w}
}

func (o logstashOutput) Write(fields map[string]interface{}) error {
	lsf := map[string]interface{}{
		"@version": 1,
	}
	for k, v := range fields {
		switch k {
		case KeyTime:
			k = "@timestamp"
		case KeyLevel:
			if s, ok := v.(string); ok {
				v = strings.ToUpper(s)
			}
		case KeyFile:
			if s, ok := v.(string); ok {
				if i := strings.IndexByte(s, ':'); i != -1 {
					if n, err := strconv.ParseInt(s[i+1:], 10, 32); err == nil {
						lsf["line_number"] = n
						k = "file"
						v = s[:i]
					}
				}
			}
		}
		if t, ok := v.(time.Time); ok {
			lsf[k] = t.Format(time.RFC3339)
		} else {
			lsf[k] = v
		}
	}
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	b, err := json.Marshal(lsf)
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

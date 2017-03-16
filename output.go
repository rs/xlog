package xlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sort"
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

// OutputFunc is an adapter to allow the use of ordinary functions as Output handlers.
// If it is a function with the appropriate signature, OutputFunc(f) is a Output object
// that calls f on Write().
type OutputFunc func(fields map[string]interface{}) error

func (of OutputFunc) Write(fields map[string]interface{}) error {
	return of(fields)
}

// OutputChannel is a send buffered channel between xlog and an Output.
type OutputChannel struct {
	input  chan map[string]interface{}
	output Output
	stop   chan struct{}
}

// ErrBufferFull is returned when the output channel buffer is full and messages
// are discarded.
var ErrBufferFull = errors.New("buffer full")

// NewOutputChannel creates a consumer buffered channel for the given output
// with a default buffer of 100 messages.
func NewOutputChannel(o Output) *OutputChannel {
	return NewOutputChannelBuffer(o, 100)
}

// NewOutputChannelBuffer creates a consumer buffered channel for the given output
// with a customizable buffer size.
func NewOutputChannelBuffer(o Output, bufSize int) *OutputChannel {
	oc := &OutputChannel{
		input:  make(chan map[string]interface{}, bufSize),
		output: o,
		stop:   make(chan struct{}),
	}

	go func() {
		for {
			select {
			case msg := <-oc.input:
				if err := o.Write(msg); err != nil {
					critialLogger.Print("cannot write log message: ", err.Error())
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
		err = ErrBufferFull
	}
	return err
}

// Flush flushes all the buffered message to the output
func (oc *OutputChannel) Flush() {
	for {
		select {
		case msg := <-oc.input:
			if err := oc.output.Write(msg); err != nil {
				critialLogger.Print("cannot write log message: ", err.Error())
			}
		default:
			return
		}
	}
}

// Close closes the output channel and release the consumer's go routine.
func (oc *OutputChannel) Close() {
	if oc.stop == nil {
		return
	}
	oc.stop <- struct{}{}
	<-oc.stop
	oc.stop = nil
	oc.Flush()
}

// Discard is an Output that discards all log message going thru it.
var Discard = OutputFunc(func(fields map[string]interface{}) error {
	return nil
})

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
	Fatal Output
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
	case "fatal":
		o = l.Fatal
	}
	if o != nil {
		return o.Write(fields)
	}
	return nil
}

// RecorderOutput stores the raw messages in it's Messages field. This output is useful for testing.
type RecorderOutput struct {
	Messages []F
}

func (l *RecorderOutput) Write(fields map[string]interface{}) error {
	if l.Messages == nil {
		l.Messages = []F{fields}
	} else {
		l.Messages = append(l.Messages, fields)
	}
	return nil
}

// Reset empty the output from stored messages
func (l *RecorderOutput) Reset() {
	l.Messages = []F{}
}

type consoleOutput struct {
	w io.Writer
}

var isTerminal = term.IsTerminal

// NewConsoleOutput returns a Output printing message in a colored human readable form on the
// stderr. If the stderr is not on a terminal, a LogfmtOutput is returned instead.
func NewConsoleOutput() Output {
	return NewConsoleOutputW(os.Stderr, NewLogfmtOutput(os.Stderr))
}

// NewConsoleOutputW returns a Output printing message in a colored human readable form with
// the provided writer. If the writer is not on a terminal, the noTerm output is returned.
func NewConsoleOutputW(w io.Writer, noTerm Output) Output {
	if isTerminal(w) {
		return consoleOutput{w: w}
	}
	return noTerm
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

// NewJSONOutput returns a new JSON output with the given writer.
func NewJSONOutput(w io.Writer) Output {
	enc := json.NewEncoder(w)
	return OutputFunc(func(fields map[string]interface{}) error {
		return enc.Encode(fields)
	})
}

// NewLogstashOutput returns an output to generate logstash friendly JSON format.
func NewLogstashOutput(w io.Writer) Output {
	return OutputFunc(func(fields map[string]interface{}) error {
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
			}
			if t, ok := v.(time.Time); ok {
				lsf[k] = t.Format(time.RFC3339)
			} else {
				lsf[k] = v
			}
		}
		b, err := json.Marshal(lsf)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	})
}

// NewUIDOutput returns an output filter adding a globally unique id (using github.com/rs/xid)
// to all message going thru this output. The o parameter defines the next output to pass data
// to.
func NewUIDOutput(field string, o Output) Output {
	return OutputFunc(func(fields map[string]interface{}) error {
		fields[field] = xid.New().String()
		return o.Write(fields)
	})
}

// NewTrimOutput trims any field of type string with a value length greater than maxLen
// to maxLen.
func NewTrimOutput(maxLen int, o Output) Output {
	return OutputFunc(func(fields map[string]interface{}) error {
		for k, v := range fields {
			if s, ok := v.(string); ok && len(s) > maxLen {
				fields[k] = s[:maxLen]
			}
		}
		return o.Write(fields)
	})
}

// NewTrimFieldsOutput trims listed field fields of type string with a value length greater than maxLen
// to maxLen.
func NewTrimFieldsOutput(trimFields []string, maxLen int, o Output) Output {
	return OutputFunc(func(fields map[string]interface{}) error {
		for _, f := range trimFields {
			if s, ok := fields[f].(string); ok && len(s) > maxLen {
				fields[f] = s[:maxLen]
			}
		}
		return o.Write(fields)
	})
}

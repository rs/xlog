package xlog

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testOutput struct {
	err error
	w   chan map[string]interface{}
}

func newTestOutput() *testOutput {
	return &testOutput{w: make(chan map[string]interface{}, 10)}
}

func newTestOutputErr(err error) *testOutput {
	return &testOutput{w: make(chan map[string]interface{}, 10), err: err}
}

func (o *testOutput) Write(fields map[string]interface{}) (err error) {
	o.w <- fields
	return o.err
}

func (o *testOutput) reset() {
	o.w = make(chan map[string]interface{}, 10)
}

func (o *testOutput) empty() bool {
	select {
	case <-o.w:
		return false
	default:
		return true
	}
}

func (o *testOutput) get() map[string]interface{} {
	select {
	case last := <-o.w:
		return last
	case <-time.After(2 * time.Second):
		return nil
	}
}

func TestOutputChannel(t *testing.T) {
	o := newTestOutput()
	oc := NewOutputChannel(o)
	defer oc.Close()
	oc.input <- F{"foo": "bar"}
	assert.Equal(t, F{"foo": "bar"}, F(o.get()))
}

func TestOutputChannelError(t *testing.T) {
	// Trigger error path
	r, w := io.Pipe()
	go func() {
		critialLoggerMux.Lock()
		defer critialLoggerMux.Unlock()
		oldCritialLogger := critialLogger
		critialLogger = log.New(w, "", 0)
		o := newTestOutputErr(errors.New("some error"))
		oc := NewOutputChannel(o)
		oc.input <- F{"foo": "bar"}
		o.get()
		oc.Close()
		critialLogger = oldCritialLogger
		w.Close()
	}()
	b, err := ioutil.ReadAll(r)
	assert.NoError(t, err)
	assert.Contains(t, string(b), "cannot write log message: some error")
}

func TestOutputChannelClose(t *testing.T) {
	oc := NewOutputChannel(newTestOutput())
	defer oc.Close()
	assert.NotNil(t, oc.stop)
	oc.Close()
	assert.Nil(t, oc.stop)
	oc.Close()
}

func TestDiscard(t *testing.T) {
	assert.NoError(t, Discard.Write(F{}))
}

func TestMultiOutput(t *testing.T) {
	o1 := newTestOutput()
	o2 := newTestOutput()
	mo := MultiOutput{o1, o2}
	err := mo.Write(F{"foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, F{"foo": "bar"}, F(<-o1.w))
	assert.Equal(t, F{"foo": "bar"}, F(<-o2.w))
}

func TestMultiOutputWithError(t *testing.T) {
	o1 := newTestOutputErr(errors.New("some error"))
	o2 := newTestOutput()
	mo := MultiOutput{o1, o2}
	err := mo.Write(F{"foo": "bar"})
	assert.EqualError(t, err, "some error")
	// Still send data to all outputs
	assert.Equal(t, F{"foo": "bar"}, F(<-o1.w))
	assert.Equal(t, F{"foo": "bar"}, F(<-o2.w))
}

func TestFilterOutput(t *testing.T) {
	o := newTestOutput()
	f := FilterOutput{
		Cond: func(fields map[string]interface{}) bool {
			return fields["foo"] == "bar"
		},
		Output: o,
	}
	err := f.Write(F{"foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, F{"foo": "bar"}, F(o.get()))

	o.reset()
	err = f.Write(F{"foo": "baz"})
	assert.NoError(t, err)
	assert.True(t, o.empty())

	f.Output = nil
	err = f.Write(F{"foo": "baz"})
	assert.NoError(t, err)
}

func TestLevelOutput(t *testing.T) {
	oInfo := newTestOutput()
	oError := newTestOutput()
	oFatal := newTestOutput()
	oWarn := &testOutput{err: errors.New("some error")}
	reset := func() {
		oInfo.reset()
		oError.reset()
		oFatal.reset()
		oWarn.reset()
	}
	l := LevelOutput{
		Info:  oInfo,
		Error: oError,
		Fatal: oFatal,
		Warn:  oWarn,
	}

	err := l.Write(F{"level": "fatal", "foo": "bar"})
	assert.NoError(t, err)
	assert.True(t, oInfo.empty())
	assert.True(t, oError.empty())
	assert.Equal(t, F{"level": "fatal", "foo": "bar"}, F(<-oFatal.w))
	assert.True(t, oWarn.empty())

	reset()
	err = l.Write(F{"level": "error", "foo": "bar"})
	assert.NoError(t, err)
	assert.True(t, oInfo.empty())
	assert.Equal(t, F{"level": "error", "foo": "bar"}, F(<-oError.w))
	assert.True(t, oFatal.empty())
	assert.True(t, oWarn.empty())

	reset()
	err = l.Write(F{"level": "info", "foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, F{"level": "info", "foo": "bar"}, F(<-oInfo.w))
	assert.True(t, oFatal.empty())
	assert.True(t, oError.empty())
	assert.True(t, oWarn.empty())

	reset()
	err = l.Write(F{"level": "warn", "foo": "bar"})
	assert.EqualError(t, err, "some error")
	assert.True(t, oInfo.empty())
	assert.True(t, oError.empty())
	assert.True(t, oFatal.empty())
	assert.Equal(t, F{"level": "warn", "foo": "bar"}, F(<-oWarn.w))

	reset()
	err = l.Write(F{"level": "debug", "foo": "bar"})
	assert.NoError(t, err)
	assert.True(t, oInfo.empty())
	assert.True(t, oError.empty())
	assert.True(t, oFatal.empty())
	assert.True(t, oWarn.empty())

	reset()
	err = l.Write(F{"foo": "bar"})
	assert.NoError(t, err)
	assert.True(t, oInfo.empty())
	assert.True(t, oError.empty())
	assert.True(t, oFatal.empty())
	assert.True(t, oWarn.empty())
}

func TestSyslogOutput(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	critialLoggerMux.Lock()
	oldCritialLogger := critialLogger
	critialLogger = log.New(buf, "", 0)
	defer func() {
		critialLogger = oldCritialLogger
		critialLoggerMux.Unlock()
	}()
	m := NewSyslogOutput("udp", "127.0.0.1:1234", "mytag")
	assert.IsType(t, LevelOutput{}, m)
	assert.Panics(t, func() {
		NewSyslogOutput("tcp", "an invalid host name", "mytag")
	})
	assert.Regexp(t, "syslog dial error: dial tcp:.*missing port in address.*", buf.String())
}

func TestRecorderOutput(t *testing.T) {
	o := RecorderOutput{}
	o.Write(F{"foo": "bar"})
	o.Write(F{"bar": "baz"})
	assert.Equal(t, []F{{"foo": "bar"}, {"bar": "baz"}}, o.Messages)
	o.Reset()
	assert.Equal(t, []F{}, o.Messages)
}

func TestNewConsoleOutput(t *testing.T) {
	old := isTerminal
	defer func() { isTerminal = old }()
	isTerminal = func(w io.Writer) bool { return true }
	c := NewConsoleOutput()
	if assert.IsType(t, consoleOutput{}, c) {
		assert.Equal(t, os.Stderr, c.(consoleOutput).w)
	}
	isTerminal = func(w io.Writer) bool { return false }
	c = NewConsoleOutput()
	if assert.IsType(t, logfmtOutput{}, c) {
		assert.Equal(t, os.Stderr, c.(logfmtOutput).w)
	}
}

func TestNewConsoleOutputW(t *testing.T) {
	b := bytes.NewBuffer([]byte{})
	c := NewConsoleOutputW(b, NewLogfmtOutput(b))
	assert.IsType(t, logfmtOutput{}, c)
	old := isTerminal
	defer func() { isTerminal = old }()
	isTerminal = func(w io.Writer) bool { return true }
	c = NewConsoleOutputW(b, NewLogfmtOutput(b))
	if assert.IsType(t, consoleOutput{}, c) {
		assert.Equal(t, b, c.(consoleOutput).w)
	}
}

func TestConsoleOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	c := consoleOutput{w: buf}
	err := c.Write(F{"message": "some message", "level": "info", "time": time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC), "foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "2000/01/02 03:04:05 \x1b[34mINFO\x1b[0m some message \x1b[32mfoo\x1b[0m=bar\n", buf.String())
	buf.Reset()
	err = c.Write(F{"message": "some debug", "level": "debug"})
	assert.NoError(t, err)
	assert.Equal(t, "\x1b[37mDEBU\x1b[0m some debug\n", buf.String())
	buf.Reset()
	err = c.Write(F{"message": "some warning", "level": "warn"})
	assert.NoError(t, err)
	assert.Equal(t, "\x1b[33mWARN\x1b[0m some warning\n", buf.String())
	buf.Reset()
	err = c.Write(F{"message": "some error", "level": "error"})
	assert.NoError(t, err)
	assert.Equal(t, "\x1b[31mERRO\x1b[0m some error\n", buf.String())
}

func TestLogfmtOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	c := NewLogfmtOutput(buf)
	err := c.Write(F{
		"time":    time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		"message": "some message",
		"level":   "info",
		"string":  "foo",
		"null":    nil,
		"quoted":  "needs \" quotes",
		"err":     errors.New("error"),
		"errq":    errors.New("error with \" quote"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "level=info message=\"some message\" time=\"2000-01-02 03:04:05 +0000 UTC\" err=error errq=\"error with \\\" quote\" null=null quoted=\"needs \\\" quotes\" string=foo\n", buf.String())
}

func TestJSONOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	j := NewJSONOutput(buf)
	err := j.Write(F{"message": "some message", "level": "info", "foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "{\"foo\":\"bar\",\"level\":\"info\",\"message\":\"some message\"}\n", buf.String())
}

func TestLogstashOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	o := NewLogstashOutput(buf)
	err := o.Write(F{
		"message": "some message",
		"level":   "info",
		"time":    time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		"file":    "test.go:234",
		"foo":     "bar",
	})
	assert.NoError(t, err)
	assert.Equal(t, "{\"@timestamp\":\"2000-01-02T03:04:05Z\",\"@version\":1,\"file\":\"test.go:234\",\"foo\":\"bar\",\"level\":\"INFO\",\"message\":\"some message\"}", buf.String())
}

func TestUIDOutput(t *testing.T) {
	o := newTestOutput()
	i := NewUIDOutput("id", o)
	err := i.Write(F{"message": "some message", "level": "info", "foo": "bar"})
	last := o.get()
	assert.NoError(t, err)
	assert.NotNil(t, last["id"])
	assert.Len(t, last["id"], 20)
}

func TestTrimOutput(t *testing.T) {
	o := newTestOutput()
	i := NewTrimOutput(10, o)
	err := i.Write(F{"short": "short", "long": "too long message", "number": 20})
	last := o.get()
	assert.NoError(t, err)
	assert.Equal(t, "short", last["short"])
	assert.Equal(t, "too long m", last["long"])
	assert.Equal(t, 20, last["number"])
}

func TestTrimFieldsOutput(t *testing.T) {
	o := newTestOutput()
	i := NewTrimFieldsOutput([]string{"short", "trim", "number"}, 10, o)
	err := i.Write(F{"short": "short", "long": "too long message", "trim": "too long message", "number": 20})
	last := o.get()
	assert.NoError(t, err)
	assert.Equal(t, "short", last["short"])
	assert.Equal(t, "too long m", last["trim"])
	assert.Equal(t, "too long message", last["long"])
	assert.Equal(t, 20, last["number"])
}

package xlog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testOutput struct {
	err  error
	last map[string]interface{}
}

func (o *testOutput) Write(fields map[string]interface{}) (err error) {
	o.last = fields
	return o.err
}

func (o *testOutput) reset() {
	o.last = nil
	o.err = nil
}

func TestOutputChannel(t *testing.T) {
	o := &testOutput{}
	oc := NewOutputChannel(o)
	defer oc.Close()
	oc.input <- F{"foo": "bar"}
	assert.Nil(t, o.last)
	runtime.Gosched()
	assert.Equal(t, F{"foo": "bar"}, F(o.last))

	// Trigger error path
	buf := bytes.NewBuffer(nil)
	oldCritialLogger := critialLogger
	defer func() { critialLogger = oldCritialLogger }()
	critialLogger = func(v ...interface{}) {
		fmt.Fprint(buf, v...)
	}
	o.err = errors.New("some error")
	oc.input <- F{"foo": "bar"}
	// Wait for log output to go through
	runtime.Gosched()
	for i := 0; i < 10 && buf.Len() == 0; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	assert.Contains(t, buf.String(), "cannot write log message: some error")
}

func TestOutputChannelClose(t *testing.T) {
	oc := NewOutputChannel(&testOutput{})
	assert.NotNil(t, oc.stop)
	oc.Close()
	assert.Nil(t, oc.stop)
	oc.Close()
}

func TestDiscard(t *testing.T) {
	assert.NoError(t, Discard.Write(F{}))
}

func TestMultiOutput(t *testing.T) {
	o1 := &testOutput{}
	o2 := &testOutput{}
	mo := MultiOutput{o1, o2}
	err := mo.Write(F{"foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, F{"foo": "bar"}, F(o1.last))
	assert.Equal(t, F{"foo": "bar"}, F(o2.last))
}

func TestMultiOutputWithError(t *testing.T) {
	o1 := &testOutput{}
	o2 := &testOutput{}
	o1.err = errors.New("some error")
	mo := MultiOutput{o1, o2}
	err := mo.Write(F{"foo": "bar"})
	assert.EqualError(t, err, "some error")
	// Still send data to all outputs
	assert.Equal(t, F{"foo": "bar"}, F(o1.last))
	assert.Equal(t, F{"foo": "bar"}, F(o2.last))
}

func TestFilterOutput(t *testing.T) {
	o := &testOutput{}
	f := FilterOutput{
		Cond: func(fields map[string]interface{}) bool {
			return fields["foo"] == "bar"
		},
		Output: o,
	}
	err := f.Write(F{"foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, F{"foo": "bar"}, F(o.last))

	o.last = nil
	err = f.Write(F{"foo": "baz"})
	assert.NoError(t, err)
	assert.Nil(t, o.last)

	f.Output = nil
	err = f.Write(F{"foo": "baz"})
	assert.NoError(t, err)
}

func TestLevelOutput(t *testing.T) {
	oInfo := &testOutput{}
	oError := &testOutput{}
	oFatal := &testOutput{}
	oWarn := &testOutput{err: errors.New("some error")}
	reset := func() {
		oInfo.reset()
		oError.reset()
		oFatal.reset()
		oWarn.last = nil
	}
	l := LevelOutput{
		Info:  oInfo,
		Error: oError,
		Fatal: oFatal,
		Warn:  oWarn,
	}

	err := l.Write(F{"level": "fatal", "foo": "bar"})
	assert.NoError(t, err)
	assert.Nil(t, oInfo.last)
	assert.Nil(t, oError.last)
	assert.Equal(t, F{"level": "fatal", "foo": "bar"}, F(oFatal.last))
	assert.Nil(t, oWarn.last)

	reset()
	err = l.Write(F{"level": "error", "foo": "bar"})
	assert.NoError(t, err)
	assert.Nil(t, oInfo.last)
	assert.Equal(t, F{"level": "error", "foo": "bar"}, F(oError.last))
	assert.Nil(t, oFatal.last)
	assert.Nil(t, oWarn.last)

	reset()
	err = l.Write(F{"level": "info", "foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, F{"level": "info", "foo": "bar"}, F(oInfo.last))
	assert.Nil(t, oFatal.last)
	assert.Nil(t, oError.last)
	assert.Nil(t, oWarn.last)

	reset()
	err = l.Write(F{"level": "warn", "foo": "bar"})
	assert.EqualError(t, err, "some error")
	assert.Nil(t, oInfo.last)
	assert.Nil(t, oError.last)
	assert.Nil(t, oFatal.last)
	assert.Equal(t, F{"level": "warn", "foo": "bar"}, F(oWarn.last))

	reset()
	err = l.Write(F{"level": "debug", "foo": "bar"})
	assert.NoError(t, err)
	assert.Nil(t, oInfo.last)
	assert.Nil(t, oError.last)
	assert.Nil(t, oFatal.last)
	assert.Nil(t, oWarn.last)

	reset()
	err = l.Write(F{"foo": "bar"})
	assert.NoError(t, err)
	assert.Nil(t, oInfo.last)
	assert.Nil(t, oError.last)
	assert.Nil(t, oFatal.last)
	assert.Nil(t, oWarn.last)
}

func TestSyslogOutput(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	oldCritialLogger := critialLogger
	defer func() { critialLogger = oldCritialLogger }()
	critialLogger = func(v ...interface{}) {
		fmt.Fprint(buf, v...)
	}
	m := NewSyslogOutput("udp", "127.0.0.1:1234", "mytag")
	assert.IsType(t, LevelOutput{}, m)
	assert.Panics(t, func() {
		NewSyslogOutput("tcp", "an invalid host name", "mytag")
	})
	assert.Equal(t, "syslog dial error: dial tcp: missing port in address an invalid host name", buf.String())
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
	assert.Equal(t, "{\"foo\":\"bar\",\"level\":\"info\",\"message\":\"some message\"}", buf.String())
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
	o := &testOutput{}
	i := NewUIDOutput("id", o)
	err := i.Write(F{"message": "some message", "level": "info", "foo": "bar"})
	assert.NoError(t, err)
	assert.NotNil(t, o.last["id"])
	assert.Len(t, o.last["id"], 20)
}

func TestTrimOutput(t *testing.T) {
	o := &testOutput{}
	i := NewTrimOutput(10, o)
	err := i.Write(F{"short": "short", "long": "too long message", "number": 20})
	assert.NoError(t, err)
	assert.Equal(t, "short", o.last["short"])
	assert.Equal(t, "too long m", o.last["long"])
	assert.Equal(t, 20, o.last["number"])
}

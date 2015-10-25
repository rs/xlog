package xlog

import (
	"bytes"
	"errors"
	"testing"

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
	oWarn := &testOutput{err: errors.New("some error")}
	l := LevelOutput{
		Info:  oInfo,
		Error: oError,
		Warn:  oWarn,
	}

	err := l.Write(F{"level": "error", "foo": "bar"})
	assert.NoError(t, err)
	assert.Nil(t, oInfo.last)
	assert.Equal(t, F{"level": "error", "foo": "bar"}, F(oError.last))
	assert.Nil(t, oWarn.last)

	oInfo.last = nil
	oError.last = nil
	oWarn.last = nil
	err = l.Write(F{"level": "info", "foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, F{"level": "info", "foo": "bar"}, F(oInfo.last))
	assert.Nil(t, oError.last)
	assert.Nil(t, oWarn.last)

	oInfo.last = nil
	oError.last = nil
	oWarn.last = nil
	err = l.Write(F{"level": "warn", "foo": "bar"})
	assert.EqualError(t, err, "some error")
	assert.Nil(t, oInfo.last)
	assert.Nil(t, oError.last)
	assert.Equal(t, F{"level": "warn", "foo": "bar"}, F(oWarn.last))

	oInfo.last = nil
	oError.last = nil
	oWarn.last = nil
	err = l.Write(F{"level": "debug", "foo": "bar"})
	assert.NoError(t, err)
	assert.Nil(t, oInfo.last)
	assert.Nil(t, oError.last)
	assert.Nil(t, oWarn.last)

	oInfo.last = nil
	oError.last = nil
	oWarn.last = nil
	err = l.Write(F{"foo": "bar"})
	assert.NoError(t, err)
	assert.Nil(t, oInfo.last)
	assert.Nil(t, oError.last)
	assert.Nil(t, oWarn.last)
}

func TestSyslogOutput(t *testing.T) {
	m := NewSyslogOutput("", "", "mytag")
	assert.IsType(t, LevelOutput{}, m)
	assert.Panics(t, func() {
		NewSyslogOutput("tcp", "an invalid host name", "mytag")
	})
}

func TestNewConsoleOutput(t *testing.T) {
	c := NewConsoleOutput()
	assert.IsType(t, LevelOutput{}, c)
	l := c.(LevelOutput)
	assert.IsType(t, ConsoleOutput{}, l.Debug)
	assert.IsType(t, ConsoleOutput{}, l.Info)
	assert.IsType(t, ConsoleOutput{}, l.Warn)
	assert.IsType(t, ConsoleOutput{}, l.Error)
}

func TestConsoleOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	c := ConsoleOutput{w: buf}
	err := c.Write(F{"message": "some message", "level": "info", "foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "some message {\"foo\":\"bar\",\"level\":\"info\"}\n", buf.String())
}

func TestJSONOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	j := NewJSONOutput(buf)
	err := j.Write(F{"message": "some message", "level": "info", "foo": "bar"})
	assert.NoError(t, err)
	assert.Equal(t, "{\"foo\":\"bar\",\"level\":\"info\",\"message\":\"some message\"}\n", buf.String())
}

package xlog

import (
	"bytes"
	"log"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var fakeNow = time.Date(0, 0, 0, 0, 0, 0, 0, time.Local)

func init() {
	now = func() time.Time {
		return fakeNow
	}
}

func TestNew(t *testing.T) {
	c := Config{
		Level:  LevelError,
		Output: NewOutputChannel(&testOutput{}),
		Fields: F{"foo": "bar"},
	}
	L := New(c)
	l, ok := L.(*logger)
	if assert.True(t, ok) {
		assert.Equal(t, LevelError, l.level)
		assert.Equal(t, c.Output, l.output)
		assert.Equal(t, F{"foo": "bar"}, F(l.fields))
		// Ensure l.fields is a clone
		c.Fields["bar"] = "baz"
		assert.Equal(t, F{"foo": "bar"}, F(l.fields))
		l.close()
	}
}

func TestNewDefautOutput(t *testing.T) {
	L := New(Config{})
	l, ok := L.(*logger)
	if assert.True(t, ok) {
		assert.NotNil(t, l.output)
		l.close()
	}
}

func TestSend(t *testing.T) {
	o := testOutput{}
	oc := NewOutputChannel(&o)
	l := New(Config{Output: oc}).(*logger)
	l.send(LevelDebug, 1, "test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test", "foo": "bar"}, o.last)

	l.SetField("bar", "baz")
	l.send(LevelInfo, 1, "test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test", "foo": "bar", "bar": "baz"}, o.last)

	l = New(Config{Output: oc, Level: 1}).(*logger)
	o.last = nil
	l.send(0, 2, "test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Nil(t, o.last)
}

func TestSendDrop(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	critialLogOutput = buf
	defer func() { critialLogOutput = os.Stderr }()
	oc := NewOutputChannelBuffer(&testOutput{}, 1)
	l := New(Config{Output: oc}).(*logger)
	l.send(LevelDebug, 2, "test", F{"foo": "bar"})
	l.send(LevelDebug, 2, "test", F{"foo": "bar"})
	l.send(LevelDebug, 2, "test", F{"foo": "bar"})
	assert.Len(t, oc.input, 1)
	assert.Equal(t, "xlog: send error: buffer fullxlog: send error: buffer full", buf.String())
}

func TestWxtractFields(t *testing.T) {
	v := []interface{}{"a", 1, map[string]interface{}{"foo": "bar"}}
	f := extractFields(&v)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, f)
	assert.Equal(t, []interface{}{"a", 1}, v)

	v = []interface{}{map[string]interface{}{"foo": "bar"}, "a", 1}
	f = extractFields(&v)
	assert.Nil(t, f)
	assert.Equal(t, []interface{}{map[string]interface{}{"foo": "bar"}, "a", 1}, v)

	v = []interface{}{"a", 1, F{"foo": "bar"}}
	f = extractFields(&v)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, f)
	assert.Equal(t, []interface{}{"a", 1}, v)

	v = []interface{}{}
	f = extractFields(&v)
	assert.Nil(t, f)
	assert.Equal(t, []interface{}{}, v)
}

func TestDebug(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Debug("test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test", "foo": "bar"}, o.last)
}

func TestDebugf(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Debugf("test %d", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test 1", "foo": "bar"}, o.last)
}

func TestInfo(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Info("test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test", "foo": "bar"}, o.last)
}

func TestInfof(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Infof("test %d", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test 1", "foo": "bar"}, o.last)
}

func TestWarn(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Warn("test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "warn", "message": "test", "foo": "bar"}, o.last)
}

func TestWarnf(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Warnf("test %d", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "warn", "message": "test 1", "foo": "bar"}, o.last)
}

func TestError(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Error("test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test", "foo": "bar"}, o.last)
}

func TestErrorf(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Errorf("test %d%v", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test 1", "foo": "bar"}, o.last)
}

func TestFatal(t *testing.T) {
	e := exit1
	exited := 0
	exit1 = func() { exited++ }
	defer func() { exit1 = e }()
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Fatal("test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test", "foo": "bar"}, o.last)
	assert.Equal(t, 1, exited)
}

func TestFatalf(t *testing.T) {
	e := exit1
	exited := 0
	exit1 = func() { exited++ }
	defer func() { exit1 = e }()
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Fatalf("test %d%v", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test 1", "foo": "bar"}, o.last)
	assert.Equal(t, 1, exited)
}

func TestWrite(t *testing.T) {
	o := testOutput{}
	xl := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l := log.New(xl, "prefix ", 0)
	l.Printf("test")
	runtime.Gosched()
	assert.Contains(t, o.last["file"], "log_test.go:")
	delete(o.last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "prefix test"}, o.last)
}

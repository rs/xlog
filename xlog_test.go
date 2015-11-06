package xlog

import (
	"log"
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

func TestSend(t *testing.T) {
	o := testOutput{}
	oc := NewOutputChannel(&o)
	l := New(Config{Output: oc}).(*logger)
	l.send(LevelDebug, 2, "test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test", "foo": "bar", "file": "testing.go:456"}, o.last)

	l.SetField("bar", "baz")
	l.send(LevelInfo, 2, "test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test", "foo": "bar", "bar": "baz", "file": "testing.go:456"}, o.last)

	l = New(Config{Output: oc, Level: 1}).(*logger)
	o.last = nil
	l.send(0, 2, "test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Nil(t, o.last)
}

func TestSendDrop(t *testing.T) {
	oc := NewOutputChannel(&testOutput{})
	oc.input = make(chan map[string]interface{}, 1)
	l := New(Config{Output: oc}).(*logger)
	l.send(LevelDebug, 2, "test", F{"foo": "bar"})
	l.send(LevelDebug, 2, "test", F{"foo": "bar"})
	l.send(LevelDebug, 2, "test", F{"foo": "bar"})
	assert.Len(t, oc.input, 1)
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
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test", "foo": "bar", "file": "xlog_test.go:94"}, o.last)
}

func TestDebugf(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Debugf("test %d", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test 1", "foo": "bar", "file": "xlog_test.go:102"}, o.last)
}

func TestInfo(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Info("test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test", "foo": "bar", "file": "xlog_test.go:110"}, o.last)
}

func TestInfof(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Infof("test %d", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test 1", "foo": "bar", "file": "xlog_test.go:118"}, o.last)
}

func TestWarn(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Warn("test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "warn", "message": "test", "foo": "bar", "file": "xlog_test.go:126"}, o.last)
}

func TestWarnf(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Warnf("test %d", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "warn", "message": "test 1", "foo": "bar", "file": "xlog_test.go:134"}, o.last)
}

func TestError(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Error("test", F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test", "foo": "bar", "file": "xlog_test.go:142"}, o.last)
}

func TestErrorf(t *testing.T) {
	o := testOutput{}
	l := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l.Errorf("test %d", 1, F{"foo": "bar"})
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test 1", "foo": "bar", "file": "xlog_test.go:150"}, o.last)
}

func TestWrite(t *testing.T) {
	o := testOutput{}
	xl := New(Config{Output: NewOutputChannel(&o)}).(*logger)
	l := log.New(xl, "prefix ", 0)
	l.Printf("test")
	runtime.Gosched()
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "prefix test", "file": "xlog_test.go:159"}, o.last)
}

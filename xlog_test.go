package xlog

import (
	"log"
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

func TestSend(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.send(LevelDebug, "test", F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test", "foo": "bar"}, m)

	l.SetField("bar", "baz")
	l.send(LevelInfo, "test", F{"foo": "bar"})
	m = <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test", "foo": "bar", "bar": "baz"}, m)

	l.level = 1
	l.send(0, "test", F{"foo": "bar"})
	select {
	case <-c:
		t.Fatal("should skip message if level to low")
	default:
	}
}

func TestSendDrop(t *testing.T) {
	c := make(chan map[string]interface{}, 1)
	l := &logger{output: c}
	l.send(LevelDebug, "test", F{"foo": "bar"})
	l.send(LevelDebug, "test", F{"foo": "bar"})
	l.send(LevelDebug, "test", F{"foo": "bar"})
	assert.Len(t, c, 1)
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
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.Debug("test", F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test", "foo": "bar"}, m)
}

func TestDebugf(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.Debugf("test %d", 1, F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "debug", "message": "test 1", "foo": "bar"}, m)
}

func TestInfo(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.Info("test", F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test", "foo": "bar"}, m)
}

func TestInfof(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.Infof("test %d", 1, F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "test 1", "foo": "bar"}, m)
}

func TestWarn(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.Warn("test", F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "warn", "message": "test", "foo": "bar"}, m)
}

func TestWarnf(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.Warnf("test %d", 1, F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "warn", "message": "test 1", "foo": "bar"}, m)
}

func TestError(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.Error("test", F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test", "foo": "bar"}, m)
}

func TestErrorf(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	l := &logger{output: c}
	l.Errorf("test %d", 1, F{"foo": "bar"})
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test 1", "foo": "bar"}, m)
}

func TestWrite(t *testing.T) {
	c := make(chan map[string]interface{}, 10)
	xl := &logger{output: c}
	l := log.New(xl, "prefix ", 0)
	l.Printf("test")
	m := <-c
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "info", "message": "prefix test"}, m)
}

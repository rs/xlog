package xlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalLogger(t *testing.T) {
	o := newTestOutput()
	oldStd := std
	defer func() { std = oldStd }()
	SetLogger(New(Config{Output: o}))
	Debug("test")
	last := o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "debug", last["level"])
	o.reset()
	Debugf("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "debug", last["level"])
	o.reset()
	Info("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "info", last["level"])
	o.reset()
	Infof("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "info", last["level"])
	o.reset()
	Warn("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "warn", last["level"])
	o.reset()
	Warnf("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "warn", last["level"])
	o.reset()
	Error("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "error", last["level"])
	o.reset()
	Errorf("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "error", last["level"])
	o.reset()
	oldExit := exit1
	exit1 = func() {}
	defer func() { exit1 = oldExit }()
	Fatal("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "fatal", last["level"])
	o.reset()
	Fatalf("test")
	last = o.get()
	assert.Equal(t, "test", last["message"])
	assert.Equal(t, "fatal", last["level"])
	o.reset()
}

func TestStdError(t *testing.T) {
	o := newTestOutput()
	oldStd := std
	defer func() { std = oldStd }()
	SetLogger(New(Config{Output: o}))
	Error("test", F{"foo": "bar"})
	last := <-o.w
	assert.Contains(t, last["file"], "std_test.go:")
	delete(last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test", "foo": "bar"}, last)
}

func TestStdErrorf(t *testing.T) {
	o := newTestOutput()
	oldStd := std
	defer func() { std = oldStd }()
	SetLogger(New(Config{Output: o}))
	Errorf("test %d%v", 1, F{"foo": "bar"})
	last := <-o.w
	assert.Contains(t, last["file"], "std_test.go:")
	delete(last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "error", "message": "test 1", "foo": "bar"}, last)
}

func TestStdFatal(t *testing.T) {
	e := exit1
	exited := 0
	exit1 = func() { exited++ }
	defer func() { exit1 = e }()
	o := newTestOutput()
	oldStd := std
	defer func() { std = oldStd }()
	SetLogger(New(Config{Output: o}))
	Fatal("test", F{"foo": "bar"})
	last := <-o.w
	assert.Contains(t, last["file"], "std_test.go:")
	delete(last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "fatal", "message": "test", "foo": "bar"}, last)
	assert.Equal(t, 1, exited)
}

func TestStdFatalf(t *testing.T) {
	e := exit1
	exited := 0
	exit1 = func() { exited++ }
	defer func() { exit1 = e }()
	o := newTestOutput()
	oldStd := std
	defer func() { std = oldStd }()
	SetLogger(New(Config{Output: o}))
	Fatalf("test %d%v", 1, F{"foo": "bar"})
	last := <-o.w
	assert.Contains(t, last["file"], "std_test.go:")
	delete(last, "file")
	assert.Equal(t, map[string]interface{}{"time": fakeNow, "level": "fatal", "message": "test 1", "foo": "bar"}, last)
	assert.Equal(t, 1, exited)
}

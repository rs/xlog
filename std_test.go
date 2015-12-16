package xlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalLogger(t *testing.T) {
	o := testOutput{}
	oldStd := std
	defer func() { std = oldStd }()
	SetLogger(New(Config{Output: &o}))
	Debug("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "debug", o.last["level"])
	o.reset()
	Debugf("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "debug", o.last["level"])
	o.reset()
	Info("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "info", o.last["level"])
	o.reset()
	Infof("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "info", o.last["level"])
	o.reset()
	Warn("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "warn", o.last["level"])
	o.reset()
	Warnf("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "warn", o.last["level"])
	o.reset()
	Error("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "error", o.last["level"])
	o.reset()
	Errorf("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "error", o.last["level"])
	o.reset()
	oldExit := exit1
	exit1 = func() {}
	defer func() { exit1 = oldExit }()
	Fatal("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "fatal", o.last["level"])
	o.reset()
	Fatalf("test")
	assert.Equal(t, "test", o.last["message"])
	assert.Equal(t, "fatal", o.last["level"])
	o.reset()
}

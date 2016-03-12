package xlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelFromString(t *testing.T) {
	assert.Equal(t, LevelDebug, LevelFromString("debug"))
	assert.Equal(t, LevelInfo, LevelFromString("info"))
	assert.Equal(t, LevelWarn, LevelFromString("warn"))
	assert.Equal(t, LevelError, LevelFromString("error"))
	assert.Equal(t, LevelFatal, LevelFromString("fatal"))
	assert.Equal(t, LevelInfo, LevelFromString("info"))
}

func TestLevelUnmarshaler(t *testing.T) {
	l := Level(-1)
	err := l.UnmarshalText([]byte("debug"))
	assert.NoError(t, err)
	assert.Equal(t, LevelDebug, l)
	err = l.UnmarshalText([]byte("info"))
	assert.NoError(t, err)
	assert.Equal(t, LevelInfo, l)
	err = l.UnmarshalText([]byte("warn"))
	assert.NoError(t, err)
	assert.Equal(t, LevelWarn, l)
	err = l.UnmarshalText([]byte("error"))
	assert.NoError(t, err)
	assert.Equal(t, LevelError, l)
	err = l.UnmarshalText([]byte("fatal"))
	assert.NoError(t, err)
	assert.Equal(t, LevelFatal, l)
	assert.Error(t, l.UnmarshalText([]byte("invalid")))
}

func TestLevelString(t *testing.T) {
	assert.Equal(t, "debug", LevelDebug.String())
	assert.Equal(t, "info", LevelInfo.String())
	assert.Equal(t, "warn", LevelWarn.String())
	assert.Equal(t, "error", LevelError.String())
	assert.Equal(t, "fatal", LevelFatal.String())
	assert.Equal(t, "10", Level(10).String())
}

package xlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelFromString(t *testing.T) {
	l, err := LevelFromString("debug")
	assert.NoError(t, err)
	assert.Equal(t, LevelDebug, l)
	l, err = LevelFromString("info")
	assert.NoError(t, err)
	assert.Equal(t, LevelInfo, l)
	l, err = LevelFromString("warn")
	assert.NoError(t, err)
	assert.Equal(t, LevelWarn, l)
	l, err = LevelFromString("error")
	assert.NoError(t, err)
	assert.Equal(t, LevelError, l)
	l, err = LevelFromString("fatal")
	assert.NoError(t, err)
	assert.Equal(t, LevelFatal, l)
	_, err = LevelFromString("foo")
	assert.Error(t, err, "")
}

func TestLevelUnmarshalerText(t *testing.T) {
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

func TestLevelMarshalerText(t *testing.T) {
	b, err := LevelDebug.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, string(levelBytesDebug), string(b))
	b, err = LevelInfo.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, string(levelBytesInfo), string(b))
	b, err = LevelWarn.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, string(levelBytesWarn), string(b))
	b, err = LevelError.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, string(levelBytesError), string(b))
	b, err = LevelFatal.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, string(levelBytesFatal), string(b))
	b, err = Level(10).MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, "10", string(b))
}

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

func TestLevelString(t *testing.T) {
	assert.Equal(t, "debug", LevelDebug.String())
	assert.Equal(t, "info", LevelInfo.String())
	assert.Equal(t, "warn", LevelWarn.String())
	assert.Equal(t, "error", LevelError.String())
	assert.Equal(t, "fatal", LevelFatal.String())
	assert.Equal(t, "10", Level(10).String())
}

package xlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelString(t *testing.T) {
	assert.Equal(t, "debug", LevelDebug.String())
	assert.Equal(t, "info", LevelInfo.String())
	assert.Equal(t, "warn", LevelWarn.String())
	assert.Equal(t, "error", LevelError.String())
	assert.Equal(t, "10", Level(10).String())
}

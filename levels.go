package xlog

import (
	"bytes"
	"fmt"
	"strconv"
)

// Level defines log levels
type Level int

// Log levels
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// Log level strings
var (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
	levelFatal = "fatal"

	levelBytesDebug = []byte(levelDebug)
	levelBytesInfo  = []byte(levelInfo)
	levelBytesWarn  = []byte(levelWarn)
	levelBytesError = []byte(levelError)
	levelBytesFatal = []byte(levelFatal)
)

// LevelFromString returns the level based on its string representation
func LevelFromString(t string) (Level, error) {
	l := Level(0)
	err := (&l).UnmarshalText([]byte(t))
	return l, err
}

// UnmarshalText lets Level implements the TextUnmarshaler interface used by encoding packages
func (l *Level) UnmarshalText(text []byte) (err error) {
	if bytes.Equal(text, levelBytesDebug) {
		*l = LevelDebug
	} else if bytes.Equal(text, levelBytesInfo) {
		*l = LevelInfo
	} else if bytes.Equal(text, levelBytesWarn) {
		*l = LevelWarn
	} else if bytes.Equal(text, levelBytesError) {
		*l = LevelError
	} else if bytes.Equal(text, levelBytesFatal) {
		*l = LevelFatal
	} else {
		err = fmt.Errorf("Uknown level %v", string(text))
	}
	return
}

// String returns the string representation of the level.
func (l Level) String() string {
	var t string
	switch l {
	case LevelDebug:
		t = levelDebug
	case LevelInfo:
		t = levelInfo
	case LevelWarn:
		t = levelWarn
	case LevelError:
		t = levelError
	case LevelFatal:
		t = levelFatal
	default:
		t = strconv.FormatInt(int64(l), 10)
	}
	return t
}

// MarshalText lets Level implements the TextMarshaler interface used by encoding packages
func (l Level) MarshalText() ([]byte, error) {
	var t []byte
	switch l {
	case LevelDebug:
		t = levelBytesDebug
	case LevelInfo:
		t = levelBytesInfo
	case LevelWarn:
		t = levelBytesWarn
	case LevelError:
		t = levelBytesError
	case LevelFatal:
		t = levelBytesFatal
	default:
		t = []byte(strconv.FormatInt(int64(l), 10))
	}
	return t, nil
}

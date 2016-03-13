package xlog

import (
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
	levelBytesDebug = "debug"
	levelBytesInfo  = "info"
	levelBytesWarn  = "warn"
	levelBytesError = "error"
	levelBytesFatal = "fatal"
)

// LevelFromString returns the level based on its string representation
func LevelFromString(t string) (Level, error) {
	l := Level(0)
	err := (&l).UnmarshalText([]byte(t))
	return l, err
}

// UnmarshalText lets Level implements the TextUnmarshaler interface used by encoding packages
func (l *Level) UnmarshalText(text []byte) (err error) {
	switch string(text) {
	case levelBytesDebug:
		*l = LevelDebug
	case levelBytesInfo:
		*l = LevelInfo
	case levelBytesWarn:
		*l = LevelWarn
	case levelBytesError:
		*l = LevelError
	case levelBytesFatal:
		*l = LevelFatal
	default:
		err = fmt.Errorf("Uknown level %v", string(text))
	}
	return
}

// String returns the string representation of the level.
func (l Level) String() string {
	t, _ := l.MarshalText()
	return string(t)
}

// MarshalText lets Level implements the TextMarshaler interface used by encoding packages
func (l Level) MarshalText() ([]byte, error) {
	var t string
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
		t = strconv.FormatInt(int64(l), 10)
	}
	return []byte(t), nil
}

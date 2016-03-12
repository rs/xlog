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

// Log level bytes
var (
	levelBytesDebug = []byte("debug")
	levelBytesInfo  = []byte("info")
	levelBytesWarn  = []byte("warn")
	levelBytesError = []byte("error")
	levelBytesFatal = []byte("fatal")
)

// LevelFromString returns the level based on its string representation
func LevelFromString(l string) Level {
	switch l {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
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
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	default:
		return strconv.FormatInt(int64(l), 10)
	}
}

// MarshalText lets Level implements the TextMarshaler interface used by encoding packages
func (l Level) MarshalText() ([]byte, error) {
	switch l {
	case LevelDebug:
		return levelBytesDebug, nil
	case LevelInfo:
		return levelBytesInfo, nil
	case LevelWarn:
		return levelBytesWarn, nil
	case LevelError:
		return levelBytesError, nil
	case LevelFatal:
		return levelBytesFatal, nil
	default:
		return []byte(strconv.FormatInt(int64(l), 10)), nil
	}
}

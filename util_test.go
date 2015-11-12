package xlog

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestColorPrint(t *testing.T) {
	buf := &bytes.Buffer{}
	colorPrint(buf, "test", red)
	assert.Equal(t, "\x1b[31mtest\x1b[0m", buf.String())
	buf.Reset()
	colorPrint(buf, "test", green)
	assert.Equal(t, "\x1b[32mtest\x1b[0m", buf.String())
	buf.Reset()
	colorPrint(buf, "test", yellow)
	assert.Equal(t, "\x1b[33mtest\x1b[0m", buf.String())
	buf.Reset()
	colorPrint(buf, "test", blue)
	assert.Equal(t, "\x1b[34mtest\x1b[0m", buf.String())
	buf.Reset()
	colorPrint(buf, "test", gray)
	assert.Equal(t, "\x1b[37mtest\x1b[0m", buf.String())
}

func TestNeedsQuotedValueRune(t *testing.T) {
	assert.True(t, needsQuotedValueRune('='))
	assert.True(t, needsQuotedValueRune('"'))
	assert.True(t, needsQuotedValueRune(' '))
	assert.False(t, needsQuotedValueRune('a'))
	assert.False(t, needsQuotedValueRune('\''))
}

func TestWriteValue(t *testing.T) {
	buf := &bytes.Buffer{}
	write := func(v interface{}) string {
		buf.Reset()
		err := writeValue(buf, v)
		if err == nil {
			return buf.String()
		}
		return ""
	}
	assert.Equal(t, `foobar`, write(`foobar`))
	assert.Equal(t, `"foo=bar"`, write(`foo=bar`))
	assert.Equal(t, `"foo bar"`, write(`foo bar`))
	assert.Equal(t, `"foo\"bar"`, write(`foo"bar`))
	assert.Equal(t, `"foo\nbar"`, write("foo\nbar"))
	assert.Equal(t, `null`, write(nil))
	assert.Equal(t, `"2000-01-02 03:04:05 +0000 UTC"`, write(time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)))
	assert.Equal(t, `"error \"with quote\""`, write(errors.New(`error "with quote"`)))
}

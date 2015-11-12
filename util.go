package xlog

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type color int

const (
	red    color = 31
	green  color = 32
	yellow color = 33
	blue   color = 34
	gray   color = 37
)

func colorPrint(w io.Writer, s string, c color) {
	w.Write([]byte{0x1b, '[', byte('0' + c/10), byte('0' + c%10), 'm'})
	w.Write([]byte(s))
	w.Write([]byte("\x1b[0m"))
}

func needsQuotedValueRune(r rune) bool {
	return r <= ' ' || r == '=' || r == '"'
}

// writeValue writes a value on the writer in a logfmt compatible way
func writeValue(w io.Writer, v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		_, err = w.Write([]byte("null"))
	case string:
		if strings.IndexFunc(v, needsQuotedValueRune) != -1 {
			var b []byte
			b, err = json.Marshal(v)
			if err == nil {
				w.Write(b)
			}
		} else {
			_, err = w.Write([]byte(v))
		}
	case error:
		s := v.Error()
		err = writeValue(w, s)
	default:
		s := fmt.Sprint(v)
		err = writeValue(w, s)
	}
	return
}

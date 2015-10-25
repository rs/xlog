package xlog_test

import (
	"errors"

	"github.com/rs/xlog"
	"golang.org/x/net/context"
)

func Example_log() {
	ctx := context.TODO() // got from xhandler
	l := xlog.FromContext(ctx)

	// Log a simple message
	l.Debug("message")

	if err := errors.New("some error"); err != nil {
		l.Errorf("Some error happened: %v", err)
	}

	// With optional fields
	l.Debugf("foo %s", "bar", xlog.F{
		"field": "value",
	})
}

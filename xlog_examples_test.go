// +build go1.7

package xlog_test

import (
	"context"
	"errors"
	"log"

	"github.com/rs/xlog"
)

func Example_log() {
	ctx := context.TODO()
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

func Example_stdlog() {
	// Define logger conf
	conf := xlog.Config{
		Output: xlog.NewConsoleOutput(),
	}

	// Remove timestamp and other decorations of the std logger
	log.SetFlags(0)

	// Plug a xlog instance to Go's std logger
	log.SetOutput(xlog.New(conf))
}

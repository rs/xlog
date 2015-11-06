package xlog_test

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/rs/xhandler"
	"github.com/rs/xlog"
	"golang.org/x/net/context"
)

func Example_handler() {
	c := xhandler.Chain{}

	host, _ := os.Hostname()
	conf := xlog.Config{
		// Set some global env fields
		Fields: xlog.F{
			"role": "my-service",
			"host": host,
		},
	}

	// Install the logger handler with default output on the console
	c.UseC(xlog.NewHandler(conf))

	// Plug the xlog handler's input to Go's default logger
	log.SetOutput(xlog.New(conf))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to those handler, all our logs will come with some pre-populated fields.
	c.UseC(xlog.RemoteAddrHandler("ip"))
	c.UseC(xlog.UserAgentHandler("user-agent"))
	c.UseC(xlog.RefererHandler("referer"))

	// Here is your final handler
	h := c.Handler(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Get the logger from the context. You can safely assume it will be always there,
		// if the handler is removed, xlog.FromContext will return a NopLogger
		l := xlog.FromContext(ctx)

		// Then log some errors
		if err := errors.New("some error from elsewhere"); err != nil {
			l.Errorf("Here is an error: %v", err)
		}

		// Or some info with fields
		l.Info("Something happend", xlog.F{
			"user":   "current user id",
			"status": "ok",
		})
	}))
	http.Handle("/", h)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func Example_stdlog() {
	// Define logger conf
	conf := xlog.Config{
		Output: xlog.NewOutputChannel(xlog.NewConsoleOutput()),
	}

	// Plug the xlog handler's input to Go's default logger
	log.SetOutput(xlog.New(conf))
}

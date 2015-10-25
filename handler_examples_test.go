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
	var xh xhandler.HandlerC

	// Here is your handler
	xh = xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
	})

	// Install some provided extra handler to set some request's context fields.
	// Thanks to those handler, all our logs will come with some pre-populated fields.
	xh = xlog.NewRemoteAddrHandler("ip", xh)
	xh = xlog.NewUserAgentHandler("user-agent", xh)
	xh = xlog.NewRefererHandler("referer", xh)

	// Install the logger handler with default output on the console
	lh := xlog.NewHandler(xlog.LevelDebug, xh)

	// Set some global env fields
	host, _ := os.Hostname()
	lh.SetFields(xlog.F{
		"role": "my-service",
		"host": host,
	})

	// Plug the xlog handler's input to Go's default logger
	log.SetOutput(lh.NewLogger())

	// Root context
	var h http.Handler
	ctx := context.Background()
	h = xhandler.New(ctx, lh)
	http.Handle("/", h)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func Example_stdlog() {
	// Plug handler's xlog to Go's default log.Logger output
	xh := xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Your handler
	})

	// Install the logger handler
	lh := xlog.NewHandler(xlog.LevelDebug, xh)

	// Plug the xlog handler's input to Go's default logger
	log.SetOutput(lh.NewLogger())

	// Root context
	var h http.Handler
	ctx := context.Background()
	h = xhandler.New(ctx, lh)
	http.Handle("/", h)
}

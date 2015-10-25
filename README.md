# HTTP Handler Logger

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/xlog) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/xlog/master/LICENSE) [![Build Status](https://travis-ci.org/rs/xlog.svg?branch=master)](https://travis-ci.org/rs/xlog)

xlog is a logger coupled with HTTP net/context aware middleware.

## Features

- Per request log context
- Per request and/or per message key/value fields
- Log levels (Debug, Info, Warn, Error)
- Custom output
- Automatically gathers request context like User-Agent, IP etc.
- Drops message rather than blocking execution

It works best in combination with [github.com/rs/xhandler](https://github.com/rs/xhandler).

## Install

    go get github.com/rs/xlog

## Usage

```go
var xh xhandler.HandlerC

// Here is your handler
xh = xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    // Get the logger from the context. You can safely assume it will be always there,
    // if the handler is removed, xlog.FromContext will return a nopLogger
    log := xlog.FromContext(ctx)

    // Then log some errors
    log.Errorf("Here is an error: %v", err)

    // Or some info with fields
    log.Info("Something happend", xlog.F{
        "user": user.ID,
        "status": status,
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

// Root context
var h http.Handler
ctx := context.Background()
h = xhandler.New(ctx, lh)
http.Handle("/", h)

if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal(err)
}
```

### Configure Output

By default, output is setup to output debug and info message on `STDOUT` and warning and errors to `STDERR`. You can easily change this setup.

XLog output can be customized using composable output handlers. Thanks to the [LevelOutput](https://godoc.org/github.com/rs/xlog#LevelOutput), [MultiOutput](https://godoc.org/github.com/rs/xlog#MultiOutput) and [FilterOutput](https://godoc.org/github.com/rs/xlog#FilterOutput), it is easy to route messages precisely.

```go
logH = xlog.NewHandler(xlog.LevelDebug, nextHandler)
logH.SetOutput(xlog.MultiOutput{
    // Send all logs with field type=mymodule to a remote syslog
    xlog.FilterOutput{
        Cond: func(fields map[string]interface{}) bool {
            return fields["type"] == "mymodule"
        },
        Output: xlog.NewSyslogOutput("tcp", "1.2.3.4:1234", "mymodule"),
    },
    // Setup different output per log level
    xlog.LevelOutput{
        // Send errors to the console
        Error: xlog.NewConsoleOutput(),
        // Send syslog output for error level
        Info: xlog.NewSyslogOutput("", "", ""),
    },
})
```

## Licenses

All source code is licensed under the [MIT License](https://raw.github.com/rs/xlog/master/LICENSE).

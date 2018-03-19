package xlog_test

import "github.com/rs/xlog"

func ExampleSetLogger() {
	xlog.SetLogger(xlog.New(xlog.Config{
		Level:  xlog.LevelInfo,
		Output: xlog.NewConsoleOutput(),
		Fields: xlog.F{
			"role": "my-service",
		},
	}))
}

package xlog_test

import "github.com/Ak-Army/xlog"

func ExampleSetLogger() {
	xlog.SetLogger(xlog.New(xlog.Config{
		Level:  xlog.LevelInfo,
		Output: xlog.NewConsoleOutput(),
		Fields: xlog.F{
			"role": "my-service",
		},
	}))
}

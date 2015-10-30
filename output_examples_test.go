package xlog_test

import "github.com/rs/xlog"

func Example_combinedOutputs() {
	lh := xlog.NewHandler(xlog.LevelDebug)

	lh.SetOutput(xlog.MultiOutput{
		// Output interesting messages to console
		xlog.FilterOutput{
			Cond: func(fields map[string]interface{}) bool {
				val, found := fields["type"]
				return found && val == "interesting"
			},
			Output: xlog.NewConsoleOutput(),
		},
		// Also setup by-level loggers
		xlog.LevelOutput{
			// Send debug messages to console if they match type
			Debug: xlog.FilterOutput{
				Cond: func(fields map[string]interface{}) bool {
					val, found := fields["type"]
					return found && val == "interesting"
				},
				Output: xlog.NewConsoleOutput(),
			},
		},
		// Also send everything over syslog
		xlog.NewSyslogOutput("", "", ""),
	})
}

func ExampleMultiOutput() {
	lh := xlog.NewHandler(xlog.LevelDebug)

	lh.SetOutput(xlog.MultiOutput{
		// Output everything to console
		xlog.NewConsoleOutput(),
		// and also to local syslog
		xlog.NewSyslogOutput("", "", ""),
	})
}

func ExampleFilterOutput() {
	lh := xlog.NewHandler(xlog.LevelDebug)

	lh.SetOutput(xlog.FilterOutput{
		// Match messages containing a field type = interesting
		Cond: func(fields map[string]interface{}) bool {
			val, found := fields["type"]
			return found && val == "interesting"
		},
		// Output matching messages to the console
		Output: xlog.NewConsoleOutput(),
	})
}

func ExampleLevelOutput() {
	lh := xlog.NewHandler(xlog.LevelDebug)

	lh.SetOutput(xlog.LevelOutput{
		// Send debug message to console
		Debug: xlog.NewConsoleOutput(),
		// and error messages to syslog
		Error: xlog.NewSyslogOutput("", "", ""),
		// other levels are discarded
	})
}

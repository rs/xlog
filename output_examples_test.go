package xlog_test

import (
	"log/syslog"

	"github.com/rs/xlog"
)

func Example_combinedOutputs() {
	conf := xlog.Config{
		Output: xlog.NewOutputChannel(xlog.MultiOutput{
			// Output interesting messages to console
			0: xlog.FilterOutput{
				Cond: func(fields map[string]interface{}) bool {
					val, found := fields["type"]
					return found && val == "interesting"
				},
				Output: xlog.NewConsoleOutput(),
			},
			// Also setup by-level loggers
			1: xlog.LevelOutput{
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
			2: xlog.NewSyslogOutput("", "", ""),
		}),
	}

	lh := xlog.NewHandler(conf)
	_ = lh
}

func ExampleMultiOutput() {
	conf := xlog.Config{
		Output: xlog.NewOutputChannel(xlog.MultiOutput{
			// Output everything to console
			0: xlog.NewConsoleOutput(),
			// and also to local syslog
			1: xlog.NewSyslogOutput("", "", ""),
		}),
	}
	lh := xlog.NewHandler(conf)
	_ = lh
}

func ExampleFilterOutput() {
	conf := xlog.Config{
		Output: xlog.NewOutputChannel(xlog.FilterOutput{
			// Match messages containing a field type = interesting
			Cond: func(fields map[string]interface{}) bool {
				val, found := fields["type"]
				return found && val == "interesting"
			},
			// Output matching messages to the console
			Output: xlog.NewConsoleOutput(),
		}),
	}

	lh := xlog.NewHandler(conf)
	_ = lh
}

func ExampleLevelOutput() {
	conf := xlog.Config{
		Output: xlog.NewOutputChannel(xlog.LevelOutput{
			// Send debug message to console
			Debug: xlog.NewConsoleOutput(),
			// and error messages to syslog
			Error: xlog.NewSyslogOutput("", "", ""),
			// other levels are discarded
		}),
	}

	lh := xlog.NewHandler(conf)
	_ = lh
}

func ExampleNewSyslogWriter() {
	conf := xlog.Config{
		Output: xlog.NewOutputChannel(xlog.LevelOutput{
			Debug: xlog.NewLogstashOutput(xlog.NewSyslogWriter("", "", syslog.LOG_LOCAL0|syslog.LOG_DEBUG, "")),
			Info:  xlog.NewLogstashOutput(xlog.NewSyslogWriter("", "", syslog.LOG_LOCAL0|syslog.LOG_INFO, "")),
			Warn:  xlog.NewLogstashOutput(xlog.NewSyslogWriter("", "", syslog.LOG_LOCAL0|syslog.LOG_WARNING, "")),
			Error: xlog.NewLogstashOutput(xlog.NewSyslogWriter("", "", syslog.LOG_LOCAL0|syslog.LOG_ERR, "")),
		}),
	}

	lh := xlog.NewHandler(conf)
	_ = lh
}

// +build !windows

package xlog

import (
	"io"
	"log/syslog"
)

// NewSyslogOutput returns JSONOutputs in a LevelOutput with writers set to syslog
// with the proper priority added to a LOG_USER facility.
// If network and address are empty, Dial will connect to the local syslog server.
func NewSyslogOutput(network, address, tag string) Output {
	return NewSyslogOutputFacility(network, address, tag, syslog.LOG_USER)
}

// NewSyslogOutputFacility returns JSONOutputs in a LevelOutput with writers set to syslog
// with the proper priority added to the passed facility.
// If network and address are empty, Dial will connect to the local syslog server.
func NewSyslogOutputFacility(network, address, tag string, facility syslog.Priority) Output {
	o := LevelOutput{
		Debug: NewJSONOutput(NewSyslogWriter(network, address, facility|syslog.LOG_DEBUG, tag)),
		Info:  NewJSONOutput(NewSyslogWriter(network, address, facility|syslog.LOG_INFO, tag)),
		Warn:  NewJSONOutput(NewSyslogWriter(network, address, facility|syslog.LOG_WARNING, tag)),
		Error: NewJSONOutput(NewSyslogWriter(network, address, facility|syslog.LOG_ERR, tag)),
	}
	return o
}

// NewSyslogWriter returns a writer ready to be used with output modules.
// If network and address are empty, Dial will connect to the local syslog server.
func NewSyslogWriter(network, address string, prio syslog.Priority, tag string) io.Writer {
	s, err := syslog.Dial(network, address, prio, tag)
	if err != nil {
		m := "syslog dial error: " + err.Error()
		critialLogger.Print(m)
		panic(m)
	}
	return s
}

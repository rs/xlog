package xlog

var std = New(Config{
	Output: NewConsoleOutput(),
})

// SetLogger changes the global logger instance
func SetLogger(logger Logger) {
	std = logger
}

// Debug calls the Debug() method on the default logger
func Debug(v ...interface{}) {
	std.Debug(v...)
}

// Debugf calls the Debugf() method on the default logger
func Debugf(format string, v ...interface{}) {
	std.Debugf(format, v...)
}

// Info calls the Info() method on the default logger
func Info(v ...interface{}) {
	std.Info(v...)
}

// Infof calls the Infof() method on the default logger
func Infof(format string, v ...interface{}) {
	std.Infof(format, v...)
}

// Warn calls the Warn() method on the default logger
func Warn(v ...interface{}) {
	std.Warn(v...)
}

// Warnf calls the Warnf() method on the default logger
func Warnf(format string, v ...interface{}) {
	std.Warnf(format, v...)
}

// Error calls the Error() method on the default logger
func Error(v ...interface{}) {
	std.Error(v...)
}

// Errorf calls the Errorf() method on the default logger
//
// Go vet users: you may append %v at the end of you format when using xlog.F{} as a last
// argument to workaround go vet false alarm.
func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

// Fatal calls the Fatal() method on the default logger
func Fatal(v ...interface{}) {
	std.Fatal(v...)
}

// Fatalf calls the Fatalf() method on the default logger
//
// Go vet users: you may append %v at the end of you format when using xlog.F{} as a last
// argument to workaround go vet false alarm.
func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}

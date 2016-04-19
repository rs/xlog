package xlog

import "fmt"

var std = New(Config{
	Output: NewConsoleOutput(),
})

// SetLogger changes the global logger instance
func SetLogger(logger Logger) {
	std = logger
}

// Debug calls the Debug() method on the default logger
func Debug(v ...interface{}) {
	f := extractFields(&v)
	std.OutputF(LevelDebug, 2, fmt.Sprint(v...), f)
}

// Debugf calls the Debugf() method on the default logger
func Debugf(format string, v ...interface{}) {
	f := extractFields(&v)
	std.OutputF(LevelDebug, 2, fmt.Sprintf(format, v...), f)
}

// Info calls the Info() method on the default logger
func Info(v ...interface{}) {
	f := extractFields(&v)
	std.OutputF(LevelInfo, 2, fmt.Sprint(v...), f)
}

// Infof calls the Infof() method on the default logger
func Infof(format string, v ...interface{}) {
	f := extractFields(&v)
	std.OutputF(LevelInfo, 2, fmt.Sprintf(format, v...), f)
}

// Warn calls the Warn() method on the default logger
func Warn(v ...interface{}) {
	f := extractFields(&v)
	std.OutputF(LevelWarn, 2, fmt.Sprint(v...), f)
}

// Warnf calls the Warnf() method on the default logger
func Warnf(format string, v ...interface{}) {
	f := extractFields(&v)
	std.OutputF(LevelWarn, 2, fmt.Sprintf(format, v...), f)
}

// Error calls the Error() method on the default logger
func Error(v ...interface{}) {
	f := extractFields(&v)
	std.OutputF(LevelError, 2, fmt.Sprint(v...), f)
}

// Errorf calls the Errorf() method on the default logger
//
// Go vet users: you may append %v at the end of you format when using xlog.F{} as a last
// argument to workaround go vet false alarm.
func Errorf(format string, v ...interface{}) {
	f := extractFields(&v)
	if f != nil {
		// Let user add a %v at the end of the message when fields are passed to satisfy go vet
		l := len(format)
		if l > 2 && format[l-2] == '%' && format[l-1] == 'v' {
			format = format[0 : l-2]
		}
	}
	std.OutputF(LevelError, 2, fmt.Sprintf(format, v...), f)
}

// Fatal calls the Fatal() method on the default logger
func Fatal(v ...interface{}) {
	f := extractFields(&v)
	std.OutputF(LevelFatal, 2, fmt.Sprint(v...), f)
	if l, ok := std.(*logger); ok {
		if o, ok := l.output.(*OutputChannel); ok {
			o.Close()
		}
	}
	exit1()
}

// Fatalf calls the Fatalf() method on the default logger
//
// Go vet users: you may append %v at the end of you format when using xlog.F{} as a last
// argument to workaround go vet false alarm.
func Fatalf(format string, v ...interface{}) {
	f := extractFields(&v)
	if f != nil {
		// Let user add a %v at the end of the message when fields are passed to satisfy go vet
		l := len(format)
		if l > 2 && format[l-2] == '%' && format[l-1] == 'v' {
			format = format[0 : l-2]
		}
	}
	std.OutputF(LevelFatal, 2, fmt.Sprintf(format, v...), f)
	if l, ok := std.(*logger); ok {
		if o, ok := l.output.(*OutputChannel); ok {
			o.Close()
		}
	}
	exit1()
}

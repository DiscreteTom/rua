package rua

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

type basicSimpleLogger struct {
	logger SmallestLogger
}

func NewBasicSimpleLogger(l SmallestLogger) *basicSimpleLogger {
	return &basicSimpleLogger{
		logger: l,
	}
}

func (l *basicSimpleLogger) log(level LogLevel, v ...interface{}) {
	// build output string
	var s string
	switch level {
	case TRACE:
		s = "[TRACE] "
	case DEBUG:
		s = "[DEBUG] "
	case INFO:
		s = "[INFO] "
	case WARN:
		s = "[WARN] "
	case ERROR:
		s = "[ERROR] "
	case FATAL:
		s = "[FATAL] "
	case PANIC:
		s = "[PANIC] "
	}
	s += fmt.Sprint(v...)
	s += "\n"

	l.logger.Print(s)

	if level == FATAL {
		os.Exit(1)
	} else if level == PANIC {
		panic(s)
	}
}

func (l *basicSimpleLogger) Trace(v ...interface{}) {
	l.log(TRACE, v...)
}
func (l *basicSimpleLogger) Debug(v ...interface{}) {
	l.log(DEBUG, v...)
}
func (l *basicSimpleLogger) Info(v ...interface{}) {
	l.log(INFO, v...)
}
func (l *basicSimpleLogger) Warn(v ...interface{}) {
	l.log(WARN, v...)
}
func (l *basicSimpleLogger) Error(v ...interface{}) {
	l.log(ERROR, v...)
}
func (l *basicSimpleLogger) Fatal(v ...interface{}) {
	l.log(FATAL, v...)
}
func (l *basicSimpleLogger) Panic(v ...interface{}) {
	l.log(PANIC, v...)
}

type basicLogger struct {
	logger SimpleLogger
	level  LogLevel
}

func NewBasicLogger(l SimpleLogger) *basicLogger {
	return &basicLogger{
		logger: l,
		level:  INFO,
	}
}

func (l *basicLogger) WithLevel(lvl LogLevel) *basicLogger {
	l.level = lvl
	return l
}

func (l *basicLogger) Trace(v ...interface{}) {
	if l.level <= TRACE {
		l.logger.Trace(v...)
	}
}
func (l *basicLogger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Debug(v...)
	}
}
func (l *basicLogger) Info(v ...interface{}) {
	if l.level <= INFO {
		l.logger.Info(v...)
	}
}
func (l *basicLogger) Warn(v ...interface{}) {
	if l.level <= WARN {
		l.logger.Warn(v...)
	}
}
func (l *basicLogger) Error(v ...interface{}) {
	if l.level <= ERROR {
		l.logger.Error(v...)
	}
}
func (l *basicLogger) Fatal(v ...interface{}) {
	if l.level <= FATAL {
		l.logger.Fatal(v...)
	}
}
func (l *basicLogger) Panic(v ...interface{}) {
	if l.level <= PANIC {
		l.logger.Panic(v...)
	}
}
func (l *basicLogger) Tracef(format string, v ...interface{}) {
	if l.level <= TRACE {
		l.logger.Trace(fmt.Sprintf(format, v...))
	}
}
func (l *basicLogger) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}
func (l *basicLogger) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		l.logger.Info(fmt.Sprintf(format, v...))
	}
}
func (l *basicLogger) Warnf(format string, v ...interface{}) {
	if l.level <= WARN {
		l.logger.Warn(fmt.Sprintf(format, v...))
	}
}
func (l *basicLogger) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.logger.Error(fmt.Sprintf(format, v...))
	}
}
func (l *basicLogger) Fatalf(format string, v ...interface{}) {
	if l.level <= FATAL {
		l.logger.Fatal(fmt.Sprintf(format, v...))
	}
}
func (l *basicLogger) Panicf(format string, v ...interface{}) {
	if l.level <= PANIC {
		l.logger.Panic(fmt.Sprintf(format, v...))
	}
}

var defaultLogger Logger

// Get the existing or create a new default logger.
func DefaultLogger() Logger {
	if defaultLogger == nil {
		defaultLogger = NewDefaultLogger()
	}
	return defaultLogger
}

func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

func NewDefaultLogger() *basicLogger {
	return NewBasicLogger(NewBasicSimpleLogger(log.Default()))
}

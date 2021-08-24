package rua

import "log"

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

type defaultLogger struct {
	level LogLevel
}

var logger *defaultLogger

func GetDefaultLogger() *defaultLogger {
	if logger == nil {
		logger = newDefaultLogger()
	}
	return logger
}

func newDefaultLogger() *defaultLogger {
	return &defaultLogger{
		level: INFO,
	}
}

func (l *defaultLogger) WithLevel(lvl LogLevel) *defaultLogger {
	l.level = lvl
	return l
}

func (l *defaultLogger) Trace(v ...interface{}) {
	if l.level <= TRACE {
		log.Print("[TRACE] ")
		log.Println(v...)
	}
}
func (l *defaultLogger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		log.Print("[DEBUG] ")
		log.Println(v...)
	}
}
func (l *defaultLogger) Info(v ...interface{}) {
	if l.level <= INFO {
		log.Print("[INFO] ")
		log.Println(v...)
	}
}
func (l *defaultLogger) Warn(v ...interface{}) {
	if l.level <= WARN {
		log.Print("[WARN] ")
		log.Println(v...)
	}
}
func (l *defaultLogger) Error(v ...interface{}) {
	if l.level <= ERROR {
		log.Print("[ERROR] ")
		log.Println(v...)
	}
}
func (l *defaultLogger) Fatal(v ...interface{}) {
	if l.level <= FATAL {
		log.Fatal("[FATAL] ")
		log.Fatalln(v...)
	}
}
func (l *defaultLogger) Panic(v ...interface{}) {
	if l.level <= PANIC {
		log.Panic("[PANIC] ")
		log.Panicln(v...)
	}
}

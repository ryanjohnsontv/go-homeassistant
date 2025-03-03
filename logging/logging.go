package logging

import "log"

type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
}

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type DefaultLogger struct {
	level LogLevel
}

func (d *DefaultLogger) SetLevel(level LogLevel) {
	d.level = level
}

func (d *DefaultLogger) Debug(msg string, fields ...any) {
	if d.level <= DebugLevel {
		log.Printf("DEBUG: "+msg, fields...)
	}
}

func (d *DefaultLogger) Info(msg string, fields ...any) {
	if d.level <= InfoLevel {
		log.Printf("INFO: "+msg, fields...)
	}
}

func (d *DefaultLogger) Warn(msg string, fields ...any) {
	if d.level <= WarnLevel {
		log.Printf("WARN: "+msg, fields...)
	}
}

func (d *DefaultLogger) Error(msg string, fields ...any) {
	if d.level <= ErrorLevel {
		log.Printf("ERROR: "+msg, fields...)
	}
}

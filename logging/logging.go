package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(err error, msg string, fields ...any)
}

type logLevel int

const (
	debugLevel logLevel = iota
	infoLevel
	warnLevel
	errorLevel

	defaultLogLevel = debugLevel
)

var (
	logLevelMappings = map[string]logLevel{
		"DEBUG": debugLevel,
		"INFO":  infoLevel,
		"WARN":  warnLevel,
		"ERROR": errorLevel,
	}

	logLevelStrings = []string{
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
	}
)

type DefaultLogger struct {
	level logLevel
}

func NewLogger() *DefaultLogger {
	levelStr := os.Getenv("LOG_LEVEL")
	if level, ok := logLevelMappings[strings.ToUpper(levelStr)]; ok {
		return &DefaultLogger{
			level: level,
		}
	}

	return &DefaultLogger{
		level: defaultLogLevel,
	}
}

func (d *DefaultLogger) Debug(msg string, fields ...any) {
	if d.level <= debugLevel {
		d.log(debugLevel, nil, msg, fields...)
	}
}

func (d *DefaultLogger) Info(msg string, fields ...any) {
	if d.level <= infoLevel {
		d.log(infoLevel, nil, msg, fields...)
	}
}

func (d *DefaultLogger) Warn(msg string, fields ...any) {
	if d.level <= warnLevel {
		d.log(warnLevel, nil, msg, fields...)
	}
}

func (d *DefaultLogger) Error(err error, msg string, fields ...any) {
	if d.level <= errorLevel {
		d.log(errorLevel, err, msg, fields...)
	}
}

func (d *DefaultLogger) log(level logLevel, err error, msg string, fields ...any) {
	// Create log entry
	logEntry := map[string]any{
		"level":   logLevelStrings[level],
		"message": fmt.Sprintf(msg, fields...),
	}
	if err != nil {
		logEntry["error"] = err.Error()
	}

	logEntryJSON, jsonErr := json.Marshal(logEntry)
	if jsonErr != nil {
		log.Printf("Error marshalling log entry to JSON: %v", jsonErr)
		return
	}

	log.Println(string(logEntryJSON))
}

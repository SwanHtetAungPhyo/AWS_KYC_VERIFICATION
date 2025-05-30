package logger

import (
	"github.com/sirupsen/logrus"
)

// Logger interface defines logging methods
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}
type LogrusLogger struct {
	*logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &LogrusLogger{Logger: logger}
}

// WithField creates a new logger with a field
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusEntry{l.Logger.WithField(key, value)}
}

// WithFields creates a new logger with multiple fields
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	logrusFields := make(logrus.Fields)
	for k, v := range fields {
		logrusFields[k] = v
	}
	return &LogrusEntry{l.Logger.WithFields(logrusFields)}
}

func (l *LogrusLogger) WithError(err error) Logger {
	return &LogrusEntry{l.Logger.WithError(err)}
}

type LogrusEntry struct {
	*logrus.Entry
}

func (l *LogrusEntry) WithField(key string, value interface{}) Logger {
	return &LogrusEntry{l.Entry.WithField(key, value)}
}

func (l *LogrusEntry) WithFields(fields map[string]interface{}) Logger {
	logrusFields := make(logrus.Fields)
	for k, v := range fields {
		logrusFields[k] = v
	}
	return &LogrusEntry{l.Entry.WithFields(logrusFields)}
}

func (l *LogrusEntry) WithError(err error) Logger {
	return &LogrusEntry{l.Entry.WithError(err)}
}

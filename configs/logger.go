package configs

import (
	"log"
	"os"
)

// Logger defines simple logging utilities.
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

var Log *Logger

func init() {
	Log = &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs messages with info prefix.
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error logs messages with error prefix.
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

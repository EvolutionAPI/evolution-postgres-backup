package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelDebug LogLevel = "DEBUG"
)

type Logger struct {
	logDir     string
	consoleLog *log.Logger
	fileLog    *log.Logger
	logFile    *os.File
}

func NewLogger(logDir string) (*Logger, error) {
	// Create logs directory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	logFileName := fmt.Sprintf("backup_%s.log", time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		logDir:     logDir,
		consoleLog: log.New(os.Stdout, "", log.LstdFlags),
		fileLog:    log.New(logFile, "", log.LstdFlags),
		logFile:    logFile,
	}, nil
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) log(level LogLevel, component, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf(message, args...)

	logEntry := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, component, formattedMessage)

	// Log to console
	l.consoleLog.Println(logEntry)

	// Log to file
	if l.fileLog != nil {
		l.fileLog.Println(logEntry)
	}
}

func (l *Logger) Info(component, message string, args ...interface{}) {
	l.log(LogLevelInfo, component, message, args...)
}

func (l *Logger) Warn(component, message string, args ...interface{}) {
	l.log(LogLevelWarn, component, message, args...)
}

func (l *Logger) Error(component, message string, args ...interface{}) {
	l.log(LogLevelError, component, message, args...)
}

func (l *Logger) Debug(component, message string, args ...interface{}) {
	l.log(LogLevelDebug, component, message, args...)
}

// Job-specific logging
func (l *Logger) LogJobStart(jobID, jobType, details string) {
	l.Info("JOB", "Started %s job [%s]: %s", jobType, jobID, details)
}

func (l *Logger) LogJobProgress(jobID, message string, args ...interface{}) {
	l.Info("JOB", "[%s] %s", jobID, fmt.Sprintf(message, args...))
}

func (l *Logger) LogJobSuccess(jobID, message string, args ...interface{}) {
	l.Info("JOB", "[%s] ✅ SUCCESS: %s", jobID, fmt.Sprintf(message, args...))
}

func (l *Logger) LogJobError(jobID, message string, args ...interface{}) {
	l.Error("JOB", "[%s] ❌ ERROR: %s", jobID, fmt.Sprintf(message, args...))
}

func (l *Logger) LogJobWarn(jobID, message string, args ...interface{}) {
	l.Warn("JOB", "[%s] ⚠️  WARNING: %s", jobID, fmt.Sprintf(message, args...))
}

// Global logger instance
var globalLogger *Logger

func InitLogger(logDir string) error {
	var err error
	globalLogger, err = NewLogger(logDir)
	return err
}

func GetLogger() *Logger {
	return globalLogger
}

func CloseLogger() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}

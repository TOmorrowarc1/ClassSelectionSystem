package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Logger struct {
	level_       LogLevel
	lock_        sync.Mutex
	log_channel_ chan string
	waitgroup_   sync.WaitGroup
	log_file     *os.File
}

var (
	logger_instance *Logger
	once            sync.Once
)

func GetLogger() *Logger {
	once.Do(func() {
		logger_instance = &Logger{
			level_:       DEBUG,
			log_channel_: make(chan string, 100),
		}
		logger_instance.waitgroup_.Add(1)
		go logger_instance.output()
	})
	return logger_instance
}

func (l *Logger) SetLogFile(filePath string) error {
	l.lock_.Lock()
	defer l.lock_.Unlock()
	if l.log_file != nil {
		l.log_file.Close()
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	l.log_file = file
	return nil
}

func (l *Logger) SetLogLevel(level LogLevel) {
	l.lock_.Lock()
	defer l.lock_.Unlock()
	l.level_ = level
}

func (l *Logger) Log(level LogLevel, format string, args ...interface{}) {
	l.lock_.Lock()
	defer l.lock_.Unlock()
	if level < l.level_ {
		return
	}
	message := fmt.Sprintf(format, args...)
	logMessage := fmt.Sprintf("[%s] %s", levelToString(level), message)
	l.log_channel_ <- logMessage
}

func (l *Logger) Close() {
	close(l.log_channel_)
	l.waitgroup_.Wait()
	if l.log_file != nil {
		l.log_file.Close()
	}
}

func (l *Logger) output() {
	defer l.waitgroup_.Done()
	for logMessage := range l.log_channel_ {
		now := time.Now().Format("2006-01-02 15:04:05")
		msg := fmt.Sprintf("%s %s\n", now, logMessage)
		_, err := l.log_file.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Failed to write log to file: %v\n", err)
		}
	}
}

func levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

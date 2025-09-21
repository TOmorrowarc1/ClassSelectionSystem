package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
	Fatal
)

type Logger struct {
	level      LogLevel
	lock       sync.Mutex
	logChannel chan string
	waitGroup  sync.WaitGroup
	logFile    *os.File
}

var (
	logger_instance *Logger
	once            sync.Once
)

func GetLogger() *Logger {
	once.Do(func() {
		logger_instance = &Logger{
			level:      Debug,
			logChannel: make(chan string, 100),
		}
		logger_instance.waitGroup.Add(1)
		go logger_instance.output()
	})
	return logger_instance
}

func (l *Logger) SetLogFile(filePath string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.logFile != nil {
		l.logFile.Close()
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	l.logFile = file
	return nil
}

func (l *Logger) SetLogLevel(level LogLevel) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level
}

func (l *Logger) Log(level LogLevel, format string, args ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if level < l.level {
		return
	}
	message := fmt.Sprintf(format, args...)
	logMessage := fmt.Sprintf("[%s] %s", levelToString(level), message)
	l.logChannel <- logMessage
}

func (l *Logger) Close() {
	close(l.logChannel)
	l.waitGroup.Wait()
	if l.logFile != nil {
		l.logFile.Close()
	}
}

func (l *Logger) output() {
	defer l.waitGroup.Done()
	for logMessage := range l.logChannel {
		now := time.Now().Format("2006-01-02 15:04:05")
		msg := fmt.Sprintf("%s %s\n", now, logMessage)
		_, err := l.logFile.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Failed to write log to file: %v\n", err)
		}
	}
}

func levelToString(level LogLevel) string {
	switch level {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

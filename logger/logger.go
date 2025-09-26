package logger

import (
	"log"
	"os"
	"path/filepath"
)

type LogLevel int

const (
	DEBUG = 0
	INFO  = 1
	WARN  = 2
	ERROR = 3
)

type LogInfo struct {
	Err  error
	Info string
	Meta interface{}
}

type Logger struct {
	logLevel LogLevel
	logger   *log.Logger
}

type WebLogger interface {
	Debug(err LogInfo)
	Info(err LogInfo)
	Warn(err LogInfo)
	Error(err LogInfo)
}

func NewLogger(name string, outPath string, logLevel LogLevel) *Logger {
	dir, _ := filepath.Split(outPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Printf("Не удалось создать директорию для %s: %s", outPath, err)
	}

	file, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Не удалось создать лог файл %s: %s", outPath, err)
	}

	return &Logger{
		logLevel: logLevel,
		logger:   log.New(file, name+" ", log.LstdFlags),
	}
}

func (l *Logger) Error(info LogInfo) {
	l.logger.Printf("ERROR -- %s: %s %s", info.Info, info.Err, info.Meta)
}

func (l *Logger) Warn(info LogInfo) {
	if l.logLevel <= WARN {
		l.logger.Printf("WARN -- %s: %s %s", info.Info, info.Err, info.Meta)
	}
}

func (l *Logger) Info(info LogInfo) {
	if l.logLevel <= INFO {
		l.logger.Printf("INFO -- %s: %s", info.Info, info.Meta)
	}
}

func (l *Logger) Debug(info LogInfo) {
	if l.logLevel <= DEBUG {
		l.logger.Printf("DEBUG -- %s: %s", info.Info, info.Meta)
	}
}

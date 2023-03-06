package logger

import (
	"log"
	"os"
	"sync"
)

const (
	SilentLevel = iota
	ErrorLevel
	InfoLevel
)

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile)
	logLevel = InfoLevel
	mu       sync.Mutex
)

func SetLogLevel(level int) {
	mu.Lock()
	logLevel = level
	mu.Unlock()
}

func Error(v ...any) {
	if logLevel >= ErrorLevel {
		errorLog.Println(v...)
	}
}

func Errorf(format string, v ...any) {
	if logLevel >= ErrorLevel {
		errorLog.Printf(format, v...)
	}
}

func Info(v ...any) {
	if logLevel >= InfoLevel {
		infoLog.Println(v...)
	}
}

func Infof(format string, v ...any) {
	if logLevel >= InfoLevel {
		infoLog.Printf(format, v...)
	}
}

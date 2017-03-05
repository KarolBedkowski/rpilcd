//
// logging.go
//

package main

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
	logLevel int
}

func (l *Logger) SetLogLevel(level int) {
	l.logLevel = level
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.logLevel > 2 {
		l.Printf(format, v...)
	}
}

func (l *Logger) Debugln(v ...interface{}) {
	if l.logLevel > 2 {
		l.Println(v...)
	}
}

func (l *Logger) Debug(v ...interface{}) {
	if l.logLevel > 2 {
		l.Print(v...)
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.logLevel > 1 {
		l.Printf(format, v...)
	}
}

func (l *Logger) Infoln(v ...interface{}) {
	if l.logLevel > 1 {
		l.Println(v...)
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.logLevel > 1 {
		l.Print(v...)
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.logLevel > 0 {
		l.Printf(format, v...)
	}
}

func (l *Logger) Errorln(v ...interface{}) {
	if l.logLevel > 0 {
		l.Println(v...)
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.logLevel > 0 {
		l.Print(v...)
	}
}

var logger = Logger{
	log.New(os.Stdout, "", log.LstdFlags),
	1,
}

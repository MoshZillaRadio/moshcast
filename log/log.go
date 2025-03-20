// Copyright 2019 Setin Sergei
// Licensed under the Apache License, Version 2.0 (the "License")

package log

import (
	"log"
	"os"
	"strings"
)

type LogsLevel int

const (
	levelError   = 1
	levelWarning = 2
	levelInfo    = 3
	levelDebug   = 4
)

type iceLogger struct {
	level     LogsLevel
	logError  *log.Logger
	logAccess *log.Logger

	logErrorFile  *os.File
	logAccessFile *os.File
}

func NewLogger(level LogsLevel, logsPath string) (*iceLogger, error) {
	newLogger := &iceLogger{
		level: level,
	}

	errorFileName := logsPath + "error.log"
	accessFileName := logsPath + "access.log"

	var err error
	newLogger.logErrorFile, err = os.OpenFile(errorFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	newLogger.logAccessFile, err = os.OpenFile(accessFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	newLogger.logError = log.New(newLogger.logErrorFile, "", log.Ldate|log.Ltime)
	newLogger.logAccess = log.New(newLogger.logAccessFile, "", 0)

	return newLogger, nil
}

func (l *iceLogger) output(errorLevel LogsLevel, format string, v ...interface{}) {
	out := strings.Builder{}
	if errorLevel <= l.level {
		switch errorLevel {
		case 1:
			out.WriteString("E: ")
		case 2:
			out.WriteString("W: ")
		case 3:
			out.WriteString("I: ")
		case 4:
			out.WriteString("D: ")
		}
		out.WriteString(format)
		l.logError.Printf(out.String(), v...)
	}
}

func (l *iceLogger) Error(format string, v ...interface{}) {
	l.output(levelError, format, v...)
}

func (l *iceLogger) Debug(format string, v ...interface{}) {
	l.output(levelDebug, format, v...)
}

func (l *iceLogger) Warning(format string, v ...interface{}) {
	l.output(levelWarning, format, v...)
}

func (l *iceLogger) Info(format string, v ...interface{}) {
	l.output(levelInfo, format, v...)
}

func (l *iceLogger) Access(format string, v ...interface{}) {
	l.logAccess.Printf(format, v...)
}

func (l *iceLogger) Log(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *iceLogger) Close() {
	_ = l.logErrorFile.Close()
	_ = l.logAccessFile.Close()
}

package main

import "log"

type Logger interface {
	Printf(format string, v ...interface{})
}

type nopLogger struct{}
type errLogger struct{}

func (l nopLogger) Printf(format string, v ...interface{}) {
}

func (l errLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

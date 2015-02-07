package godispatch

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Debugf(s string, args ...interface{})
	Debug(s string)
	Infof(s string, args ...interface{})
	Info(s string)
	Warnf(s string, args ...interface{})
	Warn(s string)
	Errorf(s string, args ...interface{})
	Error(s string)
}

type StandardLogger struct {
	log *log.Logger
}

func NewStandardLogger() *StandardLogger {
	return &StandardLogger{
		log: log.New(os.Stdout, "", log.Ltime),
	}
}

func (me *StandardLogger) Debugf(s string, args ...interface{}) {
	me.log.Println(fmt.Sprintf("\x1b[32mdebug\x1b[0m "+s, args...))
}
func (me *StandardLogger) Debug(s string) {
	me.log.Println("\x1b[32mdebug\x1b[0m " + s)
}
func (me *StandardLogger) Infof(s string, args ...interface{}) {
	me.log.Println(fmt.Sprintf("\x1b[34minfo\x1b[0m  "+s, args...))
}
func (me *StandardLogger) Info(s string) {
	me.log.Println("\x1b[34minfo\x1b[0m  " + s)
}
func (me *StandardLogger) Warnf(s string, args ...interface{}) {
	me.log.Println(fmt.Sprintf("\x1b[33mwarn\x1b[0m  "+s, args...))
}
func (me *StandardLogger) Warn(s string) {
	me.log.Println("\x1b[33mwarn\x1b[0m  " + s)
}
func (me *StandardLogger) Errorf(s string, args ...interface{}) {
	me.log.Println(fmt.Sprintf("\x1b[31merror\x1b[0m "+s, args...))
}
func (me *StandardLogger) Error(s string) {
	me.log.Println("\x1b[31merror\x1b[0m " + s)
}

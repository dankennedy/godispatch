package godispatch

import (
	"fmt"
	"io"
	"log"
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
	WithPrefix(s string) Logger
}

type StandardLogger struct {
	log      *log.Logger
	prefix   string
	levelmap map[string]int
}

func NewStandardLogger(writer io.Writer) *StandardLogger {
	return &StandardLogger{
		log: log.New(writer, "", log.Ltime),
		levelmap: map[string]int{
			"debug": 32,
			"info":  34,
			"warn":  33,
			"error": 31,
		},
	}
}

func (me *StandardLogger) WithPrefix(s string) Logger {
	me.prefix = s + " "
	return me
}
func (me *StandardLogger) logimpl(level, s string, args ...interface{}) {
	me.log.Println(fmt.Sprintf("\x1b[%dm%-6s\x1b[0m%s%s", me.levelmap[level], level, me.prefix, fmt.Sprintf(s, args...)))
}
func (me *StandardLogger) Debugf(s string, args ...interface{}) {
	me.logimpl("debug", s, args...)
}
func (me *StandardLogger) Debug(s string) {
	me.logimpl("debug", s)
}
func (me *StandardLogger) Infof(s string, args ...interface{}) {
	me.logimpl("info", s, args...)
}
func (me *StandardLogger) Info(s string) {
	me.logimpl("info", s)
}
func (me *StandardLogger) Warnf(s string, args ...interface{}) {
	me.logimpl("warn", s, args...)
}
func (me *StandardLogger) Warn(s string) {
	me.logimpl("warn", s)
}
func (me *StandardLogger) Errorf(s string, args ...interface{}) {
	me.logimpl("error", s, args...)
}
func (me *StandardLogger) Error(s string) {
	me.logimpl("error", s)
}

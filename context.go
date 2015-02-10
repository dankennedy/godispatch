package godispatch

import (
	"bytes"
	"fmt"
	"math"

	"github.com/streadway/amqp"
)

const (
	AbortIndex = math.MaxInt8 / 2
)

type errorMsg struct {
	Err  string      `json:"error"`
	Meta interface{} `json:"meta"`
}

type errorMsgs []errorMsg

func (a errorMsgs) String() string {
	if len(a) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	for i, msg := range a {
		text := fmt.Sprintf("Error #%02d: %s \n     Meta: %v\n", (i + 1), msg.Err, msg.Meta)
		buffer.WriteString(text)
	}
	return buffer.String()
}

type Context struct {
	Delivery *amqp.Delivery
	Message  interface{}
	Log      Logger
	Keys     map[string]interface{}
	Errors   errorMsgs
	handlers []HandlerFunc
	index    int8
}

func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Abort() {
	c.index = AbortIndex
}

func (c *Context) Fail(err error) {
	c.Error(err, "Operation aborted")
	c.Abort()
}

func (c *Context) Error(err error, meta interface{}) {
	c.Errors = append(c.Errors, errorMsg{
		Err:  err.Error(),
		Meta: meta,
	})
}

package godispatch

import (
	"os"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func CreateTestBus() *Bus {
	return &Bus{
		conf: &BusConfig{
			Url:        "amqp://mystacklocal:localNotProduction!@mystack.vm/mystack",
			Queue:      "agent_actions",
			Exchange:   "agent_actions",
			RoutingKey: "",
		},
		log: NewStandardLogger(os.Stdout).WithPrefix("[bus]"),
	}
}

func TestConnect(t *testing.T) {
	bus := CreateTestBus()
	if err := bus.Connect(); err != nil {
		t.Error("%v", err)
	} else {
		go bus.Run()
		time.Sleep(1 * time.Second)
		bus.Close()
	}
}

func TestSimpleRoute(t *testing.T) {
	passed := false
	bus := CreateTestBus()
	bus.Handle("messagetype1", []HandlerFunc{func(c *Context) {
		passed = true
	}})

	bus.ProcessMessage(&amqp.Delivery{ContentType: "messagetype1"})

	if !passed {
		t.Errorf("route handler was not invoked.")
	}
}

func TestSingleMiddleware(t *testing.T) {
	middleware, msg := false, false
	bus := CreateTestBus()
	bus.Use(func(c *Context) {
		middleware = true
		c.Next()
	})
	bus.Handle("messagetype1", []HandlerFunc{func(c *Context) {
		msg = true
	}})

	bus.ProcessMessage(&amqp.Delivery{ContentType: "messagetype1"})

	if !middleware {
		t.Errorf("middleware was not invoked.")
	}

	if !msg {
		t.Errorf("route handler was not invoked.")
	}
}

func TestMultiMiddleware(t *testing.T) {
	middleware1, middleware2, middleware3, msg := false, false, false, false
	bus := CreateTestBus()
	bus.Use(func(c *Context) {
		middleware1 = true
		c.Next()
	})
	bus.Use(func(c *Context) {
		middleware2 = true
		c.Next()
	})
	bus.Use(func(c *Context) {
		middleware3 = true
		c.Next()
	})
	bus.Handle("messagetype1", []HandlerFunc{func(c *Context) {
		msg = true
	}})

	bus.ProcessMessage(&amqp.Delivery{ContentType: "messagetype1"})

	if !middleware1 {
		t.Errorf("middleware1 was not invoked.")
	}
	if !middleware2 {
		t.Errorf("middleware2 was not invoked.")
	}
	if !middleware3 {
		t.Errorf("middleware3 was not invoked.")
	}

	if !msg {
		t.Errorf("route handler was not invoked.")
	}
}

func TestMiddlewareAbort(t *testing.T) {
	middleware, msg := false, false
	bus := CreateTestBus()
	bus.Use(func(c *Context) {
		middleware = true
		c.Abort()
	})
	bus.Handle("messagetype1", []HandlerFunc{func(c *Context) {
		msg = true
	}})

	bus.ProcessMessage(&amqp.Delivery{ContentType: "messagetype1"})

	if !middleware {
		t.Errorf("middleware was not invoked.")
	}

	if msg {
		t.Errorf("route handler was invoked.")
	}
}

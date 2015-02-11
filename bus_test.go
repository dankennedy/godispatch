package godispatch

import (
	"os"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

var delayBus = 100 * time.Millisecond

type DeliveryAcknowledgerStub struct {
	HasAckd     bool
	HasNackd    bool
	HasRejected bool
}

func (me *DeliveryAcknowledgerStub) Ack(tag uint64, multiple bool) error {
	me.HasAckd = true
	return nil
}
func (me *DeliveryAcknowledgerStub) Nack(tag uint64, multiple bool, requeue bool) error {
	me.HasNackd = true
	return nil
}
func (me *DeliveryAcknowledgerStub) Reject(tag uint64, requeue bool) error {
	me.HasRejected = true
	return nil
}

func createDummyDelivery(contentType string) *amqp.Delivery {
	return &amqp.Delivery{
		ContentType:  contentType,
		Acknowledger: &DeliveryAcknowledgerStub{},
	}
}

func CreateTestBus() *Bus {
	return &Bus{
		// uses rabbit docker container
		// docker run -d -p 5672:5672 -p 15672:15672 dockerfile/rabbitmq
		// but you could use any rabbitmq instance you can connect to
		conf: &BusConfig{
			Url:                       "amqp://guest:guest@192.168.59.103",
			InputQueue:                "input_queue",
			ErrorQueue:                "error_queue",
			RetryQueue:                "retry_queue",
			RetryIntervalMilliseconds: 1000,
			RetryLimit:                2,
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
		time.Sleep(delayBus)
		bus.Close()
	}
}

func TestSendAndReceive(t *testing.T) {
	bus := CreateTestBus()
	if err := bus.Connect(); err != nil {
		t.Error("%v", err)
	} else {
		ok := false
		bus.Handle("sendandreceivemsg", []HandlerFunc{func(c *Context) {
			ok = true
		}})
		bus.Publish(bus.conf, "sendandreceivemsg")
		go bus.Run()
		time.Sleep(delayBus)
		if !ok {
			t.Errorf("message not received.")
		}
		bus.Close()
	}
}

func TestSendAndError(t *testing.T) {
	bus := CreateTestBus()
	if err := bus.Connect(); err != nil {
		t.Error("%v", err)
	} else {
		bus.SendToError(createDummyDelivery("errormessage"))
		time.Sleep(delayBus)
		bus.Close()
	}
}

func TestDeferAndReceive(t *testing.T) {
	bus := CreateTestBus()
	if err := bus.Connect(); err != nil {
		t.Error("%v", err)
	} else {
		ok := false
		bus.Handle("sendandreceivemsg", []HandlerFunc{func(c *Context) {
			ok = true
		}})
		bus.Defer(createDummyDelivery("sendandreceivemsg"))
		go bus.Run()
		time.Sleep(1200 * time.Millisecond)
		if !ok {
			t.Errorf("message not received.")
		}
		bus.Close()
	}
}

func TestSimpleRoute(t *testing.T) {
	passed := false
	bus := CreateTestBus()
	bus.Handle("messagetype1", []HandlerFunc{func(c *Context) {
		passed = true
	}})

	bus.ProcessMessage(createDummyDelivery("messagetype1"))

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

	bus.ProcessMessage(createDummyDelivery("messagetype1"))

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
	}, func(c *Context) {
		middleware2 = true
		c.Next()
	}, func(c *Context) {
		middleware3 = true
		c.Next()
	})
	bus.Handle("messagetype1", []HandlerFunc{func(c *Context) {
		msg = true
	}})

	bus.ProcessMessage(createDummyDelivery("messagetype1"))

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

	bus.ProcessMessage(createDummyDelivery("messagetype1"))

	if !middleware {
		t.Errorf("middleware was not invoked.")
	}

	if msg {
		t.Errorf("route handler was invoked.")
	}
}

func TestHandlerRecovery(t *testing.T) {
	msg := false
	bus := CreateTestBus()
	bus.Use(Recovery())
	bus.Handle("messagetype1", []HandlerFunc{func(c *Context) {
		msg = true
		panic("something horrible went wrong")
	}})

	bus.ProcessMessage(createDummyDelivery("messagetype1"))

	if !msg {
		t.Errorf("route handler was not invoked.")
	}
}

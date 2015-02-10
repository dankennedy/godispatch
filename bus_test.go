package godispatch

import (
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	bus := &Bus{
		conf: &BusConfig{
			Url: "amqp://mystacklocal:localNotProduction!@mystack.vm/mystack",
		},
		log: NewStandardLogger(os.Stdout).WithPrefix("bus_test"),
	}
	if err := bus.Connect(); err != nil {
		t.Error("%v", err)
	}
}

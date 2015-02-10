package godispatch

import (
	"os"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	bus := &Bus{
		conf: &BusConfig{
			Url:        "amqp://mystacklocal:localNotProduction!@mystack.vm/mystack",
			Queue:      "agent_actions",
			Exchange:   "agent_actions",
			RoutingKey: "",
		},
		log: NewStandardLogger(os.Stdout).WithPrefix("[bus]"),
	}
	if err := bus.Connect(); err != nil {
		t.Error("%v", err)
	} else {
		go bus.Run()
		time.Sleep(1 * time.Second)
		bus.Close()
	}
}

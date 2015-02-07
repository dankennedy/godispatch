package godispatch

import "github.com/streadway/amqp"

type Context struct {
	Delivery *amqp.Delivery
	Message  interface{}
}

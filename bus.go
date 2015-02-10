package godispatch

import (
	"sync"

	"github.com/streadway/amqp"
)

type Bus struct {
	conf       *BusConfig
	log        Logger
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
	messages   <-chan amqp.Delivery
	doneCh     chan struct{}
	wg         sync.WaitGroup
}

type BusConfig struct {
	Url        string
	Exchange   string
	Queue      string
	RoutingKey string
}

func (bus *Bus) Connect() error {

	var err error

	bus.log.Info("Connecting...")
	if bus.connection, err = amqp.Dial(bus.conf.Url); err != nil {
		return err
	}

	if bus.channel, err = bus.connection.Channel(); err != nil {
		return err
	}

	bus.log.Info("Connected")

	if err = bus.channel.Qos(1, 0, false); err != nil {
		return err
	}

	if err = bus.channel.ExchangeDeclare(bus.conf.Exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		true,     // nowait
		nil,      // args
	); err != nil {
		return err
	}

	bus.log.Debugf("Declared exchange '%s'", bus.conf.Exchange)

	if bus.queue, err = bus.channel.QueueDeclare(bus.conf.Queue, // name
		true,  // durable
		false, // auto-deleted
		false, // exclusive
		true,  // nowait
		nil,   // args
	); err != nil {
		return err
	}

	bus.log.Debugf("Declared queue '%s'", bus.conf.Queue)

	if err = bus.channel.QueueBind(bus.conf.Queue, // queue
		bus.conf.RoutingKey, // bindingKey
		bus.conf.Exchange,   // sourceExchange
		true,                // noWait
		nil,                 // args
	); err != nil {
		return err
	}

	bus.log.Debugf("Bound '%s' queue to '%s' exchange with routing key '%s'",
		bus.conf.Queue,
		bus.conf.Exchange,
		bus.conf.RoutingKey)

	if bus.messages, err = bus.channel.Consume(
		bus.conf.Queue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		true,  // no-wait
		nil,   // args
	); err != nil {
		return err
	}

	bus.log.Infof("Receiving messages from '%s'", bus.conf.Queue)

	return nil
}

func (bus *Bus) Run() {
	bus.wg.Add(1)
	defer bus.wg.Done()
	bus.doneCh = make(chan struct{})
	bus.log.Info("Running")
	for {
		select {
		case <-bus.doneCh:
			return
		case msg, ok := <-bus.messages:
			if !ok {
				// reconnect or close here?
				bus.log.Info("Not OK")
				return
			}
			bus.ProcessMessage(&msg)
		}
	}
}

func (bus *Bus) Close() {
	bus.log.Info("Closing")

	close(bus.doneCh)
	bus.wg.Wait()
	bus.channel.Close()
	bus.connection.Close()
}

func (bus *Bus) ProcessMessage(msg *amqp.Delivery) {

	bus.log.Info("Processing message")
}

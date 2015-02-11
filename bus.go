package godispatch

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type HandlerFunc func(*Context)

type Bus struct {
	conf       *BusConfig
	log        Logger
	connection *amqp.Connection
	channel    *amqp.Channel
	inputQ     amqp.Queue
	retryQ     amqp.Queue
	errorQ     amqp.Queue
	messages   <-chan amqp.Delivery
	doneCh     chan struct{}
	wg         sync.WaitGroup
	routes     map[string][]HandlerFunc
	middleware []HandlerFunc
}

type BusConfig struct {
	Url                       string
	InputQueue                string
	ErrorQueue                string
	RetryQueue                string
	RetryIntervalMilliseconds int32
	RetryLimit                int32
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

	if bus.inputQ, err = bus.declareAndBind(bus.conf.InputQueue,
		bus.conf.InputQueue,
		bus.conf.InputQueue,
		nil); err != nil {
		return err
	}

	if bus.errorQ, err = bus.declareAndBind(bus.conf.ErrorQueue,
		bus.conf.ErrorQueue,
		bus.conf.ErrorQueue,
		nil); err != nil {
		return err
	}

	queueArgs := amqp.Table{
		"x-dead-letter-exchange":    bus.conf.InputQueue,
		"x-dead-letter-routing-key": bus.conf.InputQueue,
		"x-message-ttl":             bus.conf.RetryIntervalMilliseconds,
	}

	if bus.retryQ, err = bus.declareAndBind(bus.conf.RetryQueue,
		bus.conf.RetryQueue,
		bus.conf.RetryQueue,
		queueArgs); err != nil {
		return err
	}

	if bus.messages, err = bus.channel.Consume(
		bus.conf.InputQueue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		true,  // no-wait
		nil,   // args
	); err != nil {
		return err
	}

	bus.log.Infof("Receiving messages from '%s'", bus.conf.InputQueue)

	return nil
}

func (bus *Bus) declareAndBind(queue, exchange, routingKey string, args amqp.Table) (q amqp.Queue, err error) {

	if err = bus.channel.ExchangeDeclare(exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		true,     // no-wait
		nil,      // arguments)
	); err != nil {
		return q, err
	}

	bus.log.Debugf("Declared exchange '%s'", exchange)

	if q, err = bus.channel.QueueDeclare(queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		true,  // nowait
		args,  // args
	); err != nil {
		return q, err
	}

	bus.log.Debugf("Declared queue '%s'", queue)

	if err = bus.channel.QueueBind(queue, // queue
		routingKey, // bindingKey
		exchange,   // sourceExchange
		true,       // noWait
		nil,        // args
	); err != nil {
		return q, err
	}

	bus.log.Debugf("Bound '%s' queue to '%s' exchange with routing key '%s'",
		queue,
		exchange,
		routingKey)

	return q, nil
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
			go bus.ProcessMessage(&msg)
		}
	}
}

func (bus *Bus) Close() {
	bus.log.Info("Closing")

	if bus.doneCh != nil {
		close(bus.doneCh)
	}

	bus.wg.Wait()
	bus.channel.Close()
	bus.connection.Close()
}

func (bus *Bus) ProcessMessage(msg *amqp.Delivery) {

	bus.wg.Add(1)
	defer bus.wg.Done()

	if handlers, found := bus.routes[msg.ContentType]; !found {
		bus.log.Warnf("No route registered for message type %s", msg.ContentType)
	} else {
		handlers = bus.combineHandlers(handlers)
		bus.createContext(msg, handlers).Next()
	}

	if err := msg.Ack(false); err != nil {
		bus.log.Errorf("Failed to acknowledge %s message. %v", msg.ContentType, err)
	}
}

func (bus *Bus) Use(middlewares ...HandlerFunc) {
	bus.middleware = append(bus.middleware, middlewares...)
}

func (bus *Bus) Handle(msgType string, handlers []HandlerFunc) {
	if bus.routes == nil {
		bus.routes = map[string][]HandlerFunc{msgType: handlers}
	} else {
		if _, ok := bus.routes[msgType]; ok {
			bus.log.Warnf("Overwriting route for %s", msgType)
		}
		bus.routes[msgType] = handlers
	}
}

func (bus *Bus) Publish(msg interface{}, msgType string) error {

	bodyBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return bus.sendToQueue(msgType, bus.conf.InputQueue, bodyBytes)
}

func (bus *Bus) SendToError(msg *amqp.Delivery) error {
	return bus.sendToQueue(msg.ContentType, bus.conf.ErrorQueue, msg.Body)
}

func (bus *Bus) Defer(msg *amqp.Delivery) error {
	return bus.sendToQueue(msg.ContentType, bus.conf.RetryQueue, msg.Body)
}

func (bus *Bus) sendToQueue(msgType, queue string, body []byte) error {
	publishing := amqp.Publishing{
		ContentType: msgType,
		Body:        body,
		Timestamp:   time.Now(),
	}

	return bus.channel.Publish(
		queue, // exchange
		queue, // routing key
		true,  // mandatory
		false, // immediate
		publishing)
}

func (bus *Bus) createContext(msg *amqp.Delivery, handlers []HandlerFunc) *Context {
	return &Context{
		Delivery: msg,
		Log:      bus.log,
		Keys:     make(map[string]interface{}),
		Errors:   []errorMsg{},
		handlers: handlers,
		index:    -1,
	}
}

func (bus *Bus) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	s := len(bus.middleware) + len(handlers)
	h := make([]HandlerFunc, 0, s)
	h = append(h, bus.middleware...)
	h = append(h, handlers...)
	return h
}

package amqp

import (
	"context"
	"fmt"
	"strings"

	"github.com/MouseHatGames/mice/broker"
	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	"github.com/streadway/amqp"
)

const exchangeName = "mice-amqp"

type rabbitmqBroker struct {
	addr string
	conn *amqp.Connection
	ch   *amqp.Channel
	log  logger.Logger

	boundQueues map[string]string
	connection  chan bool // Will be closed on connection
}

// Broker instructs Mice to use AMQP as the message broker. Must be used after setting up a codec.
func Broker(addr string) options.Option {
	return func(o *options.Options) {
		o.Broker = &rabbitmqBroker{
			addr:        addr,
			log:         o.Logger.GetLogger("amqp"),
			boundQueues: make(map[string]string),
			connection:  make(chan bool),
		}
	}
}

func (r *rabbitmqBroker) Close() error {
	return r.conn.Close()
}

func (r *rabbitmqBroker) Start() error {
	if !strings.HasPrefix(r.addr, "amqp://") {
		r.addr = "amqp://" + r.addr
	}

	conn, err := amqp.Dial(r.addr)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("amqp channel: %w", err)
	}

	r.log.Infof("connected")

	r.conn = conn
	r.ch = ch

	if err := r.declareExchange(); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	close(r.connection)
	return nil
}

func (r *rabbitmqBroker) declareExchange() error {
	r.log.Debugf("declaring exchange %s", exchangeName)

	if err := r.ch.ExchangeDeclare(exchangeName, "topic", false, false, false, false, nil); err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}

	return nil
}

func (r *rabbitmqBroker) bindQueue(topic string) (queueName string, err error) {
	if n, ok := r.boundQueues[topic]; ok {
		return n, nil
	}

	q, err := r.ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return "", fmt.Errorf("declare queue: %w", err)
	}

	r.log.Debugf("declared queue %s", q.Name)

	if err := r.ch.QueueBind(q.Name, topic, exchangeName, false, nil); err != nil {
		return "", fmt.Errorf("bind queue: %w", err)
	}

	r.boundQueues[topic] = q.Name
	return q.Name, nil
}

func (r *rabbitmqBroker) unbindQueue(topic string) error {
	qname, ok := r.boundQueues[topic]
	if !ok {
		return fmt.Errorf("queue for topic %s does not exist", topic)
	}

	_, err := r.ch.QueueDelete(qname, false, false, false)
	if err != nil {
		return fmt.Errorf("queue delete: %w", err)
	}

	return nil
}

func (r *rabbitmqBroker) Publish(ctx context.Context, topic string, msg *broker.Message) error {
	err := r.ch.Publish(exchangeName, topic, false, false, amqp.Publishing{
		Body: msg.Data,
	})
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}

func (r *rabbitmqBroker) Subscribe(ctx context.Context, topic string, callback func(*broker.Message)) error {
	// If we aren't yet connected, wait for connection on a goroutine then subscribe again with the same parameters
	if r.conn == nil {
		go func() {
			<-r.connection
			if err := r.Subscribe(ctx, topic, callback); err != nil {
				r.log.Errorf("deferred subscription failed: %s", err)
			}
		}()
		return nil
	}

	q, err := r.bindQueue(topic)
	if err != nil {
		return fmt.Errorf("create queue: %w", err)
	}

	msgs, err := r.ch.Consume(q, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	r.log.Infof("subscribed to topic %s", topic)

	done := ctx.Done()

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				callback(&broker.Message{Data: msg.Body})

			case <-done:
				return
			}
		}
	}()

	return nil
}

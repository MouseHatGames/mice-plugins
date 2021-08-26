package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/MouseHatGames/mice/broker"
	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	"github.com/go-redis/redis/v8"
)

type redisBroker struct {
	rdb *redis.Client
	log logger.Logger
}

func Broker(ropt *redis.Options) options.Option {
	return BrokerClient(redis.NewClient(ropt))
}

func BrokerClient(cl *redis.Client) options.Option {
	return func(o *options.Options) {
		o.Broker = &redisBroker{
			rdb: cl,
			log: o.Logger.GetLogger("redispb"),
		}
	}
}

func (r *redisBroker) Close() error {
	return r.rdb.Close()
}

func (r *redisBroker) Publish(ctx context.Context, topic string, data *broker.Message) error {
	cmd := r.rdb.Publish(ctx, topic, data.Data)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	r.log.Debugf("published message %s to %d receivers", topic, cmd.Val())
	return nil
}

func (r *redisBroker) Subscribe(ctx context.Context, topic string, callback func(*broker.Message)) error {
	pubsub := r.rdb.Subscribe(ctx, topic)

	rectx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()

	_, err := pubsub.Receive(rectx)
	if err != nil {
		return fmt.Errorf("subscribe to topic: %w", err)
	}

	go func() {
		c := pubsub.Channel()

		for msg := range c {
			callback(&broker.Message{Data: []byte(msg.Payload)})
		}
	}()

	return nil
}

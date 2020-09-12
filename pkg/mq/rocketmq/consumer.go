package rocketmq

import (
	"context"
	"fmt"
	"strings"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

type ConsumerConfig struct {
	Endpoint  []string
	AccessKey string
	SecretKey string
	Namespace string

	Topic string
	Gid   string
	Tags  []string

	Orderly      bool
	interceptors []primitive.Interceptor
}

type PushConsumer struct {
	consumer rocketmq.PushConsumer
	config   *ConsumerConfig
}

func NewPushConsumer(config *ConsumerConfig) *PushConsumer {
	var credentials = primitive.Credentials{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	}

	consumer, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(config.Endpoint),
		consumer.WithCredentials(credentials),
		consumer.WithNamespace(config.Namespace),
		consumer.WithGroupName(config.Gid),
		consumer.WithConsumerOrder(config.Orderly),
		consumer.WithInterceptor(pushConsumerMetricInterceptor(config)),
		consumer.WithInterceptor(config.interceptors...),
	)

	if err != nil {
		panic(fmt.Sprintf("init consumer fail, err msg: %s", err.Error()))
	}

	pc := &PushConsumer{
		consumer: consumer,
		config:   config,
	}

	return pc
}

func (pc *PushConsumer) Start() error {
	return pc.consumer.Start()
}

func (pc *PushConsumer) Shutdown() error {
	return pc.consumer.Shutdown()
}

func (pc *PushConsumer) Subscribe(cb func(context.Context, *primitive.MessageExt) error) *PushConsumer {
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: strings.Join(pc.config.Tags, "||"),
	}

	err := pc.consumer.Subscribe(pc.config.Topic, selector, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			err := cb(ctx, msg)
			if err != nil {
				return consumer.ConsumeRetryLater, err
			}
		}

		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		panic(fmt.Sprintf("subscribe rocketmq fail, err msg: %s", err.Error()))
	}

	return pc
}

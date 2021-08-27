/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package rocketmq

import (
	"context"
	"fmt"
	"strings"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// ConsumerConfig RocketMQ consumer config
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

// PushConsumer push consumer mode
type PushConsumer struct {
	consumer rocketmq.PushConsumer
	config   *ConsumerConfig
}

// NewPushConsumer returns a PushConsumer instance
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

// Start start consumer
func (pc *PushConsumer) Start() error {
	return pc.consumer.Start()
}

// Shutdown consumer
func (pc *PushConsumer) Shutdown() error {
	return pc.consumer.Shutdown()
}

// Subscribe subscribe topic
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

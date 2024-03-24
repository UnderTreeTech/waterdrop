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
	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"strings"
	"time"
)

// ConsumerConfig RocketMQ consumer config
type ConsumerConfig struct {
	Endpoint  []string
	AccessKey string
	SecretKey string

	Topic string
	Gid   string
	Tags  []string
	Retry int32

	Orderly      bool
	interceptors []primitive.Interceptor

	PullTimeout   time.Duration
	PullBatchSize int32
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
	opts := []consumer.Option{
		consumer.WithNsResolver(primitive.NewPassthroughResolver(config.Endpoint)),
		consumer.WithCredentials(credentials),
		consumer.WithGroupName(config.Gid),
		consumer.WithConsumerOrder(config.Orderly),
		consumer.WithMaxReconsumeTimes(config.Retry),
		consumer.WithInterceptor(consumerMetricInterceptor(config)),
		consumer.WithInterceptor(config.interceptors...),
	}
	if config.PullBatchSize > 0 && config.PullBatchSize <= 1024 {
		opts = append(opts, consumer.WithPullBatchSize(config.PullBatchSize))
	} else {
		opts = append(opts, consumer.WithPullBatchSize(32))
	}
	consumer, err := rocketmq.NewPushConsumer(opts...)
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

// PullConsumer pull consumer mode
type PullConsumer struct {
	consumer rocketmq.PullConsumer
	config   *ConsumerConfig
	cb       func(context.Context, *primitive.MessageExt) error
}

// NewPullConsumer returns a PullConsumer instance
func NewPullConsumer(config *ConsumerConfig) *PullConsumer {
	if config.PullTimeout == 0 {
		config.PullTimeout = time.Minute // default poll every 1min
	}
	var credentials = primitive.Credentials{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	}

	opts := []consumer.Option{
		consumer.WithNsResolver(primitive.NewPassthroughResolver(config.Endpoint)),
		consumer.WithCredentials(credentials),
		consumer.WithGroupName(config.Gid),
		consumer.WithConsumerOrder(config.Orderly),
		consumer.WithMaxReconsumeTimes(config.Retry),
		consumer.WithInterceptor(consumerMetricInterceptor(config)),
		consumer.WithInterceptor(config.interceptors...),
	}
	if config.PullBatchSize > 0 && config.PullBatchSize <= 1024 {
		opts = append(opts, consumer.WithPullBatchSize(config.PullBatchSize))
	} else {
		opts = append(opts, consumer.WithPullBatchSize(32))
	}
	consumer, err := rocketmq.NewPullConsumer(opts...)
	if err != nil {
		panic(fmt.Sprintf("init consumer fail, err msg: %s", err.Error()))
	}

	pc := &PullConsumer{
		consumer: consumer,
		config:   config,
	}
	return pc
}

// Start start consumer
func (pc *PullConsumer) Start() error {
	err := pc.consumer.Start()
	if err != nil {
		return err
	}
	go pc.poll(context.Background())
	return nil
}

// Shutdown consumer
func (pc *PullConsumer) Shutdown() error {
	return pc.consumer.Shutdown()
}

// Subscribe subscribe topic
func (pc *PullConsumer) Subscribe(cb func(context.Context, *primitive.MessageExt) error) error {
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: strings.Join(pc.config.Tags, "||"),
	}
	pc.cb = cb
	return pc.consumer.Subscribe(pc.config.Topic, selector)
}

// poll messages with timeout
func (pc *PullConsumer) poll(ctx context.Context) {
	for {
		cr, err := pc.consumer.Poll(ctx, pc.config.PullTimeout)
		if err != nil {
			continue
		}

		for _, msg := range cr.GetMsgList() {
			err = pc.cb(ctx, msg)
			if err != nil {
				break
			}
		}
		if err != nil {
			pc.consumer.ACK(context.Background(), cr, consumer.ConsumeRetryLater)
			continue
		}
		pc.consumer.ACK(context.Background(), cr, consumer.ConsumeSuccess)
	}
}

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
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

// ProducerConfig RocketMQ producer config
type ProducerConfig struct {
	Endpoint  []string
	AccessKey string
	SecretKey string

	Retry       int
	SendTimeout time.Duration

	Topic string

	interceptors []primitive.Interceptor

	SlowSendDuration time.Duration
}

// Producer producer config
type Producer struct {
	producer rocketmq.Producer
	config   *ProducerConfig
}

// NewProducer returns a Producer instance
func NewProducer(config *ProducerConfig) *Producer {
	var credentials = primitive.Credentials{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	}

	producer, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(config.Endpoint)),
		producer.WithRetry(config.Retry),
		producer.WithSendMsgTimeout(config.SendTimeout),
		producer.WithCredentials(credentials),
		producer.WithInterceptor(producerMetricInterceptor(config)),
		producer.WithInterceptor(config.interceptors...),
	)

	if err != nil {
		panic(fmt.Sprintf("new producer fail, err msg: %s", err.Error()))
	}

	p := &Producer{
		producer: producer,
		config:   config,
	}
	return p
}

// Start start producer
func (p *Producer) Start() error {
	return p.producer.Start()
}

// Shutdown producer
func (p *Producer) Shutdown() error {
	return p.producer.Shutdown()
}

// SendSyncMsg send message sync
func (p *Producer) SendSyncMsg(ctx context.Context, content string, tags ...string) error {
	msgs := getSendMsgs(p.config.Topic, content, tags...)
	_, err := p.producer.SendSync(ctx, msgs...)
	if err != nil {
		log.Error(ctx, "send msg fail", log.String("content", content), log.Any("tags", tags),
			log.String("error", err.Error()))
		return err
	}
	return nil
}

// SendAsyncMsg send message async
func (p *Producer) SendAsyncMsg(ctx context.Context, content string, callback func(context.Context, *primitive.SendResult, error), tags ...string) error {
	msgs := getSendMsgs(p.config.Topic, content, tags...)
	err := p.producer.SendAsync(ctx, callback, msgs...)
	if err != nil {
		log.Error(ctx, "async send msg fail", log.String("content", content), log.String("error", err.Error()))
		return err
	}
	return nil
}

// getSendMsgs format send message to primitive.Message
func getSendMsgs(topic string, content string, tags ...string) []*primitive.Message {
	var msgs []*primitive.Message
	bs := xstring.StringToBytes(content)

	if 0 == len(tags) {
		msgs = make([]*primitive.Message, 0, 1)
		msgs[0] = primitive.NewMessage(topic, bs).
			WithKeys([]string{xstring.RandomString(16)})
	} else {
		msgs = make([]*primitive.Message, 0, len(tags))
		for _, tag := range tags {
			msg := primitive.NewMessage(topic, bs).
				WithTag(tag).WithKeys([]string{xstring.RandomString(16)})
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

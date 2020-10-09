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

package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/Shopify/sarama"
)

type ConsumerConfig struct {
	Addr  []string
	Topic []string
	Gid   string

	EnableSASLAuth bool
	SASLMechanism  string
	SASLUser       string
	SASLPassword   string
	SASLHandshake  bool

	DialTimeout time.Duration

	ConsumeOldest     bool
	EnableReturnError bool

	ClientID string
}

type Consumer struct {
	consumer    sarama.ConsumerGroup
	subscribers []ConsumerHandler
	config      *ConsumerConfig
}

type ConsumerHandler func(context.Context, *sarama.ConsumerMessage) error

func NewConsumer(config *ConsumerConfig) *Consumer {
	sconfig := newKafkaConsumerConfig(config)
	consumer, err := sarama.NewConsumerGroup(config.Addr, config.Gid, sconfig)
	if err != nil {
		panic(fmt.Sprintf("create kafka consumer fail, err msg:%s", err.Error()))
	}

	c := &Consumer{
		consumer:    consumer,
		config:      config,
		subscribers: make([]ConsumerHandler, 0),
	}

	return c
}

func newKafkaConsumerConfig(config *ConsumerConfig) *sarama.Config {
	sconfig := sarama.NewConfig()
	sconfig.Net.SASL.Enable = config.EnableSASLAuth
	sconfig.Net.SASL.Mechanism = sarama.SASLMechanism(config.SASLMechanism)
	sconfig.Net.SASL.User = config.SASLUser
	sconfig.Net.SASL.Password = config.SASLPassword
	sconfig.Net.SASL.Handshake = config.SASLHandshake
	sconfig.Net.DialTimeout = config.DialTimeout
	sconfig.Version = sarama.V0_10_2_1
	sconfig.Consumer.Return.Errors = config.EnableReturnError

	if config.ConsumeOldest {
		sconfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	if "" != config.ClientID {
		sconfig.ClientID = config.ClientID
	}

	return sconfig
}

func (c *Consumer) Subscribe(handler ConsumerHandler) {
	c.subscribers = append(c.subscribers, handler)
}

func (c *Consumer) Start() {
	if len(c.subscribers) == 0 {
		panic(fmt.Sprintf("start consumer fail, must assigned at least one handler"))
	}

	go func() {
		for {
			if err := c.consumer.Consume(context.Background(), c.config.Topic, c); err != nil {
				log.Errorf("start consume kafka fail", log.String("error", err.Error()))
			}
		}
	}()
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		for _, fn := range c.subscribers {
			now := time.Now()
			err := fn(context.Background(), message)

			var errmsg string
			if err != nil {
				errmsg = err.Error()
			}
			duration := time.Since(now).Seconds()
			fields := make([]log.Field, 0, 7)
			fields = append(
				fields,
				log.Any("topic", c.config.Topic),
				log.Bytes("key", message.Key),
				log.Bytes("content", message.Value),
				log.Int32("partition", message.Partition),
				log.Int64("offset", message.Offset),
				log.Float64("duration", duration),
				log.String("error", errmsg),
			)

			if err != nil {
				log.Errorf("kafka consume fail", fields...)
				for _, topic := range c.config.Topic {
					metric.KafkaClientHandleCounter.Inc("unknown", "kafka", topic, "consume", "fail")
					metric.KafkaClientReqDuration.Observe(duration, "unknown", "kafka", topic, "consume")
				}
				continue
			} else {
				session.MarkMessage(message, "")
				log.Infof("kafka consume success", fields...)
				for _, topic := range c.config.Topic {
					metric.KafkaClientHandleCounter.Inc(strconv.Itoa(int(message.Partition)), "kafka", topic, "consume", "success")
					metric.KafkaClientReqDuration.Observe(duration, strconv.Itoa(int(message.Partition)), "kafka", topic, "consume")
				}
			}
		}
	}

	return nil
}

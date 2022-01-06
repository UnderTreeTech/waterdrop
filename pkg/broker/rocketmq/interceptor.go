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
	"strconv"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// producerMetricInterceptor producer metric
func producerMetricInterceptor(pc *ProducerConfig) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		now := time.Now()
		realReq := req.(*primitive.Message)
		realReply := reply.(*primitive.SendResult)

		err := next(ctx, realReq, realReply)

		var errmsg string
		if err != nil {
			errmsg = err.Error()
		}
		duration := time.Since(now).Seconds()

		fields := make([]log.Field, 0, 5)
		fields = append(
			fields,
			log.String("topic", pc.Topic),
			log.String("content", realReq.String()),
			log.String("response", realReply.String()),
			log.Float64("duration", duration),
			log.String("error", errmsg),
		)

		if err != nil {
			log.Error(ctx, "send msg fail", fields...)
			metric.RocketMQClientHandleCounter.Inc("unknown", "rocketmq", pc.Topic, "produce", err.Error())
			metric.RocketMQClientReqDuration.Observe(duration, "unknown", "rocketmq", pc.Topic, "produce")
		} else {
			log.Info(ctx, "send msg success", fields...)
			metric.RocketMQClientHandleCounter.Inc(realReply.MessageQueue.BrokerName, "rocketmq", pc.Topic, "produce", strconv.Itoa(int(realReply.Status)))
			metric.RocketMQClientReqDuration.Observe(duration, realReply.MessageQueue.BrokerName, "rocketmq", pc.Topic, "produce")
		}

		if pc.SlowSendDuration > 0 && time.Since(now) > pc.SlowSendDuration {
			log.Warn(ctx, "send slow", fields...)
		}
		return err
	}
}

// pushConsumerMetricInterceptor push consumer metric
func pushConsumerMetricInterceptor(pc *ConsumerConfig) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		now := time.Now()
		msgs := req.([]*primitive.MessageExt)

		err := next(ctx, msgs, reply)

		var errmsg string
		if err != nil {
			errmsg = err.Error()
		}
		holder := reply.(*consumer.ConsumeResultHolder)
		replyCode := strconv.Itoa(int(holder.ConsumeResult))
		duration := time.Since(now).Seconds()

		for _, msg := range msgs {
			metric.RocketMQClientHandleCounter.Inc(msg.StoreHost, "rocketmq", pc.Topic, "consume", replyCode)
			metric.RocketMQClientReqDuration.Observe(duration, msg.StoreHost, "rocketmq", pc.Topic, "consume")

			fields := make([]log.Field, 0, 6)
			fields = append(
				fields,
				log.String("topic", pc.Topic),
				log.Any("tags", pc.Tags),
				log.String("content", msg.String()),
				log.Int("response", int(holder.ConsumeResult)),
				log.Float64("duration", duration),
				log.String("error", errmsg),
			)

			if err != nil {
				log.Error(ctx, "consume msg fail", fields...)
			} else {
				log.Info(ctx, "consume msg success", fields...)
			}
		}
		return err
	}
}

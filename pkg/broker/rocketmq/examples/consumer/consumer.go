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

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/broker/rocketmq"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
)

func main() {
	defer log.New(nil).Sync()

	config := &rocketmq.ConsumerConfig{
		Endpoint:  []string{"your_endpoint"},
		AccessKey: "your_access_key",
		SecretKey: "your_secret_key",
		Topic:     "your_topic",
		Gid:       "your_group_id",
		Tags:      []string{"go-rocketmq"},
	}

	consumer := rocketmq.NewPushConsumer(config)
	consumer.Subscribe(consumeMsg)

	consumer.Start()
	time.Sleep(time.Hour)
}

func consumeMsg(ctx context.Context, msg *rocketmq.MessageExt) error {
	fmt.Println("msg", msg.String())
	return nil
}

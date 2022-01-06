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
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/broker/rocketmq"
)

func main() {
	defer log.New(nil).Sync()

	config := &rocketmq.ProducerConfig{
		Endpoint:    []string{"your_endpoint"},
		AccessKey:   "your_access_key",
		SecretKey:   "your_secret_key",
		Retry:       1,
		SendTimeout: time.Second,
		Topic:       "your_topic",
	}

	p := rocketmq.NewProducer(config)
	p.Start()

	for i := 0; i < 100; i++ {
		p.SendSyncMsg(context.Background(), "Hello RocketMQ Go Client!"+xstring.RandomString(16))
	}

	p.Shutdown()
}

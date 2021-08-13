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

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/broker/kafka"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
)

func main() {
	defer log.New(nil).Sync()

	config := &kafka.ProducerConfig{
		Addr:           []string{"your_instance_addr"},
		Topic:          []string{"your_instance_topic"},
		EnableSASLAuth: true,
		SASLMechanism:  "your_sasl_mechanics", //PLAIN
		SASLUser:       "your_sasl_user",
		SASLPassword:   "your_sasl_password",
		SASLHandshake:  true,

		DialTimeout:         time.Second * 5,
		EnableReturnSuccess: true,
	}

	producer := kafka.NewSyncProducer(config)

	for i := 0; i < 100000; i++ {
		err := producer.SendSyncMsg(context.Background(), xstring.RandomString(16))
		if err != nil {
			fmt.Println("error", err.Error())
		}
	}

	producer.Close()
}

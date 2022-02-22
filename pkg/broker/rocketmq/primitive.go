/*
 *
 * Copyright 2022 waterdrop authors.
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
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
)

type (
	// MessageExt is an alias of primitive.MessageExt
	MessageExt = primitive.MessageExt
	// Message is an alias of primitive.Message
	Message = primitive.Message
	// Interceptor is an alias of primitive.Interceptor
	Interceptor = primitive.Interceptor
)

// SetLogLevel set rocket mq log level
func SetLogLevel(level string) {
	if level == "" {
		return
	}
	rlog.SetLogLevel(level)
}

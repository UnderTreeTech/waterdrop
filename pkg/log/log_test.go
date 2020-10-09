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

package log

import (
	"context"
	"testing"
)

func TestLog(t *testing.T) {
	defaultLogger = newLogger(defaultConfig())

	Info(context.Background(), "info",
		Int64("age", 10),
		String("hello", "world"),
		Any("any", []string{"shanghai", "xuhui"}),
	)

	Warn(context.Background(), "warn",
		String("john", "sun"),
	)

	Debug(context.Background(), "debug",
		String("shanghai", "xuhui"),
	)

	Error(context.Background(), "division zero") //KVString("shanghai", "xuhui"),

	//Panic(context.Background(), "memory leaky", String("stop", "yes"))
}

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

import "go.uber.org/zap"

type Field = zap.Field

var (
	String   = zap.String
	Bytes    = zap.ByteString
	Duration = zap.Duration

	Int8  = zap.Int8
	Int32 = zap.Int32
	Int   = zap.Int
	Int64 = zap.Int64

	Uint8  = zap.Uint8
	Uint32 = zap.Uint32
	Uint   = zap.Uint
	Uint64 = zap.Uint64

	Float64 = zap.Float64

	Any = zap.Any
)

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

package metadata

import (
	"net/http"
	"strconv"
	"time"
)

const (
	HeaderContentType    = "Content-Type"
	HeaderUserAgent      = "User-Agent"
	HeaderAppkey         = "Appkey"
	HeaderTimestamp      = "Timestamp"
	HeaderSign           = "Sign"
	HeaderNonce          = "Nonce"
	HeaderAcceptLanguage = "Accept-Language"
	HeaderHttpTimeout    = "X-Request-Timeout"
	HeaderHttpTraceId    = "X-Trace-Id"

	DefaultContentTypeJson = "application/json;charset=utf-8"
	DefaultUserAgentVal    = "waterdrop"
	DefaultLocale          = "zh-CN"
	DefaultRequestTimeout  = 10
	DefaultNonceLen        = 16
	DefaultSecretURL       = "/api/app/secret"
	DefaultSkipsURL        = "/api/app/skips"
	DefaultAppkeyLen       = 16

	DefaultMaxBytes = 1 << 20 // 1 MiB
)

// get timeout from request header
// similar as grpc
func GetTimeout(req *http.Request) time.Duration {
	to := req.Header.Get(HeaderHttpTimeout)
	timeout, err := strconv.ParseInt(to, 10, 64)
	//reduce 5ms network transmission time for every request
	if err == nil && timeout > 5 {
		timeout -= 5
	}

	return time.Duration(timeout) * time.Millisecond
}

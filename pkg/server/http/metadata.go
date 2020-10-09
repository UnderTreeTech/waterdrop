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

package http

import (
	"net/http"
	"strconv"
	"time"
)

const (
	_contentType        = "Content-Type"
	_json               = "application/json;charset=utf-8"
	_userAgent          = "User-Agent"
	_waterdropUserAgent = "waterdrop"
	_appkey             = "Appkey"

	_timestamp      = "Timestamp"
	_sign           = "Sign"
	_nonce          = "Nonce"
	_acceptLanguage = "Accept-Language"
	_locale         = "zh-CN"

	_httpHeaderTimeout = "X-Request-Timeout"
	_httpHeaderTraceId = "X-Trace-Id"

	_requestTimeout = 10
	_nonceLen       = 16
	_secretURL      = "/api/app/secret"
	_skipsURL       = "/api/app/skips"
	_appkeyLen      = 16

	_maxBytes = 1 << 20 // 1 MiB
)

// get timeout from request header
// similar as grpc
func getTimeout(req *http.Request) time.Duration {
	to := req.Header.Get(_httpHeaderTimeout)
	timeout, err := strconv.ParseInt(to, 10, 64)
	//reduce 10ms network transmission time for every request
	if err == nil && timeout > 5 {
		timeout -= 5
	}

	return time.Duration(timeout) * time.Millisecond
}

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

package config

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ClientConfig http client config
type ClientConfig struct {
	// HostURL peer service host
	HostURL string
	// Timeout request timeout
	Timeout time.Duration
	// SlowRequestDuration slow request timeout
	SlowRequestDuration time.Duration
	// EnableDebug trace request details
	EnableDebug bool
	// Key client key
	Key string
	// Secret signature secret
	Secret string
}

// ServerConfig http server config
type ServerConfig struct {
	// Addr server addr, like :8080 or 127.0.0.1:8080
	Addr string
	// Timeout request timeout
	Timeout time.Duration
	// Mode server mode: release or debug
	Mode string
	// SlowRequestDuration slow request timeout
	SlowRequestDuration time.Duration
	// WatchConfig whether watch config file changes
	WatchConfig bool
}

// DefaultServerConfig default server configs, for start http server out of box
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Addr:                "0.0.0.0:10000",
		Mode:                gin.ReleaseMode,
		Timeout:             time.Millisecond * 1000,
		SlowRequestDuration: 500 * time.Millisecond,
	}
}

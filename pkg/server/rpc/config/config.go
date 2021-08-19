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
)

type ServerConfig struct {
	Addr string

	Timeout        time.Duration
	IdleTimeout    time.Duration
	MaxLifeTime    time.Duration
	ForceCloseWait time.Duration

	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration

	SlowRequestDuration time.Duration
	WatchConfig         bool
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Addr:                "0.0.0.0:20812",
		Timeout:             time.Second,
		IdleTimeout:         180 * time.Second,
		MaxLifeTime:         2 * time.Hour,
		ForceCloseWait:      20 * time.Second,
		KeepAliveInterval:   60 * time.Second,
		KeepAliveTimeout:    20 * time.Second,
		SlowRequestDuration: 500 * time.Millisecond,
	}
}

type ClientConfig struct {
	DialTimeout time.Duration
	Block       bool
	Balancer    string
	Target      string

	Timeout time.Duration

	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration

	SlowRequestDuration time.Duration
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		DialTimeout: 5 * time.Second,
		Block:       true,
		Balancer:    "round_robin",

		Timeout: 500 * time.Millisecond,

		KeepAliveInterval: 60 * time.Second,
		KeepAliveTimeout:  20 * time.Second,
	}
}

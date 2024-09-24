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

// ServerConfig rpc server config
type ServerConfig struct {
	// Addr server addr,it may be ":8080" or "127.0.0.1:8080"
	Addr string
	// Timeout rpc request timeout
	Timeout time.Duration
	// GRPC ServerParameters
	IdleTimeout       time.Duration
	MaxLifeTime       time.Duration
	ForceCloseWait    time.Duration
	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration
	// SlowRequestDuration slow rpc request timeout
	SlowRequestDuration time.Duration
	// WatchConfig whether watch config file changes
	WatchConfig bool
	// NotLog escape log detail path
	NotLog []string
}

// DefaultServerConfig default server config for starting rpc server out of box
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

// ClientConfig rpc client configs
type ClientConfig struct {
	// DialTimeout dial rpc server timeout
	DialTimeout time.Duration
	// Block dial mode: sync or async
	Block bool
	// Balancer client balancer, default round robbin
	Balancer string
	// Target rpc server endpoint
	Target string
	// Timeout rpc request timeout
	Timeout time.Duration
	// GRPC ClientParameters
	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration
	// SlowRequestDuration client slow request timeout
	SlowRequestDuration time.Duration
	// NotLog escape log detail path
	NotLog []string
	// MaxCallSendMsgSize default 4*1024*1024
	MaxCallSendMsgSize int
}

// DefaultClientConfig default client config for starting rpc client out of box
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

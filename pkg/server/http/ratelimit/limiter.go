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

package ratelimit

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Limiter interface {
	Allow(context.Context, string) bool
}

type Option func(*LimitOpts)

type LimitOpts struct {
	Strategy func(c *gin.Context) string
	Fallback func(c *gin.Context)
}

// Apply applies the option to limiter
func Apply(opts []Option) *LimitOpts {
	limitOpts := &LimitOpts{}
	for _, opt := range opts {
		opt(limitOpts)
	}

	return limitOpts
}

// WithFallback set the fallback handler when request is blocked
func WithFallback(fallback func(c *gin.Context)) Option {
	return func(opts *LimitOpts) {
		opts.Fallback = fallback
	}
}

// WithResourceStrategy set the verify path extract strategy of request
func WithResourceStrategy(strategy func(c *gin.Context) string) Option {
	return func(opts *LimitOpts) {
		opts.Strategy = strategy
	}
}

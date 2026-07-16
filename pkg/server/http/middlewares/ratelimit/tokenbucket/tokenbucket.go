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

// Package tokenbucket implements a lightweight in-process rate-limit middleware
// backed by golang.org/x/time/rate. It supports route-level limits: a request
// whose route has an explicit rule is limited per-route; every other route
// falls back to a single shared global limit. This is the local, dependency-free
// counterpart of the rule/cluster-based sentinel limiter.
package tokenbucket

import (
	"context"
	"net/http"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/middlewares/ratelimit"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Rate describes a token-bucket limit. Set Rate to rate.Limit(rps) for N tokens
// per second, or rate.Every(d) for one token every d. A zero-value Rate is
// treated as unlimited, so callers may leave Global unset to limit only the
// routes listed in Config.Routes.
type Rate struct {
	// Rate is the refill rate in tokens per second (rate.Inf = unlimited).
	Rate rate.Limit
	// Burst is the maximum bucket size, i.e. the number of requests allowed in
	// a single burst before the refill rate kicks in.
	Burst int
}

// Config for the token-bucket limiter.
type Config struct {
	// Global is applied to any route not listed in Routes. A zero-value Rate
	// means unlimited (no global limit).
	Global Rate
	// Routes maps a route key to a per-route limit. The key is the value
	// computed by the resource strategy, which defaults to METHOD:FullPath
	// (e.g. "GET:/api/app/secrets"). Routes take precedence over Global.
	Routes map[string]Rate
}

// Limiter is a route-aware token-bucket limiter. It implements ratelimit.Limiter.
type Limiter struct {
	global *rate.Limiter
	routes map[string]*rate.Limiter
}

// NewLimiter builds a Limiter from cfg. A nil cfg yields an unlimited limiter.
func NewLimiter(cfg *Config) *Limiter {
	if cfg == nil {
		return &Limiter{global: rate.NewLimiter(rate.Inf, 0)}
	}

	routes := make(map[string]*rate.Limiter, len(cfg.Routes))
	for path, r := range cfg.Routes {
		routes[path] = newLimiter(r)
	}

	return &Limiter{
		global: newLimiter(cfg.Global),
		routes: routes,
	}
}

// newLimiter converts a Rate into a *rate.Limiter. A zero-value Rate (no rate
// and no burst) is treated as unlimited to avoid the footgun where an unset
// Global would reject all traffic.
func newLimiter(r Rate) *rate.Limiter {
	if r.Rate <= 0 && r.Burst <= 0 {
		return rate.NewLimiter(rate.Inf, 0)
	}
	return rate.NewLimiter(r.Rate, r.Burst)
}

// Allow reports whether path may proceed. A route-specific limit wins over the
// global limit; an unknown route falls back to the global limit. The check is
// non-blocking: a request that cannot acquire a token is rejected immediately
// rather than queued.
func (l *Limiter) Allow(ctx context.Context, path string) bool {
	if limiter, ok := l.routes[path]; ok {
		return limiter.Allow()
	}
	return l.global.Allow()
}

// New returns a gin rate-limit middleware. The limited path defaults to
// METHOD:FullPath (e.g. "GET:/api/app/secrets") and may be overridden with
// ratelimit.WithResourceStrategy. When a request is rejected the middleware
// calls ratelimit.WithFallback if set, otherwise responds with 429.
func New(cfg *Config, opts ...ratelimit.Option) gin.HandlerFunc {
	limitOpts := ratelimit.Apply(opts)
	limiter := NewLimiter(cfg)

	return func(c *gin.Context) {
		path := c.Request.Method + ":" + c.FullPath()
		if limitOpts.Strategy != nil {
			path = limitOpts.Strategy(c)
		}

		if !limiter.Allow(c.Request.Context(), path) {
			if limitOpts.Fallback != nil {
				limitOpts.Fallback(c)
				// Abort so the blocked request does not fall through to the
				// route handler. AbortWithStatus already aborts in the else
				// branch; an explicit Abort is needed here because the fallback
				// may only have written a response without aborting.
				c.Abort()
			} else {
				c.AbortWithStatus(http.StatusTooManyRequests)
			}
			return
		}

		c.Next()
	}
}

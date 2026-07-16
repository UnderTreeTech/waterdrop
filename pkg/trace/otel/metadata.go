/*
 *
 * Copyright 2026 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/LICENSE.org/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package otel

import (
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

// MDCarrier adapts a gRPC metadata.MD to OTel's propagation.TextMapCarrier
// (Get/Set). gRPC metadata keys are lower-cased on Set to match gRPC's
// case-insensitive header semantics. It is used by the OTel gRPC interceptors
// so the composite propagator can inject/extract `traceparent` + `uber-trace-id`
// over gRPC metadata, mirroring the HTTP path.
type MDCarrier struct {
	MD metadata.MD
}

// Get returns the first value for the given key, or "" if none.
func (c MDCarrier) Get(key string) string {
	vs := c.MD.Get(key)
	if len(vs) == 0 {
		// metadata.Get is case-insensitive, but be defensive for any direct use.
		vs = c.MD[strings.ToLower(key)]
	}
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

// Set stores the value under the lower-cased key, appending to any existing
// values (propagation keys are single-valued in practice).
func (c MDCarrier) Set(key, val string) {
	k := strings.ToLower(key)
	c.MD[k] = append(c.MD[k], val)
}

// Keys returns the metadata keys (for debugging / propagation.Fields use).
func (c MDCarrier) Keys() []string {
	out := make([]string, 0, len(c.MD))
	for k := range c.MD {
		out = append(out, k)
	}
	return out
}

// compile-time interface check
var _ propagation.TextMapCarrier = MDCarrier{}

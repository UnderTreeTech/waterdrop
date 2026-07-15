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

// Package jaeger is the backward-compatible trace initialization entry point.
//
// The trace factory now lives in pkg/trace (see trace.Init). This package
// keeps the legacy `defer jaeger.Init()()` call site working: Init delegates
// to trace.Init, which reads [trace.jaeger] and [trace.otel] and selects the
// jaeger / otel / dual backend. The jaeger tracer construction (former Config,
// buildJaeger, newJaegerClient, defaultJaegerConfig) has moved into
// pkg/trace/jaeger_tracer.go.
package jaeger

import "github.com/UnderTreeTech/waterdrop/pkg/trace"

// Init initializes tracing from config and returns a shutdown func. It is the
// single entry point services call as `defer jaeger.Init()()`. It delegates to
// the pkg/trace factory, which selects the backend based on the [trace.jaeger]
// and [trace.otel] config sections.
func Init() func() {
	return trace.Init()
}

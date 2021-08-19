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

package etcd

import (
	"context"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
)

// mutex distributed lock based on etcd
type mutex struct {
	s *concurrency.Session
	m *concurrency.Mutex
}

// NewMutex new lock
func (er *EtcdRegistry) NewMutex(key string, opts ...concurrency.SessionOption) (m *mutex, err error) {
	m = &mutex{}

	// default session ttl 60s
	m.s, err = concurrency.NewSession(er.client, opts...)
	if err != nil {
		return
	}
	m.m = concurrency.NewMutex(m.s, key)

	return
}

// Lock do lock op
func (m *mutex) Lock(ctx context.Context, timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return m.m.Lock(ctx)
}

// Unlock release locked resource
func (m *mutex) Unlock(ctx context.Context) (err error) {
	err = m.m.Unlock(ctx)
	if err != nil {
		return
	}

	return m.s.Close()
}

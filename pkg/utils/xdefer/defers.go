/*
 *
 * Copyright 2021 waterdrop authors.
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

package xdefer

import "sync"

// Defers wrap defer func
type Defers struct {
	callbacks []func() error
	m         sync.Mutex
}

// New return a Defers pointer
func New() *Defers {
	return &Defers{
		callbacks: make([]func() error, 0),
	}
}

// Add add a func to Defers
func (d *Defers) Add(fns ...func() error) {
	d.m.Lock()
	defer d.m.Unlock()

	d.callbacks = append(d.callbacks, fns...)
}

// Close close defer funcs in Defers
func (d *Defers) Close() {
	d.m.Lock()
	defer d.m.Unlock()

	for i := len(d.callbacks) - 1; i >= 0; i-- {
		d.callbacks[i]()
	}
}

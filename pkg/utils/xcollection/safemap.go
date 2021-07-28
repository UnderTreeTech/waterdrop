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

package xcollection

import "sync"

// SafeMap is an implementation of sync.Map using a sync.RWMutex
type SafeMap struct {
	mu sync.RWMutex
	sm map[interface{}]interface{}
}

// NewSafeMap returns a SafeMap
func NewSafeMap() *SafeMap {
	return &SafeMap{
		sm: make(map[interface{}]interface{}),
	}
}

// Load returns the value stored in the map for a key, or nil if no
// value is present. The ok result indicates whether value was found
// in the map
func (m *SafeMap) Load(key interface{}) (value interface{}, ok bool) {
	m.mu.RLock()
	value, ok = m.sm[key]
	m.mu.RUnlock()
	return
}

// Store sets the value for a key
func (m *SafeMap) Store(key, value interface{}) {
	m.mu.Lock()
	if m.sm == nil {
		m.sm = make(map[interface{}]interface{})
	}
	m.sm[key] = value
	m.mu.Unlock()
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value. The loaded
// result is true if the value was loaded, false if stored
func (m *SafeMap) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	m.mu.Lock()
	actual, loaded = m.sm[key]
	if !loaded {
		actual = value
		if m.sm == nil {
			m.sm = make(map[interface{}]interface{})
		}
		m.sm[key] = value
	}
	m.mu.Unlock()
	return actual, loaded
}

// LoadAndDelete deletes the value for a key, returning the previous
// value if any. The loaded result reports whether the key was
// present
func (m *SafeMap) LoadAndDelete(key interface{}) (value interface{}, loaded bool) {
	m.mu.Lock()
	value, loaded = m.sm[key]
	if !loaded {
		m.mu.Unlock()
		return nil, false
	}
	delete(m.sm, key)
	m.mu.Unlock()
	return value, loaded
}

// Delete deletes the value for a key
func (m *SafeMap) Delete(key interface{}) {
	m.mu.Lock()
	delete(m.sm, key)
	m.mu.Unlock()
}

// Range calls f sequentially for each key and value present in the
// map. If f returns false, range stops the iteration
// Range does not necessarily correspond to any consistent snapshot
// of the Map's contents: no key will be visited more than once, but
// if the value for any key is stored or deleted concurrently, Range
// may reflect any mapping for that key from any point during the
// Range call
// Range may be O(N) with the number of elements in the map even if f
// returns false after a constant number of calls
func (m *SafeMap) Range(f func(key, value interface{}) bool) {
	m.mu.RLock()
	keys := make([]interface{}, 0, len(m.sm))
	for k := range m.sm {
		keys = append(keys, k)
	}
	m.mu.RUnlock()

	for _, k := range keys {
		v, ok := m.Load(k)
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

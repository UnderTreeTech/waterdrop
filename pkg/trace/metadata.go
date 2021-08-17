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

package trace

import "strings"

type CarrierMD struct {
	md map[string][]string
}

// Set a key:value pair to the carrier. Multiple calls to Set() for the
// same key leads to undefined behavior.
func (cm CarrierMD) Set(key, val string) {
	key = strings.ToLower(key)
	cm.md[key] = append(cm.md[key], val)
}

// ForeachKey returns TextMap contents via repeated calls to the `handler`
// function. If any call to `handler` returns a non-nil error, ForeachKey
// terminates and returns that error.
func (cm CarrierMD) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range cm.md {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

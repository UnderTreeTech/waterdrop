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

type Metadata struct {
	md map[string][]string
}

func New() *Metadata {
	return &Metadata{
		md: make(map[string][]string),
	}
}

func (md *Metadata) Set(key, val string) {
	key = strings.ToLower(key)
	md.md[key] = append(md.md[key], val)
}

func (md *Metadata) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range md.md {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

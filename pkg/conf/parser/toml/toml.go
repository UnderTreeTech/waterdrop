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

package toml

import "github.com/BurntSushi/toml"

type TOML map[string]interface{}

func NewTOMLParser() TOML {
	parser := make(TOML)
	return parser
}

// Marshal marshal TOML to bytes
func (t TOML) Marshal(m map[string]interface{}) ([]byte, error) {
	return nil, nil
}

// Unmarshal unmarshal input bytes to map[string]interface{}
func (t TOML) Unmarshal(b []byte) (map[string]interface{}, error) {
	if err := toml.Unmarshal(b, &t); err != nil {
		return nil, err
	}

	return t, nil
}

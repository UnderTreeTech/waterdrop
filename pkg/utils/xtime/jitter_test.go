/*
 *
 * Copyright 2023 waterdrop authors.
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

package xtime

import (
	"testing"
	"time"
)

func TestJitterTime(t *testing.T) {
	t.Run("sample test", func(t *testing.T) {
		const (
			d       = 10 * time.Second
			f       = 0.5
			samples = 20
		)

		for i := 0; i < samples; i++ {
			r := JitterTime(d, f)
			t.Log(r)
			if r < 5*time.Second || r > 15*time.Second {
				t.Error("sample outside of range: ", r)
			}
		}

		for i := 0; i < samples; i++ {
			r := JitterTime(d)
			t.Log(r)
			if r < 9*time.Second || r > 11*time.Second {
				t.Error("sample outside of range: ", r)
			}
		}
	})
}

func TestJitter(t *testing.T) {
	t.Run("sample test", func(t *testing.T) {
		const (
			d       = 10
			f       = 0.5
			samples = 20
		)

		for i := 0; i < samples; i++ {
			r := Jitter(d, f)
			t.Log(r)
			if r < 5 || r > 15 {
				t.Error("sample outside of range: ", r)
			}
		}

		for i := 0; i < samples; i++ {
			r := Jitter(d)
			t.Log(r)
			if r < 9 || r > 11 {
				t.Error("sample outside of range: ", r)
			}
		}
	})
}

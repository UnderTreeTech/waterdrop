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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDefers test xdefer
func TestDefers(t *testing.T) {
	defers := New()
	exitLog := "exit"
	fn1 := func() error {
		exitLog += " fn1"
		return nil
	}

	fn2 := func() error {
		exitLog += " fn2"
		return nil
	}

	fnN := func() error {
		exitLog += " fnN"
		return nil
	}

	defers.Add(fn1, fn2, fnN)
	defers.Close()
	assert.Equal(t, exitLog, "exit fnN fn2 fn1")
}

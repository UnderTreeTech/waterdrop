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

package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	locked, lockVal, err := r.Lock(ctx, "lock", 10000)
	assert.Nil(t, err)
	assert.True(t, locked)
	locked, _, err = r.Lock(ctx, "lock", 10000)
	assert.Nil(t, err)
	assert.False(t, locked)
	err = r.UnLock(ctx, "lock", lockVal)
	assert.Nil(t, err)
	err = r.ForceUnLock(ctx, "lock")
	assert.Nil(t, err)
}

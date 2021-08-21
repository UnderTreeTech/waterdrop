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

package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	etcd := New(defaultConfig)
	defer etcd.Close()

	lock, err := etcd.NewMutex("/waterdrop/test/lock")
	assert.Nil(t, err)
	err = lock.Lock(context.Background(), time.Second)
	assert.Nil(t, err)
	defer lock.Unlock(context.Background())

	lock2, err := etcd.NewMutex("/waterdrop/test/lock")
	defer lock2.Unlock(context.Background())
	assert.Nil(t, err)
	err = lock2.Lock(context.Background(), time.Second)
	assert.NotNil(t, err)
}

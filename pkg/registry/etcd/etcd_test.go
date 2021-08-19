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

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"

	"github.com/UnderTreeTech/waterdrop/pkg/registry"
)

// TestEtcdRegistry test etcd registry
func TestEtcdRegistry(t *testing.T) {
	defer log.New(nil).Sync()
	etcd := New(defaultConfig)
	defer etcd.Close()
	service := &registry.ServiceInfo{
		Name:    "service.waterdrop.v1",
		Scheme:  schemeGRPC,
		Addr:    "127.0.0.1:9999",
		Version: "v1.2",
	}

	err := etcd.Register(context.Background(), service)
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 10)
	services, err := etcd.List(context.Background(), service.Name, "")
	assert.Nil(t, err)
	assert.Equal(t, len(services), 1)
	err = etcd.DeRegister(context.Background(), service)
	assert.Nil(t, err)
	services, err = etcd.List(context.Background(), service.Name, "")
	assert.Nil(t, err)
	assert.Equal(t, len(services), 0)
}

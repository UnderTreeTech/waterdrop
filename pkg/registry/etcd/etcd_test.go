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
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/resolver"

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

// mockClientConn records the most recent UpdateState for resolver tests.
type mockClientConn struct {
	resolver.ClientConn // embed for unimplemented methods
	mu       sync.Mutex
	state    resolver.State
	updated  chan struct{}
	errCount int
}

func newMockClientConn() *mockClientConn {
	return &mockClientConn{updated: make(chan struct{}, 16)}
}

func (m *mockClientConn) UpdateState(s resolver.State) error {
	m.mu.Lock()
	m.state = s
	m.mu.Unlock()
	select {
	case m.updated <- struct{}{}:
	default:
	}
	return nil
}

func (m *mockClientConn) ReportError(error) {
	m.mu.Lock()
	m.errCount++
	m.mu.Unlock()
}

func (m *mockClientConn) lastState() resolver.State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// TestEtcdResolver exercises the resolver that Build returns: it must observe
// service registrations, and closing it must not break the shared etcd client.
// This guards the grpc 1.80 fix where Resolver.Close() previously shut down the
// whole etcd client (taking server-side KeepAlive and NewMutex down with it).
func TestEtcdResolver(t *testing.T) {
	defer log.New(nil).Sync()
	etcd := New(defaultConfig)
	defer etcd.Close()

	const svcName = "resolver.waterdrop.v1"
	tgt := resolver.Target{URL: url.URL{Scheme: "etcd", Path: "/" + svcName}}
	cc := newMockClientConn()

	rs, err := etcd.Build(tgt, cc, resolver.BuildOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, rs)

	// Register one grpc endpoint; the resolver should push it via UpdateState.
	svc := &registry.ServiceInfo{
		Name:    svcName,
		Scheme:  schemeGRPC,
		Addr:    "grpc://127.0.0.1:9999",
		Version: "v1.0",
	}
	err = etcd.Register(context.Background(), svc)
	assert.Nil(t, err)

	select {
	case <-cc.updated:
	case <-time.After(3 * time.Second):
		t.Fatal("resolver did not push UpdateState after registration")
	}
	assert.Equal(t, 1, len(cc.lastState().Addresses))

	// Close the resolver (as grpc does on ClientConn.Close()). The shared etcd
	// client must remain usable: registrations and reads still work.
	rs.Close()
	err = etcd.DeRegister(context.Background(), svc)
	assert.Nil(t, err)

	services, err := etcd.List(context.Background(), svcName, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(services))
}

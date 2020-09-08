package etcd

import (
	"context"
	"time"

	"go.etcd.io/etcd/clientv3/concurrency"
)

type mutex struct {
	s *concurrency.Session
	m *concurrency.Mutex
}

func (er *EtcdRegistry) NewMutex(key string, opts ...concurrency.SessionOption) (m *mutex, err error) {
	m = &mutex{}

	// default session ttl 60s
	m.s, err = concurrency.NewSession(er.client, opts...)
	if err != nil {
		return
	}
	m.m = concurrency.NewMutex(m.s, key)

	return
}

func (m *mutex) Lock(ctx context.Context, timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return m.m.Lock(ctx)
}

func (m *mutex) TryLock(ctx context.Context, timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return m.m.TryLock(ctx)
}

func (m *mutex) Unlock(ctx context.Context) (err error) {
	err = m.m.Unlock(ctx)
	if err != nil {
		return
	}

	return m.s.Close()
}

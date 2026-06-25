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

package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"

	"google.golang.org/grpc/attributes"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/registry"

	"google.golang.org/grpc"
)

var (
	defaultPrefix = "waterdrop"
	defaultConfig = &Config{
		Prefix:      defaultPrefix,
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 10 * time.Second,
		RegisterTTL: 90 * time.Second,
	}
	schemeHTTP = "http"
	schemeGRPC = "grpc"
)

// Config etcd config
type Config struct {
	Prefix      string
	Endpoints   []string
	DialTimeout time.Duration
	RegisterTTL time.Duration
	Username    string
	Password    string
}

// EtcdRegistry etcd registry definition
type EtcdRegistry struct {
	client   *clientv3.Client
	services sync.Map
	cancels sync.Map
	config   *Config
}

// New new a etcd registry
func New(config *Config) *EtcdRegistry {
	if nil == config {
		config = defaultConfig
	}

	cliConfig := clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: config.DialTimeout,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
		Username: config.Username,
		Password: config.Password,
	}

	cli, err := clientv3.New(cliConfig)
	if err != nil {
		panic(fmt.Sprintf("new etcd client fail, err msg %s", err.Error()))
	}

	return &EtcdRegistry{
		client: cli,
		config: config,
	}
}

// Register register a service metadata to etcd
func (e *EtcdRegistry) Register(ctx context.Context, info *registry.ServiceInfo) error {
	key := e.serviceKey(info)
	val, _ := json.Marshal(info)

	grant, err := e.client.Grant(ctx, int64(e.config.RegisterTTL/1e9))
	if err != nil {
		log.Errorf("etcd grant fail", log.Any("grant", grant), log.Any("service", info), log.String("error", err.Error()))
		return err
	}

	resp, err := e.client.Put(ctx, key, string(val), clientv3.WithLease(grant.ID))
	if err != nil {
		log.Errorf("etcd put fail", log.Any("put_resp", resp), log.Any("service", info), log.String("error", err.Error()))
		return err
	}

	keepAliveCtx, cancel := context.WithCancel(context.Background())
	keepAliveCh, err := e.client.KeepAlive(keepAliveCtx, grant.ID)
	if err != nil {
		cancel()
		log.Errorf("etcd keepalive fail", log.Any("service", info), log.String("error", err.Error()))
		return err
	}

	e.services.Store(key, val)
	e.cancels.Store(key, cancel)

	go func() {
		for {
			select {
			case <-keepAliveCtx.Done():
				return
			case _, ok := <-keepAliveCh:
				if !ok {
					log.Warnf("etcd keepalive channel closed, try to re-register", log.Any("service", info))
					time.Sleep(time.Second)

					// try to re-register
					for {
						select {
						case <-keepAliveCtx.Done():
							return
						default:
						}

						grant, err := e.client.Grant(keepAliveCtx, int64(e.config.RegisterTTL/1e9))
						if err != nil {
							log.Errorf("etcd grant fail during retry", log.String("error", err.Error()))
							time.Sleep(time.Second)
							continue
						}

						_, err = e.client.Put(keepAliveCtx, key, string(val), clientv3.WithLease(grant.ID))
						if err != nil {
							log.Errorf("etcd put fail during retry", log.String("error", err.Error()))
							time.Sleep(time.Second)
							continue
						}

						keepAliveCh, err = e.client.KeepAlive(keepAliveCtx, grant.ID)
						if err != nil {
							log.Errorf("etcd keepalive fail during retry", log.String("error", err.Error()))
							time.Sleep(time.Second)
							continue
						}

						log.Infof("etcd re-register success", log.Any("service", info))
						break
					}
				}
			}
		}
	}()

	return nil
}

// DeRegister remove a service metadata from etcd
func (e *EtcdRegistry) DeRegister(ctx context.Context, info *registry.ServiceInfo) error {
	key := e.serviceKey(info)
	return e.deRegister(ctx, key)
}

// deRegister remove service
func (e *EtcdRegistry) deRegister(ctx context.Context, key string) error {
	if cancel, ok := e.cancels.Load(key); ok {
		cancel.(context.CancelFunc)()
		e.cancels.Delete(key)
	}

	if resp, err := e.client.Delete(ctx, key); err != nil {
		log.Errorf("etcd delete fail", log.Any("del_resp", resp), log.String("service", key), log.String("error", err.Error()))
		return err
	}

	e.services.Delete(key)
	log.Infof("deregister service", log.String("service", key))
	return nil
}

// List list services from etcd with name and scheme
func (e *EtcdRegistry) List(ctx context.Context, name string, scheme string) (services []*registry.ServiceInfo, err error) {
	target := fmt.Sprintf("/%s/%s/%s", e.config.Prefix, name, scheme)
	resp, err := e.client.Get(ctx, target, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("etcd list fail", log.String("name", name), log.String("scheme", scheme), log.String("error", err.Error()))
		return
	}

	for _, kv := range resp.Kvs {
		service := &registry.ServiceInfo{}
		if err = json.Unmarshal(kv.Value, service); err != nil {
			log.Warnf("unmarshal response fail", log.Bytes("reply", kv.Value), log.String("error", err.Error()))
			continue
		}
		services = append(services, service)
	}
	return
}

// Close close connection to etcd and deRegister all service info
func (e *EtcdRegistry) Close() {
	var wg sync.WaitGroup

	e.services.Range(func(k, v interface{}) bool {
		wg.Add(1)
		go func(k interface{}) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			e.deRegister(ctx, k.(string))
			cancel()
		}(k)
		return true
	})
	wg.Wait()

	e.client.Close()
}

// Build watch service changes.
// Resolver Segment
func (e *EtcdRegistry) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	watchKey := fmt.Sprintf("/%s/%s/", e.config.Prefix, target.Endpoint)

	ctx, cancel := context.WithCancel(context.Background())
	r := &etcdResolver{
		e:        e,
		cc:       cc,
		watchKey: watchKey,
		ctx:      ctx,
		cancel:   cancel,
	}
	go r.watch()
	return r, nil
}

// Scheme return etcd's scheme
func (e *EtcdRegistry) Scheme() string {
	return "etcd"
}

// serviceKey service key format in etcd
func (e *EtcdRegistry) serviceKey(info *registry.ServiceInfo) string {
	return fmt.Sprintf("/%s/%s/%s", e.config.Prefix, info.Name, info.Addr)
}

// etcdResolver is a per-ClientConn resolver. It owns its own cancellable context
// and watch goroutine, so Close() only stops this resolver and never touches the
// shared etcd client (which is also used by server-side KeepAlive and NewMutex).
type etcdResolver struct {
	e        *EtcdRegistry
	cc       resolver.ClientConn
	watchKey string
	ctx      context.Context
	cancel   context.CancelFunc
}

// ResolveNow is a noop for Resolver.
func (r *etcdResolver) ResolveNow(_ resolver.ResolveNowOptions) {}

// Close stops the resolver. It cancels the watch context so the watch goroutine
// exits cleanly. It must NOT close the shared etcd client.
func (r *etcdResolver) Close() {
	r.cancel()
}

// watch etcd changes. Re-establishes the watch on channel close with a backoff.
func (r *etcdResolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		// Start the watch before the full Get so events between Get and Watch
		// are not lost.
		wch := r.e.client.Watch(r.ctx, r.watchKey, clientv3.WithPrefix())
		// Full snapshot on first connect and after every reconnect.
		r.updateAddrs()

		for event := range wch {
			for _, ev := range event.Events {
				if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
					r.updateAddrs()
				}
			}
		}

		// Watch channel closed. If we were cancelled, stop; otherwise backoff
		// and reconnect.
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(time.Second):
		}
		log.Warnf("etcd watch channel closed, retrying", log.String("watch_key", r.watchKey))
	}
}

// updateAddrs refreshes the service list from etcd and pushes it to the
// balancer. On Get failure it reports the error to the ClientConn so the
// balancer backs off instead of sticking to stale addresses.
func (r *etcdResolver) updateAddrs() {
	resp, err := r.e.client.Get(r.ctx, r.watchKey, clientv3.WithPrefix())
	if err != nil {
		if r.ctx.Err() == nil {
			r.cc.ReportError(err)
			log.Errorf("etcd get fail", log.String("watch_key", r.watchKey), log.String("error", err.Error()))
		}
		return
	}

	addrs := r.getAddrs(r.parse(resp))
	// Skip pushing an empty list so the balancer is not left with zero
	// backends on transient reads.
	if len(addrs) == 0 {
		log.Warnf("zero peer resolved, skip UpdateState", log.String("watch_key", r.watchKey))
		return
	}

	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		log.Errorf("update state fail", log.String("error", err.Error()))
	}
}

// parse etcd response to service info
func (r *etcdResolver) parse(resp *clientv3.GetResponse) (services []*registry.ServiceInfo) {
	services = make([]*registry.ServiceInfo, 0)
	for _, event := range resp.Kvs {
		service := &registry.ServiceInfo{}
		err := json.Unmarshal(event.Value, service)
		if err != nil {
			log.Errorf("unmarshal service fail", log.String("error", err.Error()))
			continue
		}
		services = append(services, service)
	}
	return services
}

// getAddrs get addrs from grpc resolver
func (r *etcdResolver) getAddrs(services []*registry.ServiceInfo) []resolver.Address {
	addrs := make([]resolver.Address, 0, len(services))
	for _, service := range services {
		if service.Scheme != schemeGRPC {
			continue
		}

		u, err := url.Parse(service.Addr)
		if err != nil {
			log.Errorf("parse service addr fail", log.Any("service", service), log.String("error", err.Error()))
			continue
		}

		addr := resolver.Address{
			Addr:       u.Host,
			ServerName: service.Name,
			Attributes: attributes.New("scheme", u.Scheme),
		}
		addrs = append(addrs, addr)
	}

	log.Debugf(fmt.Sprintf("resolver %d peer service", len(addrs)), log.Any("services", addrs))
	return addrs
}

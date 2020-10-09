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
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/attributes"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
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

type Config struct {
	Prefix      string
	Endpoints   []string
	DialTimeout time.Duration
	RegisterTTL time.Duration
	Username    string
	Password    string
}

type EtcdRegistry struct {
	client   *clientv3.Client
	services sync.Map
	config   *Config
}

func New(config *Config) *EtcdRegistry {
	if nil == config {
		config = defaultConfig
	}

	cliConfig := clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: config.DialTimeout,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
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

func (e *EtcdRegistry) Register(ctx context.Context, info *registry.ServiceInfo) error {
	err := e.register(ctx, info)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(e.config.RegisterTTL / 3)
		for {
			select {
			case <-ticker.C:
				e.register(ctx, info)
			}
		}
	}()

	return nil
}

func (e *EtcdRegistry) register(ctx context.Context, info *registry.ServiceInfo) error {
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

	e.services.Store(key, val)
	return nil
}

func (e *EtcdRegistry) DeRegister(ctx context.Context, info *registry.ServiceInfo) error {
	key := e.serviceKey(info)
	return e.deRegister(ctx, key)
}

func (e *EtcdRegistry) deRegister(ctx context.Context, key string) error {
	if resp, err := e.client.Delete(ctx, key); err != nil {
		log.Errorf("etcd delete fail", log.Any("del_resp", resp), log.String("service", key), log.String("error", err.Error()))
		return err
	}

	e.services.Delete(key)
	log.Infof("deregister service", log.String("service", key))
	return nil
}

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

// Resolver Segment
func (e *EtcdRegistry) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	go e.watch(cc, e.config.Prefix, target.Endpoint)
	return e, nil
}

// Scheme return etcd's scheme
func (e *EtcdRegistry) Scheme() string {
	return "etcd"
}

// ResolveNow is a noop for Resolver.
func (e *EtcdRegistry) ResolveNow(rn resolver.ResolveNowOptions) {
}

func (e *EtcdRegistry) serviceKey(info *registry.ServiceInfo) string {
	return fmt.Sprintf("/%s/%s/%s", e.config.Prefix, info.Name, info.Addr)
}

func (e *EtcdRegistry) watch(cc resolver.ClientConn, prefix string, serviceName string) {
	e.updateAddrs(cc, prefix, serviceName)
	watchKey := fmt.Sprintf("/%s/%s/", prefix, serviceName)

	respChan := e.client.Watch(context.Background(), watchKey, clientv3.WithPrefix())
	for event := range respChan {
		for _, ev := range event.Events {
			if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
				e.updateAddrs(cc, prefix, serviceName)
			}
		}
	}
}

func (e *EtcdRegistry) updateAddrs(cc resolver.ClientConn, prefix string, serviceName string) (err error) {
	watchKey := fmt.Sprintf("/%s/%s/", prefix, serviceName)
	peers, err := e.client.Get(context.Background(), watchKey, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("etcd get fail", log.String("watch_key", watchKey), log.String("error", err.Error()))
		return err
	}

	services := e.parse(peers)
	newAddrs := e.getAddrs(services)

	if len(newAddrs) > 0 {
		cc.UpdateState(resolver.State{Addresses: newAddrs})
	}

	return nil
}

func (e *EtcdRegistry) parse(resp *clientv3.GetResponse) (services []*registry.ServiceInfo) {
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

func (e *EtcdRegistry) getAddrs(services []*registry.ServiceInfo) []resolver.Address {
	addrs := make([]resolver.Address, 0, len(services))
	for _, service := range services {
		if service.Scheme != schemeGRPC {
			continue
		}

		var weight int64
		if weight, _ := strconv.ParseInt(service.Metadata[registry.MetaWeight], 10, 64); weight <= 0 {
			weight = 100
		}

		u, err := url.Parse(service.Addr)
		if err != nil {
			log.Errorf("parse service addr fail", log.Any("service", service), log.String("error", err.Error()))
			continue
		}

		addr := resolver.Address{
			Addr:       u.Host,
			ServerName: service.Name,
			Attributes: attributes.New("weight", weight, "scheme", u.Scheme),
		}
		addrs = append(addrs, addr)
	}

	log.Infof(fmt.Sprintf("resolver %d peer service", len(addrs)), log.Any("services", addrs))
	return addrs
}

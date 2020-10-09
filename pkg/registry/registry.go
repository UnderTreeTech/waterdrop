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

package registry

import "context"

// metadata common key
const (
	MetaWeight  = "weight"
	MetaCluster = "cluster"
	MetaZone    = "zone"
	MetaColor   = "color"
)

type Registry interface {
	Register(ctx context.Context, info *ServiceInfo) error
	DeRegister(ctx context.Context, info *ServiceInfo) error
	Close()
}

type ServiceInfo struct {
	// Service Name
	Name string `json:"name"`
	// Service Scheme, http/grpc
	Scheme string `json:"schema"`
	// Service Addr
	Addr string `json:"addr"`
	// Metadata is the information associated with Addr, which may be used
	// to make load balancing decision
	Metadata map[string]string `json:"metadata"`
	// Region is region
	Region string `json:"region"`
	// Zone is IDC
	Zone string `json:"zone"`
	// prod/pre/test/dev
	Env string `json:"env"`
	// Service Version
	Version string `json:"version"`
}

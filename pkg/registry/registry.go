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

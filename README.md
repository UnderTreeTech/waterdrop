[Chinese](README_CN.md)

## Waterdrop

Waterdrop is a high performance micro-service framework based on gin and grpc. Waterdrop comes from (The Three Body Problem).
![waterdrop](docs/images/waterdrop.jpg)

## Features

- HTTP Server: Based on [gin](https://github.com/gin-gonic/gin) and can reuse all its features
- RPC Server: Based on the official [gRPC-go](https://github.com/grpc/grpc-go) and use ETCD for service registration and discovery, 
default load balancing policy is roundrobin
- Conf: Support YAML, TOML, JSON and other extensions, default TOML parsing, user can decide whether or not to 
watch the config file changes for hot reload configuration
- Database: Integrated MySQL, Redis
- Log: Based on Zap encapsulation
- Trace: Integrate Opentracing access and use jaeger to record trace records
- Distribute Lock: distributed Lock is implemented based on Redis and ETCD. 
The former is suitable for final consistent business locks, while the latter is suitable for strongly consistent business locks
- Stats: Metrics & Profile for service operation
- Broker: Support RocketMQ and Kafka.
- Utils: Helper class function
- Registry: Service Registry discovery, etcd is the default service discovery component
- Status: Global error handling for error conversion between HTTP/RPC
- Dashboard: Build metrics dashboard based on Grafana, to be implemented
- Breaker: Plan to support [alibaba sentinel] (github.com/alibaba/sentinel-golang), 
[Google sre breaker] (https://landing.google.com/sre/sre-book/chapters/handling-overload/) and 
[netflix Hystrix] (https://github.com/afex/hystrix-go), to be implemented
- Middlewares & Interceptors: HTTP/RPC Server common middleware, such as token bucket/leaky bucket flow limiting, 
request signature, etc., to be implemented
- Cron: Timed task, based on [Cron](github.com/robfig/cron), to be implemented


## Installation

`go get github.com/UnderTreeTech/waterdrop`

## Documentation

## Contributing

Contributions are always welcomed! You can start with the issues labeled with bug or feature.

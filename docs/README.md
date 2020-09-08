## Waterdrop

Waterdrop(水滴) is a high performance micro-service framework based on gin and grpc. Waterdrop comes from (The Three Body Problem).


## Features

- HTTP Server：基于gin进行封装，可复用gin所有特性
- RPC Server：基于官方gRPC开发，基于etcd进行服务注册发现，默认roundrobin负载均衡
- Config：支持yaml、toml、json等多格式扩展，默认toml解析，可自定义是否监听文件热更新配置
- Database：集成MySQL、Redis
- Log：基于Zap封装
- Trace：集成Opentracing接入，jaeger落地支撑

## Installation

`go get github.com/UnderTreeTech/waterdrop`

## Documentation

## Contributing

Contributions are always welcomed! You can start with the issues labeled with bug or feature.


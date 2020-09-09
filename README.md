## Waterdrop

Waterdrop(水滴) is a high performance micro-service framework based on gin and grpc. Waterdrop comes from (The Three Body Problem).


## Features

- HTTP Server：基于gin进行封装，可复用gin所有特性
- RPC Server：基于官方gRPC开发，基于etcd进行服务注册发现，默认roundrobin负载均衡
- Conf：支持yaml、toml、json等多格式扩展，默认toml解析，可自定义是否监听文件热更新配置
- Database：集成MySQL、Redis
- Log：基于Zap封装
- Trace：集成Opentracing接入，jaeger落地支撑
- Distribute Lock：基于Redis、ETCD实现分布式锁，前者适合最终一致性业务锁，后者适合强一致性业务锁
- Stats：服务运行metrics & profile
- MQ：默认支持RocketMQ, Kafka. 实现中...
- Utils: 辅助类函数
- Registry：服务注册发现，制定通用接口定义，默认支持etcd
- Status：全局错误处理，用于HTTP/RPC之间错误转换。后续可扩展成从remote加载错误定义
- Dashboard：基于Grafana搭建metrics大盘，待实现
- Breaker：熔断器，计划支持google sre breaker及netflix hystrix，待实现
- Middlewares & Interceptors：http/rpc server通用中间件，如令牌桶/漏桶限流、signature签名等，待实现
- Cron：定时任务，基于[cron](github.com/robfig/cron)实现定时任务处理，待实现


## Installation

`go get github.com/UnderTreeTech/waterdrop`

## Documentation

## Contributing

Contributions are always welcomed! You can start with the issues labeled with bug or feature.


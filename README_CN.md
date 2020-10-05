[English](README.md)

## Waterdrop

水滴是基于gin及grpc构建的一款高性能微服务框架. 水滴之名来自于《三体》——纯洁而唯美，攻击高速有效，威力巨大！
![水滴](docs/images/waterdrop.jpg)

## Features

- HTTP Server：基于gin进行封装，可复用gin所有特性
- RPC Server：基于官方gRPC开发，基于etcd进行服务注册发现，默认roundrobin负载均衡
- Conf：支持yaml、toml、json等多格式扩展，默认toml解析，可自定义是否监听文件热更新配置
- Database：集成MySQL、Redis
- Log：基于Zap封装
- Trace：集成Opentracing接入，jaeger落地支撑
- Distribute Lock：基于Redis、ETCD实现分布式锁，前者适合最终一致性业务锁，后者适合强一致性业务锁
- Stats：服务运行metrics & profile
- Broker：默认支持RocketMQ, Kafka. 
- Utils: 辅助类函数
- Registry：服务注册发现，制定通用接口定义，默认支持etcd
- Status：全局错误处理，用于HTTP/RPC之间错误转换。后续可扩展成从remote加载错误定义
- Dashboard：基于Grafana搭建metrics大盘，待实现
- Breaker：熔断器，计划支持[alibaba sentinel](github.com/alibaba/sentinel-golang)、[google sre breaker](https://landing.google.com/sre/sre-book/chapters/handling-overload/) 及 [netflix hystrix](https://github.com/afex/hystrix-go) 
- Middlewares & Interceptors：http/rpc server通用中间件，如令牌桶/漏桶限流、signature签名等，待实现
- Cron：定时任务，基于[cron](github.com/robfig/cron)实现定时任务处理，待实现


## Installation

`go get github.com/UnderTreeTech/waterdrop`

## Documentation

## Contributing

Contributions are always welcomed! You can start with the issues labeled with bug or feature.


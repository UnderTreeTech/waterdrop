简体中文 | [English](README.md)

[![Go](https://github.com/UnderTreeTech/waterdrop/workflows/Go/badge.svg?branch=master)](https://github.com/UnderTreeTech/waterdrop/actions)
[![codecov](https://codecov.io/gh/UnderTreeTech/waterdrop/branch/master/graph/badge.svg)](https://codecov.io/gh/UnderTreeTech/waterdrop)
[![Go Report Card](https://goreportcard.com/badge/github.com/UnderTreeTech/waterdrop)](https://goreportcard.com/report/github.com/UnderTreeTech/waterdrop)
[![Release](https://img.shields.io/github/v/release/UnderTreeTech/waterdrop.svg?style=flat-square)](https://github.com/UnderTreeTech/waterdrop)
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

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
- Breaker：熔断器，支持[alibaba sentinel](github.com/alibaba/sentinel-golang)、[google sre breaker](https://landing.google.com/sre/sre-book/chapters/handling-overload/)
- Middlewares & Interceptors：http/rpc server通用中间件，如recovery, trace, metric and logger等


## Installation

`go get github.com/UnderTreeTech/waterdrop`

## Tools

waterdrop提供脚手架工具来提高开发效率。执行`go get -u github.com/UnderTreeTech/waterdrop/tools/waterdrop`安装最新版工具。

工具依赖`protc`及`protoc-gen-go`来生成protobuf代码，目前waterdrop工具并不自动安装这两个插件需要用户自主安装，实际开发中每人的版本并不相同。

waterdrop工具提供的功能如下：

- `watedrop new your_project_name` new a standard layout project
- `waterdrop protoc --grpc --swagger xx.proto` generate grpc code and swagger api file
- `waterdrop swagger serve xx.swagger.json` serve and browse swagger api
- `watedrop utgen xx.go` generate unit tests
- `watedrop upgrade` upgrade tool `watedrop`

## Contributing

Contributions are always welcomed! You can start with the issues labeled with bug or feature.


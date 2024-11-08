English | [简体中文](README_CN.md)

[![Go](https://github.com/UnderTreeTech/waterdrop/workflows/Go/badge.svg?branch=master)](https://github.com/UnderTreeTech/waterdrop/actions)
[![codecov](https://codecov.io/gh/UnderTreeTech/waterdrop/branch/master/graph/badge.svg)](https://codecov.io/gh/UnderTreeTech/waterdrop)
[![Go Report Card](https://goreportcard.com/badge/github.com/UnderTreeTech/waterdrop)](https://goreportcard.com/report/github.com/UnderTreeTech/waterdrop)
[![Release](https://img.shields.io/github/v/release/UnderTreeTech/waterdrop.svg?style=flat-square)](https://github.com/UnderTreeTech/waterdrop)
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

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
- Broker: Support RocketMQ and Kafka
- Utils: Helper class function
- Registry: Service Registry discovery, etcd is the default service discovery component
- Status: Global error handling for error conversion between HTTP/RPC
- Dashboard: Build metrics dashboard based on Grafana, to be implemented
- Breaker: Support [alibaba sentinel](https://github.com/alibaba/sentinel-golang), 
[google sre breaker](https://landing.google.com/sre/sre-book/chapters/handling-overload/)
- Middlewares & Interceptors: HTTP/RPC Server common middleware, such as recovery, trace, metric and logger,etc


## Installation

`go get github.com/UnderTreeTech/waterdrop`

## Tools

Execute the following command to get waterdrop tool to help you boost your development progress

`go get -u github.com/UnderTreeTech/waterdrop/tools/waterdrop`

You can use `waterdrop help` to find out how to use tools

You can generate protobuf codes but make sure you've already installed `protc` and `protoc-gen-go`. 
Here we don't install the two plugins automatically because we are not sure which version you will choose.

- `waterdrop new your_project_name` new a standard layout project
- `waterdrop protoc --grpc --swagger xx.proto` generate grpc code and swagger api file
- `waterdrop swagger serve xx.swagger.json` serve and browse swagger api
- `waterdrop utgen xx.go` generate unit tests
- `waterdrop upgrade` upgrade tool `waterdrop`


## Contributing

Contributions are always welcomed! You can start with the issues labeled with bug or feature.

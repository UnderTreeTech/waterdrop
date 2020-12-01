# CHANGELOG

## v1.0.0

We're exciting to release waterdrop GA version v1.0.0! Feel free to have a try.

### Optimize/Enhancement
- improve unit test coverage ([18581e2](https://github.com/UnderTreeTech/waterdrop/pull/63)) , ([0e53753](https://github.com/UnderTreeTech/waterdrop/pull/64))
- add sql breaker ([1e5dbd1](https://github.com/UnderTreeTech/waterdrop/pull/57))
- assign golangci-lint check folder ([b21c44e](https://github.com/UnderTreeTech/waterdrop/pull/56))
- provide google sre breaker and sentinel ([cebe08f](https://github.com/UnderTreeTech/waterdrop/pull/61))

### Bugfix
- fix interceptor does not take effect ([cebe08f](https://github.com/UnderTreeTech/waterdrop/pull/61))
- fix trace span leaky ([1086af6](https://github.com/UnderTreeTech/waterdrop/pull/60))

## v0.2.0

### Features
- integrate sentinel-go as rate limit component ([568554a](https://github.com/UnderTreeTech/waterdrop/pull/54))
- support websocket ([e57e676](https://github.com/UnderTreeTech/waterdrop/pull/39))
- add waterdrop tools: generate unit test, swagger definition file, pb file ([3020506](https://github.com/UnderTreeTech/waterdrop/pull/36))

### Optimize
- export redis Ping ([955278c](https://github.com/UnderTreeTech/waterdrop/pull/52))
- optimize trace context deadline ([73afc65](https://github.com/UnderTreeTech/waterdrop/pull/42))
- optimize http client X-Request-Timeout ([8b2bfd8](https://github.com/UnderTreeTech/waterdrop/pull/44))

### Bugfix
- set read limit ([eaa5b78](https://github.com/UnderTreeTech/waterdrop/pull/33))

## v0.1.0

### Features
- support grpc & http
- support global trace, default trace component is jaeger
- use etcd to govern service register and discovery
- integrate zap as the default log component
- default mysql and redis
- implement google sre breaker
- support kafka and rocketmq broker for async logic
- default support TOML config file parsing



 
# CHANGELOG

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
 
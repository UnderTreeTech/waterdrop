# CHANGELOG

## v1.3.1

### Optimize/Enhancement
- polish redis Ping code
- update README and CHANGELOG

### Tool
- add new and upgrade tool ([7cf0f29](https://github.com/UnderTreeTech/waterdrop/pull/132))

## v1.3.0

Hi erveryone, this is a major release including many improvments, features and bugfix. Highly recommend waterdrop users to upgrade to this version.

### Features
- auto set max procs ([907f039](https://github.com/UnderTreeTech/waterdrop/pull/98))
- support trie collection ([b3929aa](https://github.com/UnderTreeTech/waterdrop/pull/100))
- add slice extend methods ([b40cb61](https://github.com/UnderTreeTech/waterdrop/pull/104))
- add hash, aes and rsa crypto ([254263d](https://github.com/UnderTreeTech/waterdrop/pull/107), [cf1f66b](https://github.com/UnderTreeTech/waterdrop/pull/105), [35c32c5](https://github.com/UnderTreeTech/waterdrop/pull/106))
- add minio component ([eda78cc](https://github.com/UnderTreeTech/waterdrop/pull/123))
- use go-redis as the default redis driver ([d1d6bc0](https://github.com/UnderTreeTech/waterdrop/pull/124))

### Optimize/Enhancement
- upgrade etcd client to v3.5.0 ([1804300](https://github.com/UnderTreeTech/waterdrop/pull/95))
- use google.golang.org/protobuf globally  ([52b00d7](https://github.com/UnderTreeTech/waterdrop/pull/96))
- use testify package globally  ([f8632c7](https://github.com/UnderTreeTech/waterdrop/pull/109))
- remove business logic related http middlewares  ([6cc3af0](https://github.com/UnderTreeTech/waterdrop/pull/110))
- support set http metric namespace ([fe18d64](https://github.com/UnderTreeTech/waterdrop/pull/113))
- pass trace and request timeout between http&rpc services ([d7df45e](https://github.com/UnderTreeTech/waterdrop/pull/114))
- add breaker and metric to mongo client ([9986575](https://github.com/UnderTreeTech/waterdrop/pull/126))
- add breaker and metric to es client ([e7d2a9c](https://github.com/UnderTreeTech/waterdrop/pull/129))
- optimize breaker ([b7ff10a](https://github.com/UnderTreeTech/waterdrop/pull/130))
- fix linter ([a8c4c31](https://github.com/UnderTreeTech/waterdrop/pull/127), [e7e102e](https://github.com/UnderTreeTech/waterdrop/pull/125), [2e6474d](https://github.com/UnderTreeTech/waterdrop/pull/111))
- improve test case coverage ([c455f5c](https://github.com/UnderTreeTech/waterdrop/pull/121))

### Bugfix
- sql.parseDSN handle PostgreSQL configs ([18d78ca](https://github.com/UnderTreeTech/waterdrop/pull/117))

### Security
- fix incorrect conversion between integer types ([6b87e5b](https://github.com/UnderTreeTech/waterdrop/pull/102))

### Tool
- remove protobuf directory and upgrade swagger tool ([60ddd9c](https://github.com/UnderTreeTech/waterdrop/pull/108))

## v1.2.0

### Features
- add cors middleware ([f84e193](https://github.com/UnderTreeTech/waterdrop/pull/90))
- add SafeMap that implementation by using a sync.RWMutex ([98c25e4](https://github.com/UnderTreeTech/waterdrop/pull/91))
- automatic rotate log ([1e4df21](https://github.com/UnderTreeTech/waterdrop/pull/84))

### Optimize/Enhancement
- safe lru cache ([dc81f61](https://github.com/UnderTreeTech/waterdrop/pull/87))
- add example proto definition ([c17d1f2](https://github.com/UnderTreeTech/waterdrop/pull/92))

### Bugfix
- add lru element trigger deadlock ([5ac524d](https://github.com/UnderTreeTech/waterdrop/pull/89))

## v1.1.0

### Features
- adapter to postgresql ([8426f48](https://github.com/UnderTreeTech/waterdrop/pull/71))
- adapter to mongodb ([3feac0e](https://github.com/UnderTreeTech/waterdrop/pull/75))
- adapter to elastic search ([1c47ad7](https://github.com/UnderTreeTech/waterdrop/pull/79))
- add defer stack ([25cbae6](https://github.com/UnderTreeTech/waterdrop/pull/72))

### Optimize/Enhancement
- remove gogo protobuf dependency ([aaf8edf](https://github.com/UnderTreeTech/waterdrop/pull/70))
- generate protobuf code on windows ([342803e](https://github.com/UnderTreeTech/waterdrop/pull/76))
- support redis pipeline ([5a5374f](https://github.com/UnderTreeTech/waterdrop/pull/82))
- compatible with multi gopath ([bb4687c](https://github.com/UnderTreeTech/waterdrop/pull/80)). Thanks to @dirtyrain
- improve sql span security ([9073e9c](https://github.com/UnderTreeTech/waterdrop/pull/73))

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



 
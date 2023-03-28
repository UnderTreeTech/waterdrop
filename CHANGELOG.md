# CHANGELOG


## v1.3.5

This is a minor updates version.

### Features
- export Trie ([b9eb9cc](https://github.com/UnderTreeTech/waterdrop/commit/b9eb9cc4846254996c6ef16e47dc6ac50d30cfce))
- support set rocketmq producer group name ([e05a5aa](https://github.com/UnderTreeTech/waterdrop/commit/e05a5aa9bd6fae0411086e33b5259151180b6ece))
- export rsa private/public key parse method ([b658cd5](https://github.com/UnderTreeTech/waterdrop/commit/b658cd5e01dd0118771892b3c1146cbb5d90e6e4))
- remove operator character ([fbef575](https://github.com/UnderTreeTech/waterdrop/commit/fbef5753dac7654cfe770733d8306fd5a05ef9a7))
- support set http client request header ([5053f7f](https://github.com/UnderTreeTech/waterdrop/commit/5053f7fa118648051305766b70186282663f3e18))
- support escape large request log ([d816354](https://github.com/UnderTreeTech/waterdrop/commit/d81635403727658508d60c9054d086e80941a8ac))
- adjust error type ([961864b](https://github.com/UnderTreeTech/waterdrop/commit/961864b49460a591fbfc5e4b21dcec35793a07aa))
- generate duration/integer jitter ([7b328a8](https://github.com/UnderTreeTech/waterdrop/commit/7b328a87a0527a8e50a24f79912482cc0a446886))

### Tool
- ecode generate tool ([5aeb61f](https://github.com/UnderTreeTech/waterdrop/commit/5aeb61fc2cd2e8ec632f61a6d812e55d44304214))

## v1.3.4

Fix can't go get the latest v1.3.3 issue ([974444d](https://github.com/UnderTreeTech/waterdrop/pull/144))


## v1.3.3

This is a minor updates version.

### Optimize/Enhancement
- optimize rocket mq ([9c900fe](https://github.com/UnderTreeTech/waterdrop/pull/139))
- add more time helper functions ([a2fe5fe](https://github.com/UnderTreeTech/waterdrop/pull/138))
- add Highlight, BoolQuery and SearchResult alias ([e74b0fd](https://github.com/UnderTreeTech/waterdrop/pull/137))
- update dependencies ([17ff6c8](https://github.com/UnderTreeTech/waterdrop/pull/141)). Thanks to @andykis
- add log caller and app filed ([2ba23e0](https://github.com/UnderTreeTech/waterdrop/commit/2ba23e0da10dba65e37176e166102da9ceec8b7a))
- add more mongo var alias ([053897d](https://github.com/UnderTreeTech/waterdrop/commit/053897d565643c224817fad71a8f2add31488bcd))
- fix gjson ReDoS security ([0f43392](https://github.com/UnderTreeTech/waterdrop/pull/143))

### Bugfix
- add trace func ContextWithSpan ([ae369ac](https://github.com/UnderTreeTech/waterdrop/pull/136)). Thanks to @dirtyrain
- send msg get nil reply #142 ([eebcf35](https://github.com/UnderTreeTech/waterdrop/commit/eebcf356db57a755e3a9ce72c27d2af025d998f8)). Thanks to @Billxunyang


## v1.3.2

This is a minor updates version.

### Optimize/Enhancement
- add es sniff field ([9eb04ca](https://github.com/UnderTreeTech/waterdrop/commit/9eb04caf29ba620c2919a1fa54cf49a0e768a91c))
- add redis LPopN method, notice that it's only for redis version 6.2 and later ([9c900fe](https://github.com/UnderTreeTech/waterdrop/commit/9c900fe849971401409df947b263607a0b0becf0))
- unify host and peer_ip to peer ([eeb196b](https://github.com/UnderTreeTech/waterdrop/commit/eeb196bcbf48e41411b6ed2a603cf35cbe285fe0))
- fix typo ([dd0e83f](https://github.com/UnderTreeTech/waterdrop/pull/134)). Thanks to @coosir

### Bugfix
- fix es `nil` return panic ([a7da9d5](https://github.com/UnderTreeTech/waterdrop/commit/a7da9d5120ea3929faa36b3693d5211b71e090f8))


## v1.3.1

### Optimize/Enhancement
- polish redis Ping code
- update README and CHANGELOG
- remove examples

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



 
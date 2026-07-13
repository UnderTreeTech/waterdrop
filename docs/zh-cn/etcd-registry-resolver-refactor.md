# etcd registry / resolver 重构分析（适配 grpc 1.80.0）

> 背景：waterdrop 把 grpc 从 1.40.0 升级到 1.80.0 后，etcd registry 反复出问题。本文归纳问题定位与修复过程，并记录尚未处理的遗留项，供后续维护参考。

## 一、问题现象

grpc 升级到 1.80.0 后，基于 etcd registry 的服务发现出现：

- 客户端连接一抖动，服务端注册就掉；
- etcd watch channel 关闭后日志刷屏、goroutine 泄漏；
- 后端全部下线时 balancer 仍抱着死地址；
- 客户端拨不通、`url.Parse` 静默丢地址等。

## 二、根因：`EtcdRegistry` 四重身份揉在一个共享 client 上

`pkg/registry/etcd/etcd.go` 旧实现里，`EtcdRegistry` 同时承担：

1. **服务端 registry**：`Register` / `DeRegister` / `KeepAlive`；
2. **resolver Builder**：`Build` / `Scheme`；
3. **resolver Resolver**：`Build` 返回 `e` 本身，`ResolveNow` / `Close`；
4. **分布式锁**：`lock.go` 的 `NewMutex` 复用同一个 `er.client`。

四者共享同一个 `*clientv3.Client`，且 `Build` 返回单例 `e` 当 resolver。

### 致命点：`Close()` 会关闭 etcd client 本身

旧 `Close()`：

```go
func (e *EtcdRegistry) Close() {
	var wg sync.WaitGroup
	e.services.Range(func(k, v interface{}) bool {
		wg.Add(1)
		go func(k interface{}) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			e.deRegister(ctx, k.(string)) // ① deregister 所有服务
			cancel()
		}(k)
		return true
	})
	wg.Wait()
	e.client.Close() // ② 关闭整个 etcd client
}
```

grpc 1.80 在 `ClientConn.Close()` 时会调 `Resolver.Close()`。旧代码 `Build` 返回 `e`，于是**关一个客户端连接 → `EtcdRegistry.Close()` → `e.client.Close()`**：

- 共享 etcd client 没了 → 其他依赖的 watch / KeepAlive 全断；
- 第 ① 步把本进程通过 `Register` 注册的服务也全部 deregister → 本进程对外服务从 etcd 消失；
- `lock.go` 的 `NewMutex` 也用同一 client，分布式锁一并失效。

依赖越多越容易触发"关一个连累一片"。这是"etcd registry 总出问题"的直接元凶。

## 三、旧 `etcd.go` 的其它缺陷

| # | 缺陷 | 位置 | 影响 |
|---|---|---|---|
| 1 | `Close()` 关共享 client | 旧 `Close()` | 见上，灾难级 |
| 2 | `watch` 用 `context.Background()`、不可取消、chan 关闭后 10s 重试对已关闭 client 永远失败 | 旧 `watch()` | goroutine 泄漏 + 日志刷屏 |
| 3 | `updateAddrs` 忽略 `UpdateState` 返回值、从不 `ReportError`、空地址跳过 `UpdateState` | 旧 `updateAddrs` | balancer 留陈旧状态、不退避、抱死地址 |
| 4 | list-then-watch 顺序丢事件 | 旧 `watch()` | Get 与 Watch 之间的事件丢失 |
| 5 | `Build` 异步且不重试 | 旧 `Build` | 首次 Get 失败则永不 `UpdateState`，通道永久卡 CONNECTING |

### 附：addr 格式相关（独立隐患）

`getAddrs` 用 `url.Parse(service.Addr)` 取 `u.Host`。裸 `IP:PORT` 会让 `url.Parse` **报错**（`first path segment in URL cannot contain colon`）→ `continue` 丢地址 → 服务对消费方隐形，但 `Register`/`List` 不报错（它们不调 `url.Parse`），表现为"注册成功、List 能查到、客户端死活拨不通"。

**约定：`ServiceInfo.Addr` 必须是 `scheme://ip:port` 形式**，且 `scheme` 与 `ServiceInfo.Scheme`（`grpc`/`http`）一致。`pkg/stats/stats.go:56` 即按此约定拼装。

## 四、参考的 resolver 设计要点

正确的 resolver 设计应与 registry 解耦，关键要点：

- resolver 拆成独立轻量结构，`Close()` 只取消自身 context、停止自身 watch，**不碰共享 etcd client**；
- watcher 先建 watch 再拉全量（如 `WithRev(0)`），chan 关闭后 1s backoff 重连，每轮检查 `ctx.Done()` 可取消；
- keepalive/重注循环在重注前判 `ctx.Err()`，关闭即退不重注；重试有上限 + 指数退避 + 随机抖动；`Grant`/`Put` 套超时防卡死。

grpc 1.80 关键 API：

- `resolver.Target.Endpoint()`：返回去掉前导 `/` 的 `URL.Path`（存在，可用）；
- `resolver.ClientConn.UpdateState`：返回 `error`，需处理；
- `resolver.ClientConn.ReportError`：通知 balancer 进入退避；
- `resolver.Resolver` 接口：`ResolveNow` + `Close`，`Close` 在 `ClientConn.Close()` 时被调。

## 五、本次重构方案（已落地）

### 1. 拆出独立 `etcdResolver`

```go
type etcdResolver struct {
	e        *EtcdRegistry
	cc       resolver.ClientConn
	watchKey string
	ctx      context.Context
	cancel   context.CancelFunc
}

func (r *etcdResolver) Close()            { r.cancel() } // 只取消自身，不碰共享 client
func (r *etcdResolver) ResolveNow(_ resolver.ResolveNowOptions) {}
```

`Build` 返回 `etcdResolver` 而非 `e`；`EtcdRegistry` 只保留 `Builder` 身份（`Build`+`Scheme`），不再实现 `Resolver`。

### 2. 可取消 + 1s backoff 的 watch 循环

```go
func (r *etcdResolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		wch := r.e.client.Watch(r.ctx, r.watchKey, clientv3.WithPrefix()) // 先 watch
		r.updateAddrs()                                                    // 再拉全量，避免丢事件
		for event := range wch {
			for _, ev := range event.Events {
				if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
					r.updateAddrs()
				}
			}
		}
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(time.Second): // 1s backoff 重连
		}
	}
}
```

### 3. `updateAddrs` 错误处理

- `Get` 失败：`cc.ReportError(err)` 让 balancer 退避；
- `UpdateState` 返回值记录；
- 空地址不 `UpdateState`（避免 balancer 置零后端），记 warn；
- `Get`/`Watch` 用 `r.ctx` 而非 `context.Background()`，可取消。

### 4. 验证

- `go build` / `go vet` 通过；
- 本地起 etcd（`/usr/local/bin/etcd` v3.5.27）跑通 `TestEtcdRegistry`、`TestEtcdResolver`、`TestLock`；
- `TestEtcdResolver` 覆盖：注册触发 watch → `UpdateState` 推地址；空地址不推；`resolver.Close()` 后共享 etcd client 仍可用（核心修复断言）。

### 5. 对外 API 不变

`New` / `Register` / `DeRegister` / `List` / `Close` / `Build` / `Scheme` / `ResolveNow` 签名不变，外部 `resolver.Register` 装配无需改动。

## 六、多依赖场景下的行为

服务 A 依赖 B、C、D：

- 每个依赖 `grpc.NewClient("etcd:///X")` → gRPC 调一次 `Build` → **产生一个独立 `etcdResolver`**；
- N 个依赖 = N 个 `etcdResolver`，各有 `cc`/`ctx`/`watchKey`/`watch` goroutine，共享同一个 `e.client`；
- etcd clientv3 并发安全，多 watch 流在客户端内部复用到少量 gRPC stream；
- 关掉到 B 的 `ClientConn` → 只 cancel B 的 resolver → 只 B 的 watch goroutine 退出；C、D 与共享 client 不受影响。

对比旧实现：`Build` 返回单例 `e`，多个 `ClientConn` 共用同一 resolver 对象、状态混杂，且**关任意一个** `ClientConn` → `EtcdRegistry.Close()` → 共享 client 关闭、服务全部 deregister、分布式锁失效。

## 七、遗留项（尚未处理）

### 1. `Register` 方法的若干问题

| # | 问题 | 严重度 |
|---|---|---|
| a | 重复 Register 同一 key → `e.cancels[key]` 被覆盖、旧 keepalive goroutine + lease 永久泄漏 | 中（仅同进程重复注册时触发） |
| b | `val, _ := json.Marshal(info)` 吞掉 marshal 错误，可能写空 value | 低 |
| c | 重注循环无上限、1s 硬节奏打到底、刷屏 error | 低 |
| d | Grant 成功后 Put 失败 → 孤儿 lease（等 TTL 过期回收） | 低 |
| e | `int64(e.config.RegisterTTL/1e9)` 可读性差，建议 `int64(... / time.Second)` | nit |

### 2. keepalive goroutine 的重注逻辑（设计层面）

#### 2.1 `!ok` 重注的存在必要性

`Register` 里 `keepAliveCh` 返回 `!ok`（channel 关闭）时触发重注（`etcd.go:131-168`），目的是"**瞬态抖动后自愈**"。切到新 resolver **不改变这个需求**——它防的是服务端注册丢失，与 resolver 无关。

`keepAliveCh` 关闭的几种来源：

1. `keepAliveCtx` 被 cancel（DeRegister/Close 主动发起）——理论上由外层 `case <-keepAliveCtx.Done(): return` 先接住，不进 `!ok`；
2. **etcd client 到 server 的 gRPC 连接断开**（网络抖动、leader 切换、etcd 重启、中间件掐连接）——etcd clientv3 会 close 该 lease 的 KeepAlive channel，且重连后**不会自动恢复**之前的 KeepAlive；
3. lease 已过期/被 revoke（TTL 内没续上或被别人 revoke）。

**不重注的后果**：来源 2/3 发生后，旧 lease 无人续，会在剩余 TTL 后过期 → key 被 etcd 自动删除 → 本进程服务从 etcd 消失，直到进程重启。消费方收到 DELETE 摘地址，但本进程不会自己回来。所以**一次 transient 抖动就会让注册永久丢失**——这正是 `!ok` 重注存在的理由，不能简单删掉。

#### 2.2 keepalive channel 关闭后 etcd client 不会自动恢复

这是保留重注逻辑的关键依据：etcd clientv3 在底层 gRPC 连接断开重连后，**不会自动把之前那个 lease 的 KeepAlive 续上**（KeepAlive channel 一旦 close 就结束了）。所以即便连接很快恢复（< `RegisterTTL`），也得靠这段重注重新 `Grant`+`Put`+`KeepAlive`。可在升级 etcd client 版本后实测验证：手动掐 etcd 连接或重启 etcd，看不靠重注、lease 能否被自动续上；能则可简化，不能则保留。

#### 2.3 当前实现的竞态：诈尸（Put-after-Delete）

当前实现把"瞬态自愈"与"关闭收尾"复用同一条路径，区分两者靠 `cancel` 先到 + 1s sleep + `Delete` 兜底 + lease TTL 的接力，**是竞态不是保证**。

诈尸窗口（key 被重建在 `Delete` 之后）的精确时序：

```
1. keepAliveCh 因 transient 原因关闭（ctx 未 cancel）
2. !ok 分支：log → sleep 1s → 内层 for
3. 内层 line 138 的 ctx 检查通过（此时未关闭）→ Grant(新 lease L2)
4. 与此同时 DeRegister：cancel() + Delete(key)   ← 把旧 key 删了
5. 内层继续：Put(key, L2)                          ← key 被重建（Put-after-Delete）
6. 内层：KeepAlive(L2) → break
→ key 以 L2 重新存在，但 Delete 已跑完不会再删 → 鬼 key，挂到 L2 过期（~RegisterTTL）
```

后果：鬼 key 残留 ≤ `RegisterTTL`（默认 90s），期间消费方可能向已下线实例发流量。本就是历史行为，不是回归，属"可修不必须"。

#### 2.4 点状 ctx 检查挡不住诈尸

重注路径里有两处 ctx 检查：外层 `case <-keepAliveCtx.Done(): return`（line 129），内层 `select { case <-keepAliveCtx.Done(): return; default: }`（line 138，非阻塞）。**两者都是"进 `Grant`/`Put` 之前"的瞬时点状检查，都挡不住上面的第 5 步**——检查已过、`Delete` 在中间插入、`Put` 照样执行。

特别澄清一个常见误解：**在 `!ok` 入口加 `if keepAliveCtx.Err() != nil { return }` 预判，并不能防诈尸**。它语义上等价于 line 138 的非阻塞 select，区别只是位置——在当前结构下，它仅能省掉 line 133 的 Warnf 和 line 134 的 1s sleep（"ctx 在进入 `!ok` 时就已完成"的情况，line 138 会在 1s 后同样退出）；若 ctx 在 sleep 期间被 cancel，预判反而抓不到、只有 line 138 抓得到。所以预判是个微优化（快 1s + 少一条日志），**不是诈尸的解法**。若把 line 134 的 sleep 挪进循环当失败 backoff（入口与首次循环检查背靠背），预判就完全冗余、可删。

#### 2.5 真要堵诈尸：cancel → join → Delete 顺序

点状检查不够，得改 `DeRegister` 的收尾顺序，保证不会有 `Put` 跑在 `Delete` 之后：

```go
func (e *EtcdRegistry) deRegister(ctx context.Context, key string) error {
    if cancel, ok := e.cancels.LoadAndDelete(key); ok {
        cancel.(context.CancelFunc)()   // ① 先 cancel keepalive ctx
        <-e.done(key)                    // ② 等 keepalive goroutine 确认退出（可套超时兜底）
    }
    _, err := e.client.Delete(ctx, key)  // ③ 再 Delete
    ...
}
```

代价：`DeRegister` 要等 goroutine 退出（最多等它当前那次 `Grant`/`Put` 返回），需配超时兜底，改动跨 `deRegister` 与 goroutine 的退出同步。或接受现状（鬼 key ≤ `RegisterTTL`）。

#### 2.6 其它低风险遗留（独立于诈尸）

| # | 问题 | 严重度 |
|---|---|---|
| a | 重复 Register 同一 key → `e.cancels[key]` 被覆盖、旧 keepalive goroutine + lease 永久泄漏 | 中（仅同进程重复注册时触发） |
| b | `val, _ := json.Marshal(info)` 吞掉 marshal 错误，可能写空 value | 低 |
| c | 重注内层 for 无上限、1s 硬节奏打到底、连接长断时刷屏 error | 低 |
| d | Grant 成功后 Put 失败 → 孤儿 lease（等 TTL 过期回收） | 低 |
| e | `int64(e.config.RegisterTTL/1e9)` 可读性差，建议 `int64(... / time.Second)` | nit |

建议（低风险、不动收尾顺序）：

- 重试加上限（`maxRetry`）+ 指数退避（1s→…→30s 封顶，`min` 需 Go 1.21+，waterdrop 1.25 已满足）；
- `Grant`/`Put` 套超时防卡死；
- 开头 `e.cancels.LoadAndDelete(key)` 先停旧 goroutine，修重复注册泄漏；
- 顺带处理 marshal 错误。

诈尸（2.3/2.5）是否上 `cancel→join→Delete` 需单独决策，改动更大、要动 `deRegister` 与 goroutine 退出同步，本节仅记录分析与方案，**暂不落地**。

### 3. resolver 的 `ResolveNow` 仍是 no-op

gRPC 在 `pick_first` 全失败、outlier ejection 等场景会调 `ResolveNow` 邀请重新解析，当前是 no-op。若 watch 侧漏事件且未补全量，客户端可能卡住。可考虑 `ResolveNow` 触发一次 `updateAddrs`（加去抖防重入）。

## 八、`UpdateState` 的驱动来源澄清

- **本地唯一推 State 的调用点：`updateAddrs`**（全包仅一处 `cc.UpdateState`，外加一处 `cc.ReportError`）；
- gRPC 反向驱动（`ResolveNow`、`UpdateState` 返错重试）当前未接成动作，实际不触发 Update；
- balancer 内部摘后端（连接失败 / outlier detection / health check）让"可用后端"变化，但**不经过、也不改动** resolver 推的 `State`；
- 每次 `UpdateState` 是整体替换，不存在被框架历史值覆盖。

## 九、关联提交

- `fix(etcd): separate resolver lifecycle from registry (grpc 1.80 fix)` —— 本次 resolver 重构（拆出 `etcdResolver`、可取消 watch、错误处理、新增 `TestEtcdResolver`）。
- `[bugfix] fix etcd keepalive grant new lease`（`12e0958`）—— 把 `Register` 的 30s 重 `Put` ticker 换成原生 `Lease.KeepAlive`；是"`resolver N peer service` 周期性打印消失"的真正来源，详见第十一节。

## 十、etcd 事件语义澄清：`KeepAlive` 不产生 watch 事件

> 定位"收不到 keepalive 的 put 事件"时容易踩的坑：**etcd 的 `Lease.KeepAlive` 根本不产生 watch 事件**。下面用本地 etcd v3.5.27 实测确认。

实测步骤：`Grant(10s)` → `Put(带 lease)` → `KeepAlive` 连续 3 轮 → `Revoke`，全程 watch 前缀。结果：

| 操作 | watch 收到的事件 |
|---|---|
| `Put`（带 lease，即注册） | 1× PUT |
| `KeepAlive` × 3 轮（TTL=10） | **0** |
| `Revoke`（lease 过期/下线） | 1× DELETE |

`LeaseKeepAlive` 只是把 lease 的剩余 TTL 续回去，**不碰 keyspace、不改 revision**，所以不产生任何 mvcc 事件。这是 etcd 固有行为，与 waterdrop 版本无关。

因此 resolver 真正会响应的事件只有三类：

1. 首次 `Put`（注册）→ PUT；
2. keepalive channel 关闭后**重注的 re-Put**（`etcd.go` 的 `!ok` 分支）→ PUT；
3. `Delete` / lease 过期 / `Revoke` → DELETE。

稳态下 producer 只做 keepalive（不重 `Put`）→ 没有事件 → 事件驱动的 resolver 自然不刷新、不打印。**这是预期，不是漏事件。** 本节是理解下一节的前提。

## 十一、`resolver N peer service` 周期性打印消失的根因（`12e0958`，非 `b3924dd`）

> 现象：升级前的老服务会**定期**打印 `log.Debugf("resolver %d peer service", len(addrs))`（`getAddrs` 里，每次 `updateAddrs` 都会打），升级后不再打印。曾被误判为"`b3924dd` 重构导致 resolver 收不到 keepalive 的 put 事件"，实际不是。

### 老服务为什么"定期"打印

`12e0958` 之前的 `Register`（`ad82e15` 一路到 `490e5da`/`f538906` 都是）**不用 etcd 的 `Lease.KeepAlive`**，而是用 ticker 每 `RegisterTTL/3`（默认 30s）重新 `Grant` + `Put`，靠频繁重写"续命"：

```go
// 旧 Register（12e0958 之前）
func (e *EtcdRegistry) Register(ctx context.Context, info *registry.ServiceInfo) error {
    err := e.register(ctx, info) // Grant + Put（带 lease）
    if err != nil {
        return err
    }
    go func() {
        ticker := time.NewTicker(e.config.RegisterTTL / 3) // 90s/3 = 30s
        for {
            select {
            case <-ticker.C:
                e.register(ctx, info) // 重新 Grant + Put
            }
        }
    }()
    return nil
}
```

每次重 `Put` 都是真正的 `Put` → 产生 **PUT watch 事件** → 消费方 resolver 的 `watch` 收到 → 调 `updateAddrs` → 调 `getAddrs` → 打印 `Debugf`。**所以老服务每 30s 打一次，节奏 = producer 的重 `Put` ticker。** 当初以为的"keepalive 产生的 put 事件"，其实是这个 ticker 重 `Put`。

注意：老 `watch` 本身**一直是事件驱动**的（`client.Watch` + `for event := range`，初始拉一次 + 事件触发），并没有 ticker。周期性打印完全来自 producer 侧的周期性重 `Put`，不是 resolver 在轮询。

### 新服务为什么不再打印

`12e0958 [bugfix] fix etcd keepalive grant new lease` 删掉上面的 ticker，换成 etcd 原生 `KeepAlive`（即现在 `Register` 里 `e.client.KeepAlive(...)` 那段，`etcd.go:116`）：

```go
// 12e0958 之后（含 b3924dd）
keepAliveCh, err := e.client.KeepAlive(keepAliveCtx, grant.ID) // 原生租约续期，不产生事件
```

而原生 `KeepAlive` 不产生 watch 事件（见第十节）。于是：

- producer 不再每 30s 重 `Put` → 没有周期性 PUT 事件；
- 消费方 resolver 收不到周期性事件 → 不再周期性调 `updateAddrs`/`getAddrs`；
- `Debugf` 自然不再周期性打印。

resolver **没坏**，行为反而更对：启动拉一次全量快照（会打一次），之后只在**真正变化**时刷新——新实例注册（PUT）、实例下线/lease 过期（DELETE）。稳态没变化就不打印，符合预期。

### 和 `b3924dd` 的关系

**这个现象与 `b3924dd` 无关。** 周期性打印在 `12e0958`（`b3924dd` 的上一个提交）就已经消失。`b3924dd` 只重构 resolver 生命周期，没动 `Register`/keepalive。被归因到 `b3924dd`，多半是从 `12e0958` 之前的版本直接升到 `b3924dd`，两个改动一起落地所致。

### 验证 resolver 仍存活

想确认 resolver 只是"不再周期性刷"而非"坏了"：

- 重启任一依赖实例（或 graceful 下线再起）→ 应能看到一次 `resolver N peer service`（实例数变化触发）；
- 或临时调小某 producer 的 `RegisterTTL` 让 lease 过期 → 看到 DELETE 触发的刷新。

"有真实变化时能刷新"即说明 resolver 正常。旧版"30s 一刷"本质是 producer 反复重写 key，属旧实现的不优雅，`12e0958` 已修掉。

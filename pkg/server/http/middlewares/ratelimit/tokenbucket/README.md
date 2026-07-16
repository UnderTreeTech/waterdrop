# tokenbucket HTTP 限流中间件

基于 [`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate) 的轻量级、进程内令牌桶限流中间件，支持**路由级**限流：配置了路由规则的路由按路由限流，其余路由回退到**全局**限流。

它是 `ratelimit/sentinel` 的本地、无外部依赖对位方案——sentinel 走规则/集群，tokenbucket 走单机内存令牌桶，适合不需要分布式一致性的场景。

## 特性

- **路由级 + 全局限流**：路由规则优先，未命中的路由走全局兜底。
- **仅路由限流也支持**：全局不配置时等价于不限流，可只对个别路由限流。
- **非阻塞**：拿不到令牌立即返回 429，不排队等待（避免 goroutine 堆积）。
- **复用既有抽象**：与 `ratelimit/sentinel` 共用 `ratelimit.Option`（`WithFallback` / `WithResourceStrategy`），可互换。
- **并发安全**：令牌桶在构造期一次性建好，请求期只读；`rate.Limiter` 自身并发安全。

## 快速接入

```go
import (
    "golang.org/x/time/rate"

    tb "github.com/UnderTreeTech/waterdrop/pkg/server/http/middlewares/ratelimit/tokenbucket"
)

// 在创建 gin engine 后挂载中间件
engine.Use(tb.New(&tb.Config{
    Global: tb.Rate{Rate: rate.Limit(100), Burst: 200}, // 全局：100 QPS，突发 200
    Routes: map[string]tb.Rate{
        "GET:/api/app/secrets": {Rate: rate.Limit(10), Burst: 20}, // 该路由单独 10 QPS
    },
}))
```

被限流的请求默认返回 `429 Too Many Requests`。

## 配置说明

### `Config`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Global` | `Rate` | 应用于未在 `Routes` 中列出的路由。**零值表示不限流**。 |
| `Routes` | `map[string]Rate` | 路由 key 到限流规则的映射。key 默认为 `METHOD:FullPath`，可被 `WithResourceStrategy` 覆盖。优先级高于 `Global`。 |

### `Rate`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Rate` | `rate.Limit` | 每秒令牌数。`rate.Limit(100)` = 100 QPS；`rate.Every(time.Second/10)` = 每 100ms 一个令牌；`rate.Inf` = 不限流。 |
| `Burst` | `int` | 桶大小，即允许的瞬时突发请求数，超过后按 `Rate` 速率放行。 |

> **零值安全**：`Rate{}`（`Rate<=0 && Burst<=0`）会被当作 `rate.Inf` 不限流，避免「忘了配 Global → 全部请求被拒」的坑。

## 使用场景

### 1. 仅全局限流

所有路由共用一个桶。

```go
engine.Use(tb.New(&tb.Config{
    Global: tb.Rate{Rate: rate.Limit(100), Burst: 200},
}))
```

### 2. 仅路由限流（全局不配限流）

**支持。** 留空 `Global`（零值 = 不限流），只配 `Routes`：命中的路由按规则限流，其它路由放行。

```go
engine.Use(tb.New(&tb.Config{
    // Global 留空 → 全局不限流
    Routes: map[string]tb.Rate{
        "GET:/expensive": {Rate: rate.Limit(10), Burst: 20}, // 仅此路由 10 QPS
    },
}))
// GET:/expensive 受限；GET:/healthz、GET:/ping 等其它路由放行
```

### 3. 全局 + 路由限流（推荐）

全局兜底 + 重点路由单独收紧。

```go
engine.Use(tb.New(&tb.Config{
    Global: tb.Rate{Rate: rate.Limit(1000), Burst: 1000},
    Routes: map[string]tb.Rate{
        "POST:/api/login":   {Rate: rate.Limit(5), Burst: 5},   // 登录接口单独收紧
        "GET:/api/export":   {Rate: rate.Limit(1), Burst: 3},   // 导出接口限流
    },
}))
```

## 路由 key 规则

默认 key = `c.Request.Method + ":" + c.FullPath()`，例如：

| 路由注册 | 请求 | 默认 key |
| --- | --- | --- |
| `engine.GET("/api/app/secrets", ...)` | `GET /api/app/secrets` | `GET:/api/app/secrets` |
| `engine.POST("/api/app/validate/:id", ...)` | `POST /api/app/validate/42` | `POST:/api/app/validate/:id` |

注意：`c.FullPath()` 返回的是**路由模板**（`/api/app/validate/:id`），不是实际请求路径（`/api/app/validate/42`）。因此 `Routes` 的 key 要写模板路径，这样同一路由的所有参数化请求共享同一个桶。

可用 `ratelimit.WithResourceStrategy` 自定义 key，例如去掉方法前缀、按 header 区分租户等：

```go
engine.Use(tb.New(&tb.Config{
    Global: tb.Rate{Rate: rate.Limit(100), Burst: 100},
}, ratelimit.WithResourceStrategy(func(c *gin.Context) string {
    return c.FullPath() // 只用路径，不带方法
})))
```

## 进阶选项

### `WithFallback`：自定义被限流后的响应

默认返回 429。设置 fallback 后由你决定响应内容；中间件会在 fallback 执行后 `c.Abort()`，被限流的请求**不会**继续走到业务 handler。

```go
engine.Use(tb.New(&tb.Config{
    Global: tb.Rate{Rate: rate.Limit(100), Burst: 100},
}, ratelimit.WithFallback(func(c *gin.Context) {
    c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "请求过于频繁，请稍后重试"})
})))
```

### `WithResourceStrategy`：自定义限流 key

见上节「路由 key 规则」。

## 与 `ratelimit/sentinel` 的选择

| | tokenbucket | sentinel |
| --- | --- | --- |
| 实现 | `golang.org/x/time/rate` 单机令牌桶 | `alibaba/sentinel-golang` 规则引擎 |
| 部署 | 进程内，无外部依赖 | 单机或集群（规则可热更） |
| 配置 | 代码内 `Config` | 规则文件 / `flow.LoadRules` |
| 适合 | 单机限流、简单场景、无外部依赖 | 需要规则动态下发、多种流控策略 |

两者共用 `ratelimit.Option`，切换成本低。

## 运行测试

```bash
go test ./pkg/server/http/middlewares/ratelimit/tokenbucket/... -count=1
```

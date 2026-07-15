# waterdrop trace 工厂化重构设计（Tracer 抽象 + 工厂 + opentracing↔OTel bridge）

> 范围：把 waterdrop trace 子系统从「中间件内 jaeger+OTel 双分支合并」重构为「`Tracer` 抽象 + 工厂按配置选 jaeger/otel/dual」体系；db/broker 层的 opentracing span 经 bridge 在 OTel/dual 模式进 OTel 后端。
>
> 状态：**设计阶段，待 review，未启动编码**。本文档承接 trace unification「方案 W 合并版」暴露的运行时问题，升级为框架级抽象方案。review 通过后再决定是否进入编码。
>
> 背景：方案 W（合并版）已落地--OTel 分支并入既有 `Trace` 中间件 / `TraceForUnary*` 拦截器，`jaeger.Init()` 读平级 `[trace.otel]` 段。本文档取代其中间件双分支实现；`jaeger.Init()` 仍是服务入口（向后兼容）。

---

## 0. 背景与触发问题

方案 W（合并版）运行时实测暴露三个问题：

### 0.1 `trace.TraceID(ctx)` 只认 jaeger

`pkg/trace/trace.go:107`：
```go
func TraceID(ctx context.Context) string {
	sp := SpanFromContext(ctx)                       // opentracing span（activeSpanKey）
	if sp == nil { return "" }
	if jsc, ok := sp.Context().(jaeger.SpanContext); ok {   // 类型断言成 jaeger.SpanContext
		return jsc.TraceID().String()
	}
	return ""                                        // 否则空
}
```
- **纯 OTel 模式**：无 jaeger tracer，中间件 jaeger 分支起的是 noop span（`noopSpanContext`），断言失败 -> `trace.TraceID` 返回 `""`。
- **双轨模式**：ctx 上 jaeger span 与 OTel span 经不同 context key 共存（`activeSpanKey` vs `currentSpanKey`），`SpanFromContext` 取的是 jaeger span -> 返回 **jaeger 64-bit id**，非 OTel 128-bit id。

### 0.2 OTel span TraceID 全零

`[trace.otel] endpoint=""`（provider 交外部，如业务侧 ADK `langfuse.Setup`）时，`otel.InitWithConfig` 只装 propagator、不建 provider -> 全局 TracerProvider 是 noop -> `tracer.Start()` 返回 noop span -> **TraceID 全零**（`0000…0000`），比空更误导。若外部 provider 未就位，X-Trace-Id 头会是全零字符串。

### 0.3 db/broker 层 span 在 OTel 模式丢失

db/broker 层（`pkg/database/{mongo,sql,redis,es}`、`pkg/broker/rocketmq`）用 opentracing API：
- `trace.StartSpanFromContext(ctx, "query")` 返回 `opentracing.Span`（29 处）
- `ext.Component.Set(span, "mongo")` / `ext.PeerAddress.Set(...)` / `span.LogFields(...)`（~80 处）
- 6 个结构体字段持有 `opentracing.Span`（mongo aggregate/bulk/query、sql Stmt/Tx/conn）

纯 OTel 模式下全局 opentracing tracer 是 noop -> 这些 span **不记录** -> OTel 后端缺 db/broker span，可观测性断裂。

### 0.4 根因

trace 体系**没有抽象层**：调用方直接捏 opentracing API + 中间件内散落 OTel 分支，模式由运行时 `if otel.Enabled()` 隐式决定而非配置显式驱动；db/broker 的 opentracing span 与 OTel 后端之间无桥接。

---

## 1. 目标 / 非目标

### 1.1 目标

1. **工厂驱动模式**：`Init()` 按配置自动选 jaeger / otel / dual 三体系之一，中间件/拦截器只调抽象 `Tracer` 接口（无双分支）。
2. **`trace.TraceID` 三模式可用**：jaeger 模式返 jaeger id；otel/dual 模式返 OTel 128-bit id（非空、非全零）。
3. **db/broker OTel 覆盖**：otel/dual 模式下 db/broker 的 opentracing span 经 bridge 进 OTel 后端，db/broker 代码零改动。
4. **provider 全零防护**：otel 模式 endpoint 空且未声明 external 时退化（不输出全零 id）。
5. **零服务改动 / 前向兼容**：服务仍 `defer jaeger.Init()()`；不配 trace 段的老服务默认 jaeger（hostname），行为同改前。

### 1.2 非目标

- 不改 db/broker 层的 opentracing 调用代码（bridge 让其三模式都生效）。
- 不下线 jaeger（dual 仍并行）。
- dual 下 jaeger 后端补 db/broker span（composite opentracing tracer）列后续增强。
- 不涉及业务服务侧配置/代码（框架就绪后由各服务自行适配 `[trace.otel]` 等）。

---

## 2. 现状核实（以当前代码为准）

| 事实 | 位置 | 含义 |
|---|---|---|
| `trace.TraceID` 对 `jaeger.SpanContext` 断言 | `pkg/trace/trace.go:107` | 纯 otel 返空、双轨返 jaeger id（问题 1） |
| 中间件 jaeger 分支无条件跑 + OTel 分支 `otel.Enabled()` 门控 | `pkg/server/http/middlewares/trace.go:54/82`、`pkg/server/rpc/interceptors/trace.go:51/60/101/109` | 双分支合并，纯 otel 仍跑 noop jaeger span |
| jaeger/OTel span 经不同 context key 共存 | opentracing `activeSpanKey`（`opentracing-go/gocontext.go:7`）vs OTel `currentSpanKey`（`otel/trace@v1.43.0/context.go:9`） | 双轨同 ctx 两 span，`SpanFromContext` 取 jaeger |
| `otel.InitWithConfig` endpoint 空只装 propagator | `pkg/trace/otel/otel.go:124` | 无 provider -> noop span -> 全零 TraceID（问题 2） |
| db/broker 用 opentracing Span/ext.*/log.* | `pkg/database/{mongo,sql,redis,es}/*`、`pkg/broker/rocketmq/*`（29 处 StartSpan + ~80 处 ext/log） | 纯 otel 进 noop 不记录（问题 3） |
| `jaeger.Init()` 是服务入口 | `pkg/trace/jaeger/jaeger.go:142`（waterdrop 仓内无调用，由下游服务 `defer jaeger.Init()()` 调） | 入口需保留向后兼容 |
| 中间件/拦截器链注册硬编码 | `pkg/server/http/server/server.go:60`、`pkg/server/rpc/server/server.go:76`、`client/client.go:68` | 签名不变则链注册不动 |
| `bridge/opentracing` 不在依赖树 | go.sum 无；v1.43.0 可拉（与 otel v1.43.0 配对） | 需新增依赖 v1.43.0 |

---

## 3. 设计决策（review 时请确认）

| # | 决策点 | 选定 | 备选 / 取舍 |
|---|---|---|---|
| D1 | 模式选择 | **自动检测**（按存在的段） | 显式 `[trace] mode` 字段；自动检测零配置改动，与现行为一致 |
| D2 | dual 模式 | **保留**（composite Tracer，同 TraceID） | 砍 dual 只单体系；保留则 jaeger+OTel 同 trace 免 bridge |
| D3 | 旧 `trace.*` API | **作 facade 适配新接口**（签名不变） | 彻底换新 API 全量改调用方；facade 最小 blast radius |
| D4 | otel provider | **endpoint 自建 / external 等外部 / 都无则退化**（三态） | 强制 endpoint 自建；三态兼顾「外部 provider」场景 |
| D5 | db/broker OTel 覆盖 | **opentracing↔OTel bridge**（`bridge/opentracing`） | 手写映射 ~80 处；bridge 零改 db/broker 代码 |
| D6 | dual 下 db/broker 去向 | **进 OTel**（jaeger 侧仅收 server/client span） | composite opentracing tracer 两后端都收；本轮不做，列后续 |

---

## 4. 关键技术：opentracing↔OTel bridge（D5 的核心）

`go.opentelemetry.io/otel/bridge/opentracing`（v1.43.0）的 `BridgeTracer` 实现 `opentracing.Tracer`，其 `StartSpan` 产出的 `bridgeSpan`（`bridge.go:89`）内部持有一个真实 OTel span，并把：
- `SetTag(key, value)`（`bridge.go:154`）-> OTel attribute
- `LogFields(...)`（`bridge.go:168`）/ `LogKV(...)`（`bridge.go:237`）-> OTel event
- `Finish()`（`bridge.go:109`）-> OTel span End

转译。于是 db/broker 现有 `trace.StartSpanFromContext(...)` + `ext.Component.Set(span,"mongo")` + `span.LogFields(...)`（全 opentracing API）**零改动**即进 OTel 后端。

### 4.1 正确接线（bridge 已知坑）

```go
bridge := otelbridge.NewBridgeTracer()
bridge.SetOpenTelemetryTracer(otelTracer)            // 包真实 OTel tracer
bridge.SetTextMapPropagator(compositePropagator())   // W3C+Jaeger+Baggage
wrappedProvider := otelbridge.NewTracerProvider(bridge, otelProvider)
otel.SetTracerProvider(wrappedProvider)              // 关键：wrapped provider
opentracing.SetGlobalTracer(bridge)                  // db/broker 全局经 bridge
```

- **必须**用 `NewTracerProvider(bridge, provider)` 包一层并设全局：bridge 的 `StartSpan`（`bridge.go:419`）默认在 `context.Background()` 上起 OTel span（断 parent）。wrapped provider 经 **deferred-setup** 把 OTel span 正确挂回调用方 ctx 的 OTel context key，保证 db/broker span 是中间件 server span 的**真子 span**、同 TraceID。
- `trace.TraceID(ctx)` 不再走 `jaeger.SpanContext` 断言（bridge span context 是 `bridgeSpanContext`，断言必失败）-> 改读 `oteltrace.SpanContextFromContext(ctx)`（OTel context key，bridge/wrapped provider 已写入）-> OTel 128-bit id。

### 4.2 三模式全局 opentracing tracer 与 db/broker 去向

| 模式 | 全局 opentracing tracer | db/broker span 去向 | OTel 后端 | jaeger 后端 |
|---|---|---|---|---|
| jaeger | jaeger tracer | jaeger | 无 | 全（server+client+db） |
| otel | **BridgeTracer**(otel) | **OTel**（经 bridge） | 全（server+client+db） | 无 |
| dual | **BridgeTracer**(otel) | **OTel**（经 bridge） | 全（server+client+db） | 仅 server+client |

dual 下 jaeger 后端缺 db/broker span（D6 取舍：OTel 是主观测后端、全覆盖；jaeger 侧仅并行核对）。

---

## 5. 目标架构

```
pkg/trace/
  tracer.go         # 【新】Tracer/Span/Mode 接口 + 全局 tracer + Init() 工厂 + auto-detect + noopTracer
  jaeger_tracer.go  # 【新】jaegerTracer（opentracing/jaeger，carrier 用 CarrierMD）
  otel_tracer.go    # 【新】otelTracer（OTel SDK + BridgeTracer 作全局 opentracing）
  dual_tracer.go    # 【新】dualTracer（composite：私有 jaeger tracer 起 server/client span + OTel/bridge 起 OTel span）
  trace.go          # 旧 API 降 facade 委托 globalTracer（签名不变；TraceID 改委托）
  metadata.go       # CarrierMD 保留（jaeger/facade 用）
pkg/trace/otel/otel.go        # Config 加 External；InitWithConfig 三态
pkg/trace/jaeger/jaeger.go    # Init() 委托 trace.Init()；暴露 newJaegerTracer 给工厂
pkg/server/http/middlewares/trace.go   # 删双分支，调 trace.Tracer() 抽象接口
pkg/server/rpc/interceptors/trace.go   # 同上
```

### 5.1 `Tracer` 接口（`pkg/trace/tracer.go`）

```go
type Mode string
const ( ModeJaeger Mode = "jaeger"; ModeOTel Mode = "otel"; ModeDual Mode = "dual" )

type SpanKind int
const ( SpanServer SpanKind = iota; SpanClient; SpanInternal )

// Span 跨后端 span 抽象（中间件用；db/broker 仍走旧 opentracing facade）
type Span interface {
	End()
	SetStatus(ok bool, msg string)
	SetStringAttr(key, val string)
	SetIntAttr(key string, val int64)
	TraceID() string
}

// Tracer 工厂选定的全局体系
type Tracer interface {
	Mode() Mode
	StartServerSpan(ctx, name string, kind SpanKind, carrier IncomingCarrier) (Span, context.Context)
	StartClientSpan(ctx, name string, kind SpanKind, carrier OutgoingCarrier) (Span, context.Context)
	StartSpan(ctx, name string) (Span, context.Context)          // root/业务（cron 等）
	SpanFromContext(ctx) Span
	TraceID(ctx) string
	Shutdown()
}

var globalTracer Tracer = noopTracer{}
func Tracer() Tracer { return globalTracer }
func Init() func() { /* 工厂 auto-detect，置 globalTracer */ }
```

- `IncomingCarrier`/`OutgoingCarrier`：HTTP header（`http.Header`）与 gRPC metadata（`metadata.MD`）的统一载体适配，各 Tracer 实现内部用 composite propagator extract/inject。

### 5.2 工厂 auto-detect（`Init()`）

```
hasJaeger := [trace.jaeger] 段存在（ServiceName != "" 判，沿用已修好的判据）
hasOtel   := [trace.otel] 段存在且 enable=true（endpoint 或 external 已声明）
switch {
case hasJaeger && hasOtel: globalTracer = newDualTracer(jconf, oconf)    // dual
case hasOtel:              globalTracer = newOtelTracer(oconf)            // pure otel
case hasJaeger:            globalTracer = newJaegerTracer(jconf)          // pure jaeger
default:                   globalTracer = newJaegerTracer(defaultCfg)     // 前向兼容
}
```
- otelTracer/dualTracer 的 OTel 部分按 §6 三态建 provider；建好后装 BridgeTracer + wrapped provider + `opentracing.SetGlobalTracer(bridge)`。
- jaegerTracer 内部 `opentracing.SetGlobalTracer(jaegerTracer)`（db/broker 进 jaeger）。

---

## 6. otel provider 三态（`pkg/trace/otel/otel.go`）

`Config` 加 `External bool`（toml `external`）：

| 配置 | 行为 |
|---|---|
| `Enable=false` | noop（现状） |
| `Enable=true` + `endpoint != ""` | 建 OTLP provider + 装 propagator（自建） |
| `Enable=true` + `endpoint == ""` + `External=true` | 只装 propagator + `enabled=true`，等外部 provider（业务侧 ADK `langfuse.Setup` 等） |
| `Enable=true` + `endpoint == ""` + `External=false` | **新增**：warn「otel enabled but no endpoint/external; OTel disabled to avoid zero TraceID」，不装 propagator、`enabled=false` -> 工厂回退不发 OTel span（**修复问题 2**） |

---

## 7. 旧 API facade（`pkg/trace/trace.go`，签名全保留）

| 旧 API | 改造 |
|---|---|
| `TraceID(ctx)` | -> `globalTracer.TraceID(ctx)`（**修复问题 1**：三模式返当前体系 id，纯 otel 不空） |
| `StartSpanFromContext(ctx, op, opts...)` | -> `opentracing.StartSpanFromContext`（走全局 opentracing：jaeger 模式=jaeger span；otel/dual=bridge span->OTel）。db/broker `ext.*` 三模式生效（**修复问题 3**） |
| `HeaderInjector`/`HeaderExtractor` | -> 委托 globalTracer carrier inject/extract |
| `MetadataInjector`/`FromIncomingContext` | -> 同上（dual/otel 经 composite propagator 同写 traceparent+uber-trace-id） |
| `SpanFromContext`/`ContextWithSpan`/`SetGlobalTracer`/`NullStartSpanOption`/`CarrierMD` | 保留（jaeger facade 内部 + 测试用） |

---

## 8. 中间件/拦截器改造

`pkg/server/http/middlewares/trace.go`、`pkg/server/rpc/interceptors/trace.go`：
- 删 `if otel.Enabled() { … }` 双分支。
- 改调 `trace.Tracer().StartServerSpan(ctx, name, SpanServer, carrier)` / `StartClientSpan(...)`（单一调用，工厂分发）。
- `X-Trace-Id` = `span.TraceID()`（统一，不再区分 ospan/jaeger）。
- 错误/状态码：`span.SetStatus(...)`/`SetIntAttr(...)`。
- dual 模式 `StartServerSpan` 内部同时起 jaeger+otel server span、同 TraceID -- 等价原双分支，封装进 dualTracer。
- 链注册（`server.go:60`/`rpc server.go:76`/`rpc client.go:68`）**不变**（中间件/拦截器签名不变）。

### 8.1 入口点

- 新入口 `trace.Init() func()`（工厂）。
- `jaeger.Init()` 改 `return trace.Init()`（服务 `defer jaeger.Init()()` 零改动）；`jaeger` 包不再读 `[trace.otel]`，只暴露 `newJaegerTracer` 给工厂。

---

## 9. 改动清单（文件级）

| 文件 | 动作 |
|---|---|
| `pkg/trace/tracer.go` | 新增：Tracer/Span/Mode 接口 + 全局 + Init() 工厂 + auto-detect + noopTracer |
| `pkg/trace/jaeger_tracer.go` | 新增：jaegerTracer |
| `pkg/trace/otel_tracer.go` | 新增：otelTracer（OTel + BridgeTracer + wrapped provider） |
| `pkg/trace/dual_tracer.go` | 新增：dualTracer（composite） |
| `pkg/trace/trace.go` | 改：旧 API 降 facade；`TraceID` 改委托 |
| `pkg/trace/otel/otel.go` | 改：Config 加 External；InitWithConfig 三态 |
| `pkg/trace/jaeger/jaeger.go` | 改：Init() 委托 trace.Init()；暴露 newJaegerTracer |
| `pkg/server/http/middlewares/trace.go` | 改：删双分支，调抽象接口 |
| `pkg/server/rpc/interceptors/trace.go` | 改：同上 |
| `go.mod` | 加 `go.opentelemetry.io/otel/bridge/opentracing v1.43.0`（**保持 grpc v1.80.0 锁定**；若 bridge 拖升 grpc 则降级评估） |
| `pkg/trace/trace_test.go`（新） | 三模式 + provider 三态 + bridge db/broker span + facade + carrier round-trip 单测 |

### 9.1 复用现有

- 复合 propagator（`otel.compositePropagator`，W3C+Jaeger+Baggage）-- otelTracer/dualTracer/bridge 用。
- `otel.MDCarrier`（gRPC↔OTel propagator）、`trace.CarrierMD`（gRPC↔opentracing）-- 各 Tracer 实现各用各的。
- `jaeger.defaultJaegerConfig`/`buildJaeger`/`newJaegerClient` -- jaegerTracer 复用。
- `otel.InitWithConfig` -- otelTracer 复用（加 External 三态后）。

---

## 10. 验证方案

1. **编译**：`go build ./...` + `go vet ./pkg/trace/... ./pkg/server/...`；确认 grpc 仍 v1.80.0（bridge 不应拖升；若拖升则降级 bridge 或评估）。
2. **单测**（`trace_test.go`）：
   - 三模式 `Init()` + server span：`trace.TraceID(ctx)` 非空且位宽正确（jaeger 16 hex / otel 32 hex / dual 取 otel）。
   - **bridge db/broker**：otel/dual 模式模拟 db 层 `trace.StartSpanFromContext` + `ext.Component.Set`，断言 OTel provider 收到该 span（test exporter 验证），且 TraceID = 父 server span TraceID（parent 关系正确，验证 wrapped provider deferred-setup 接线对）。
   - otel provider 三态：endpoint 非空 -> TraceID 非零；endpoint 空+external -> propagator 装好待外部；endpoint 空+无 external -> OTel 不启用退化。
   - facade：jaeger 模式 `trace.StartSpanFromContext` 返 jaeger span，`ext.*` 可标记。
   - carrier round-trip：HTTP header / gRPC metadata inject->extract 同 TraceID（三模式）。
3. **回归**：中间件链顺序、`X-Trace-Id`、gRPC peer 属性、错误状态标记不变。
4. **前向兼容**：不配 trace 段的老服务 -> 默认 jaeger（hostname），`trace.TraceID` 返 jaeger id，行为同改前。

---

## 11. 风险与回滚

| 风险 | 说明 | 缓解 |
|---|---|---|
| bridge 拖升 grpc | `bridge/opentracing v1.43.0` 可能要求 grpc >= v1.81.1 | 实测 `go get` 后查；若拖升，评估降级 bridge 版本或接受（违背 grpc 锁定则否决） |
| bridge deferred-setup 接线错 | db/broker span 断 parent、TraceID 不续 | 单测断言 parent 关系 + TraceID 一致 |
| dual 下 jaeger 缺 db/broker span | db/broker 走 OTel 侧 | D6 取舍，文档标注；后续 composite tracer 补 |
| `trace.TraceID` 语义变 | 双轨模式从 jaeger id 变 OTel id | 设计目标（统一对外 id）；review 确认无下游强依赖 jaeger id |
| 全局 opentracing tracer 被替换 | jaeger 模式不变；otel/dual 变 bridge | facade 签名不变，调用点零改；单测覆盖三模式 |

**回滚**：`git revert` 本次提交 -> 回到方案 W 合并版（`jaeger.Init` 双分支）。服务侧零改动（入口签名不变）。

---

## 12. 决策点（review 时请明确）

1. **D1~D6 是否照此**？尤其 D5（引入 `bridge/opentracing` 新依赖）、D6（dual 下 jaeger 缺 db/broker span）。
2. **grpc 锁定**：若 `bridge/opentracing v1.43.0` 拖升 grpc，接受升 grpc 还是降级 bridge？（倾向：保 grpc v1.80.0，降级 bridge 到不拖升的版本，接受 bridge API 差异）。
3. **`trace.TraceID` 语义变更**：双轨模式对外 id 从 jaeger 64-bit 改为 OTel 128-bit -- 是否有下游（访问日志、外部消费方）强依赖原 jaeger id？
4. **是否本次启动编码**：review 通过后是否进入编码？框架改动是否一次到位（含单测）？

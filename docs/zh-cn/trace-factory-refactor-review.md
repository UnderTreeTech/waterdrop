# waterdrop trace 工厂化重构 — Review 与验证总结

> 本文是 `trace-factory-refactor.md` 落地后的 **review 结论 + 验证记录**，作为多版本共存升级（老服务 jaeger / 新服务 otel 共存）的可追查依据。
>
> 配套：设计文档 `docs/zh-cn/trace-factory-refactor.md`；测试 `pkg/trace/trace_test.go`、`e2e_test.go`、`dbbroker_test.go`。
>
> 日期：2026-07-15。

---

## 0. 结论先行

**老服务只升 waterdrop 版本、不改代码，trace 行为无不兼容；新老服务（jaeger ↔ otel）跨体系请求 trace 可串联。可放心放出多版本共存升级。**

- API 签名、opentracing 语义、jaeger tracer 构建、配置读取、TraceID 返回值、grpc 依赖 —— 全部与改前一致或等价。
- 新老跨体系串联靠复合 propagator（`traceparent` + `uber-trace-id`），实测 jaeger↔otel 四种组合都成一条 trace。
- db/broker 层（mongo/sql/redis/es/rocketmq）`StartSpanFromContext` + `ext.*` 在三模式（jaeger/otel/dual）下均正常工作、span 进后端、TraceID 与 server span 同 trace。

---

## 1. 前向兼容（老服务零改动）验证

### 1.1 API 签名全部保留

| 符号 | 状态 |
|---|---|
| `jaeger.Init() func()` | ✅ 签名不变，内部委托 `trace.Init()` |
| `trace.StartSpanFromContext` / `SpanFromContext` / `ContextWithSpan` | ✅ 一行委托 opentracing，逐字未改 |
| `trace.FromIncomingContext` / `MetadataInjector` / `HeaderInjector` / `HeaderExtractor` | ✅ 逐字未改 |
| `trace.SetGlobalTracer` / `trace.TraceID` / `NullStartSpanOption` / `CarrierMD` | ✅ 保留 |
| `otel.Enabled` / `InitWithConfig` / `Propagator` / `Tracer` / `OtelTraceID` / `SpanContextFromContext` / `ContextWithSpanContext` | ✅ 保留 |

### 1.2 行为逐字一致

老服务只配 `[trace.jaeger]` -> 工厂 `case hasJaeger` -> `newJaegerTracer`：
- 构建 `jconfig.Configuration`（ServiceName/Sampler/Reporter/RPCMetrics + MaxTagValueLength）与改前 `buildJaeger` 字段一一对应；`BufferFlushInterval` 零值透传（jaeger-client-go 自身兜底默认 1s，与改前字面一致）。
- `cfg.NewTracer()` + `opentracing.SetGlobalTracer(tracer)` —— 与改前一致。
- db/broker 的 `StartSpanFromContext` + `ext.*` 走全局 opentracing（jaeger），行为不变。

### 1.3 `trace.TraceID` 在 jaeger 模式下无差异

改前 `sp.Context().(jaeger.SpanContext)` 断言；改后委托 `globalTracer.TraceID`，jaeger 模式下 `jaegerTracer.TraceID` 同样是 `jaeger.SpanContext` 断言 —— 老服务返回值一致（64-bit jaeger id）。

唯一**增强**（非破坏）：otel/dual 模式 `TraceID` 从"空/全零"变 OTel 真实 id。老服务不配 otel -> 不触发，零影响。

### 1.4 配置读取兼容

`readJaegerConfig` 用 `ServiceName != ""` 判存在；`readOtelConfig` 用 `oconf.Active()`（enable + endpoint/external）。老服务无 otel 段 -> `Active()=false` -> 不进 otel/dual 分支。

### 1.5 依赖锁定

`google.golang.org/grpc v1.80.0`（锁定，注释 `// pinned: do not bump; OTel deps must stay == v1.43.0`）；`otel` 全家桶 `v1.43.0`、`bridge/opentracing v1.43.0`、`contrib/propagators/jaeger v1.43.0`。`go mod tidy` 重跑后版本稳定（不被顶到 v1.44.0）。老服务升级后 `tidy` 不会被动升 grpc。

### 1.6 实测

`TestForwardCompatNoTraceSection`（无 trace 段）、`TestForwardCompatJaegerOnly`（只配 jaeger）—— `Mode=jaeger`、TraceID 16-hex 非空，行为同改前。

### 1.7 唯一非破坏性变化（需知悉）

`pkg/trace/jaeger/jaeger.go` 里原 `Config`/`buildJaeger`/`newJaegerClient`/`defaultJaegerConfig`/`JaegerConfig`/`WithOption` 等类型/函数**移至** `pkg/trace/jaeger_tracer.go`（包内迁移，为打破 `pkg/trace` ↔ `pkg/trace/jaeger` 循环依赖）。若老服务**直接 import `pkg/trace/jaeger` 用这些类型**（非仅调 `Init`）会编译失败。全仓 grep 确认 waterdrop 仓内无此类引用，服务层只调 `jaeger.Init()`，实际无影响。

---

## 2. 跨体系串联（jaeger ↔ otel）验证

复合 propagator（`W3C TraceContext` + `Jaeger uber-trace-id` + `Baggage`）让两套体系互通：

| 方向 | 出站注入 | 入站提取 | TraceID 关系 |
|---|---|---|---|
| jaeger -> jaeger | `uber-trace-id`（64-bit） | jaeger Extract | 同（16 hex） |
| jaeger -> otel | `uber-trace-id`（64-bit） | otel 复合 propagator Extract，64-bit 左补零到 128-bit | server = `0000…` + client（32 hex） |
| otel -> jaeger | `traceparent` + `uber-trace-id`（128-bit） | jaeger Extract `uber-trace-id`，jaeger 承载 128-bit | 同（32 hex） |
| otel -> otel | `traceparent` + `uber-trace-id`（128-bit） | otel 复合 propagator Extract | 同（32 hex） |

**实测**（`TestE2EHTTP`、`TestE2EGRPC`、`TestCrossJaegerToOtel`、`TestCrossOtelToJaeger`）：四组合全过，client/server TraceID 经左补零归一化后相等。

### 2.1 跨体系检索约束（固有限制，非 bug）

jaeger -> otel 方向，新服务（otel）侧 TraceID 是**左补零的 128-bit**（`0000000000000000` + 老 64-bit）。拿老服务 jaeger 的 64-bit id 直接去 langfuse 搜**搜不到**（langfuse 是 32-hex 补零版），需补零后再搜。这是 64/128 位宽差异的固有限制。

---

## 3. db/broker 层覆盖验证（核心诉求）

db/broker 层（mongo/sql/redis/es/rocketmq）用 opentracing API：`trace.StartSpanFromContext` + `ext.PeerAddress/DBType/Component/SpanKind/DBInstance.Set` + `span.SetTag` + `span.LogFields`（slow_query）+ `ext.Error.Set`。测试 `pkg/trace/dbbroker_test.go` 在三模式（jaeger/otel/dual）下覆盖：

| 测试 | 覆盖 | jaeger | otel | dual |
|---|---|---|---|---|
| `TestDBBrokerSpanBasics` | mongo aggregate 模式：ext.* 全 tag + SetTag + LogFields + ext.Error | ✅ | ✅ | ✅ |
| `TestDBBrokerRedisPattern` | redis hook：Component=redis + SpanKind=client | ✅ | ✅ | ✅ |
| `TestDBBrokerSQLTxPattern` | sql Tx：span 持有于结构体 + 嵌套 exec 子 span | ✅ | ✅ | ✅ |
| `TestDBBrokerRocketMQPattern` | rocketmq consumer：FromIncomingContext + Component=MQ | ✅ | ✅ | ✅ |
| `TestDBBrokerTraceIDInLogContext` | log.go 的 `trace.TraceID(ctx)` 在 db op 内非空 | ✅ | ✅ | ✅ |
| `TestDBBrokerSpanFromContextFacade` | `SpanFromContext`（injector 依赖）非 nil | ✅ | ✅ | ✅ |
| `TestDBBrokerNilContextSafe` | 裸 `context.Background()` 起 span 不 panic | ✅ | ✅ | ✅ |

otel/dual 模式额外断言（in-memory exporter）：db span 被记录、attributes 齐全（peer.address/db.type/db.instance/db.collection/component）、parent TraceID == server TraceID、LogFields 转 otel event、ext.Error 转 otel error status。

### 3.1 已知 bridge 限制（不影响串联，仅 cosmetic）

otel/dual 模式下，db/broker 用 `ext.SpanKind.Set(span, ...)` 在 span **创建后**设置 span kind。bridge 的 `bridgeSpan.SetTag` 对 `span.kind` 有 `// TODO: Should we ignore it?` —— **忽略**创建后的 span.kind 设置。因此 db span 的 otel `SpanKind` 保持 `Internal`（默认），不会变成 client/consumer。同理 `ext.Error.Set` 被映射成 otel `Status(codes.Error)` 而非 `error` 属性（这是正确映射，非问题）。

**影响**：otel 后端里 db span 的 `span.kind` 显示为 internal 而非 client。**不影响**：trace 串联、TraceID、日志 trace_id、其他 tag、span 记录。属 bridge 上游已知 TODO，非本次引入，可后续在 bridge 层或 db/broker 改用创建时传 kind 的方式修复（非阻塞）。

### 3.2 otel 模式裸 ctx root span 的 TraceID

otel 模式下，`StartSpanFromContext(context.Background(), ...)`（无 server span 上游）起的 root db span：span **会被记录**到 otel 后端（bridge），但 `trace.TraceID(ctx)` 返回 `""`（bridge span 不在 OTel context key 上，无 server span 时 otel key 为空）。这是 bridge 固有行为。真实服务里 db op 总在请求 handler 内（有 server span 在 otel key），`TraceID` 正常；仅 cron/后台裸 ctx 的 standalone db span 会 TraceID 空，但 span 仍记录。

---

## 4. 三模式工厂行为验证

`TestFactoryJaegerMode` / `TestFactoryOTelMode` / `TestFactoryDualMode` / `TestFactoryOTelDisabledFallback` / `TestFactoryOTelBridgeDBSpan`：

- 自动检测选对 Mode；TraceID 位宽正确（jaeger 16 hex / otel·dual 32 hex）。
- otel provider 三态：endpoint 自建 / external 等外部 / 都无则退化（防全零 id）。
- bridge db/broker span 实测进 OTel exporter 且 parent 正确（`TestFactoryOTelBridgeDBSpan`）。

---

## 5. 依赖与锁定

| 模块 | 版本 |
|---|---|
| `google.golang.org/grpc` | v1.80.0（锁定，不升） |
| `go.opentelemetry.io/otel` / trace / metric / sdk | v1.43.0 |
| `go.opentelemetry.io/otel/bridge/opentracing` | v1.43.0 |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` | v1.43.0 |
| `go.opentelemetry.io/contrib/propagators/jaeger` | v1.43.0 |

otel 锁定是软锁（go.mod require + 注释），依赖"无依赖 require otel v1.44.0"。当前已确认无。未来若新增依赖拖入 v1.44.0，`tidy` 会顶上去（进而拖 grpc 到 v1.81.1）—— 注释已警示，CI 可加校验。

---

## 6. 测试清单（可追查）

`pkg/trace/`：

| 文件 | 测试 | 覆盖 |
|---|---|---|
| `trace_test.go` | `TestFactoryJaegerMode`/`OTelMode`/`OTelBridgeDBSpan`/`DualMode`/`OTelDisabledFallback` | 工厂三模式 + provider 三态 + bridge |
| | `TestCarrierRoundTripHTTP` | HTTP carrier inject/extract |
| | `TestJaegerFacadeStartSpan` | jaeger facade ext.* |
| | `TestCrossJaegerToOtel`/`TestCrossOtelToJaeger` | 跨体系双向 |
| | `TestForwardCompatNoTraceSection`/`JaegerOnly` | 前向兼容 |
| `e2e_test.go` | `TestE2EHTTP`（4 组合）/ `TestE2EGRPC`（4 组合） | client->server 端到端串联 + trace_id 打印 |
| | `TestE2ELogTraceID` | log trace_id 非空 |
| `dbbroker_test.go` | `TestDBBrokerSpanBasics`/`RedisPattern`/`SQLTxPattern`/`RocketMQPattern`/`TraceIDInLogContext`/`SpanFromContextFacade`/`NilContextSafe`（各 3 模式） | db/broker 全覆盖 |

跑法（看 trace_id）：
```bash
cd <waterdrop>
go test ./pkg/trace/ -run 'TestE2E|TestDBBroker|TestCross' -v
```

---

## 7. 升级路径（多版本共存）

1. waterdrop 提交本次改动 + 发新伪版本。
2. **老服务**：升 waterdrop 版本 -> `go mod tidy` -> 重新编译。**不改代码、不改配置**。trace 行为同改前（jaeger）。
3. **新服务**：升 waterdrop 版本 -> 配 `[trace.otel]`（endpoint 自建 或 external=true 复用 adk provider）-> 编译。
4. 共存期间：老服务（jaeger）↔ 新服务（otel）请求 trace 可串联（复合 propagator）。
5. 长期：老服务逐步配 `[trace.otel]` 切 otel；最终全 otel 后可下线 jaeger。

---

## 8. 未决 / 后续

- bridge 的 post-creation `span.kind` TODO（§3.1）：otel/dual 下 db span kind 显示 internal。非阻塞，后续修复。
- otel 软锁 v1.43.0：CI 加 `tidy` 后版本校验，防被动升 grpc。
- `TestDefaultCORS` 预存失败（与 trace 无关，未改动 cors.go/cors_test.go）。

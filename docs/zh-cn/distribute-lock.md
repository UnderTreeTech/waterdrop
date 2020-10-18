## Distribute Lock

### Redis Lock
适合最终一致的业务锁，代码路径database/redis/lock.go
```go
Lock
UnLock
ForceUnlock
```

### Etcd lock
适合强一致的业务锁，代码路径registry/etcd/lock.go
```go
NewMutex
Lock
TryLock
UnLock
```
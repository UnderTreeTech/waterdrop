## xbuffer

1. BufferPool
Based on sync.Pool, sync.Pool automatically GC's objects.

2. SizedBufferPool
Based channel, creates a new BufferPool bounded to the given size. objects in this pool are 
alive along with the application life cycle.
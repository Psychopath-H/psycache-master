package psycache

// Cache 模块负责提供对缓存算法模块的并发控制

import (
	"errors"
	cacheAlg "github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm"
	"github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm/fifo"
	"github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm/lfu"
	"github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm/lru"
	"github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm/lruk"
	"github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm/twoQ"
	"sync"
)

// 这样设计可以进行cache和算法的分离，比如我现在实现了lfu缓存模块
// 只需替换cache成员即可
type cache struct {
	mu             sync.Mutex
	specificCache  Cache
	capacity       int64 // 缓存最大容量
	expirationTime int64
}

func newLRUCache(capacity int64, callback cacheAlg.OnEliminated) *cache {
	return &cache{
		capacity:      capacity,
		specificCache: lru.New(capacity, callback),
	}
}

func newLFUCache(capacity int64, callback cacheAlg.OnEliminated) *cache {
	return &cache{
		capacity:      capacity,
		specificCache: lfu.New(capacity, callback),
	}
}

func newLRUKCache(capacity int64, k int, callback cacheAlg.OnEliminated) *cache {
	return &cache{
		capacity:      capacity,
		specificCache: lruk.New(capacity, k, callback),
	}
}

func newFIFOCache(capacity int64, callback cacheAlg.OnEliminated) *cache {
	return &cache{
		capacity:      capacity,
		specificCache: fifo.New(capacity, callback),
	}
}

func newtwoQCache(capacity int64, callback cacheAlg.OnEliminated) *cache {
	return &cache{
		capacity:      capacity,
		specificCache: twoQ.New(capacity, callback),
	}
}

func (c *cache) add(key string, value ByteView, expiretionTime int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.specificCache == nil {
		return errors.New("you should build cache first")
	}
	c.specificCache.Add(key, value, expiretionTime)
	return nil
}

func (c *cache) get(key string) (ByteView, bool) {
	if c.specificCache == nil {
		return ByteView{}, false
	}
	// 注意：Get操作需要修改lru中的双向链表，需要使用互斥锁。
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.specificCache.Get(key); ok {
		return v.(ByteView), true
	}
	return ByteView{}, false
}

func (c *cache) remove(key string) bool {
	if c.specificCache == nil {
		return false
	}
	// 注意：remove操作需要修改lru中的双向链表，需要使用互斥锁。
	c.mu.Lock()
	defer c.mu.Unlock()
	if ok := c.specificCache.Remove(key); ok {
		return true
	}
	return false
}

func (c *cache) contains(key string) bool {
	if c.specificCache == nil {
		return false
	}
	if ok := c.specificCache.Contains(key); ok {
		return true
	}
	return false
}

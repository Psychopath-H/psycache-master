package twoQ

import (
	cahce "github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm"
	"github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm/fifo"
	"github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm/lru"
)

// Cache 是LRU算法实现的缓存
// 使用lru+FIFO实现LRU
type TwoQCache struct {
	capacity int64 // Cache 最大容量(Byte)
	lru      *lru.LRUCache
	FIFO     *fifo.FIFOCache
}

func New(maxBytes int64, callback cahce.OnEliminated) *TwoQCache {
	lru := lru.New(maxBytes, callback)
	FIFO := fifo.New(maxBytes, callback)
	return &TwoQCache{
		capacity: maxBytes,
		lru:      lru,
		FIFO:     FIFO,
	}
}

// Get 在2Q缓存中获取对应key的value
func (c *TwoQCache) Get(key string) (value cahce.Lengthable, ok bool) {
	//能在lru中找到
	if value, ok = c.lru.Get(key); ok {
		return value, ok
	}
	//在lru中找不到，但能在FIFO中找到
	if value, expirationTime, ok := c.FIFO.Peek(key); ok {
		c.lru.Add(key, value, expirationTime)
		c.FIFO.Remove(key)
		return value, ok
	}
	return nil, false
}

// Add 向缓存中添加指定key的value
func (c *TwoQCache) Add(key string, value cahce.Lengthable, expirationTime int64) {
	if c.lru.Contains(key) {
		c.lru.Add(key, value, expirationTime)
	} else if ok := c.FIFO.Contains(key); ok {
		c.lru.Add(key, value, expirationTime)
		c.FIFO.Remove(key)
	} else {
		c.FIFO.Add(key, value, expirationTime)
	}
	return
}

// Remove 从缓存中移除提供的键
func (c *TwoQCache) Remove(key string) (ok bool) {
	if c.lru.Contains(key) {
		ok := c.lru.Remove(key)
		return ok
	}

	if c.FIFO.Contains(key) {
		ok := c.FIFO.Remove(key)
		return ok
	}
	return false
}

// Contains 检查某个键是否在缓存中，但不更新缓存的状态
func (c *TwoQCache) Contains(key string) (ok bool) {
	if c.lru.Contains(key) || c.FIFO.Contains(key) {
		return true
	}
	return false
}

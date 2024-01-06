package lruk

import (
	cache "github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm"
	"github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm/lru"
)

// LRUKCache 是LRU算法实现的缓存
// 使用两个LRU搭配上historyVisited表记录被访问次数实现LRU-K算法
type LRUKCache struct {
	capacity       int64 // Cache 最大容量(Byte)
	k              int   //进入缓存队列的次数标准
	datalru        *lru.LRUCache
	historylru     *lru.LRUCache
	historyVisited map[string]int
}

// New 创建指定最大容量的LRU缓存。
// 当maxBytes为0时，代表cache无内存限制，无限存放。
func New(maxBytes int64, visitedNum int, callback cache.OnEliminated) *LRUKCache {
	datalru := lru.New(maxBytes, callback)
	historylru := lru.New(maxBytes, callback)
	return &LRUKCache{
		capacity:       maxBytes,
		k:              visitedNum,
		datalru:        datalru,
		historylru:     historylru,
		historyVisited: make(map[string]int),
	}
}

// Get 从缓存获取对应key的value。
// ok 指明查询结果 false代表查无此key
func (c *LRUKCache) Get(key string) (value cache.Lengthable, ok bool) {
	if value, ok = c.datalru.Get(key); ok { //能在数据缓存中找到
		c.historyVisited[key]++
		return value, ok
	} else if c.historylru.Contains(key) && c.historyVisited[key] == c.k-1 { //不能在数据缓存中找到，但在历史缓存中找到且达到了访问次数阈值
		value, expiretionTime, _ := c.historylru.Peek(key)
		c.datalru.Add(key, value, expiretionTime)
		c.historylru.Remove(key)
		c.historyVisited[key]++
		return value, true
	} else if c.historylru.Contains(key) && c.historyVisited[key] < c.k-1 { //不能在数据缓存中找到，但在历史缓存中找到但未达到访问次数阈值
		value, _ := c.historylru.Get(key)
		c.historyVisited[key]++
		return value, true
	} else if !c.historylru.Contains(key) { //历史缓存中也找不到
		delete(c.historyVisited, key)
		return nil, false
	}
	return nil, false
}

// Add 向缓存中添加指定key的value
func (c *LRUKCache) Add(key string, value cache.Lengthable, expirationTime int64) {
	if c.datalru.Contains(key) { //能在数据缓存中找到，更新一下信息
		c.datalru.Add(key, value, expirationTime)
		c.historyVisited[key]++
	} else if c.historylru.Contains(key) && c.historyVisited[key] == c.k-1 { //不能在数据缓存中找到，但在历史缓存中找到且达到了访问次数阈值
		c.datalru.Add(key, value, expirationTime)
		c.historylru.Remove(key)
		c.historyVisited[key]++
	} else if c.historylru.Contains(key) && c.historyVisited[key] < c.k-1 { //不能在数据缓存中找到，但在历史缓存中找到但未达到访问次数阈值
		c.historylru.Add(key, value, expirationTime)
		c.historyVisited[key]++
	} else { //无法在历史缓存中找到
		c.historylru.Add(key, value, expirationTime)
		c.historyVisited[key] = 1
	}
}

// Remove 从缓存中移除提供的键
func (c *LRUKCache) Remove(key string) (ok bool) {
	if c.datalru.Contains(key) {
		ok := c.datalru.Remove(key)
		delete(c.historyVisited, key)
		return ok
	}
	if c.historylru.Contains(key) {
		ok := c.historylru.Remove(key)
		delete(c.historyVisited, key)
		return ok
	}
	delete(c.historyVisited, key)
	return false
}

// Contains 检查某个键是否在缓存中，但不更新缓存的状态
func (c *LRUKCache) Contains(key string) (ok bool) {
	if c.datalru.Contains(key) || c.historylru.Contains(key) {
		return true
	}
	delete(c.historyVisited, key)
	return false
}

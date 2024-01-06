package lru

import (
	"container/list"
	cache "github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm"
	"time"
)

// entry 定义双向链表节点所存储的对象
type entry struct {
	key            string
	value          cache.Lengthable
	expirationTime int64
}

// LRUCache 是LRU算法实现的缓存
type LRUCache struct {
	capacity         int64 // Cache 最大容量(Byte)
	nowcap           int64 // Cache 当前容量(Byte)
	hashmap          map[string]*list.Element
	doublyLinkedList *list.List // 链头表示最近使用

	callback cache.OnEliminated
}

// New 创建指定最大容量的LRU缓存。
// 当maxBytes为0时，代表cache无内存限制，无限存放。
func New(maxBytes int64, callback cache.OnEliminated) *LRUCache {
	return &LRUCache{
		capacity:         maxBytes,
		hashmap:          make(map[string]*list.Element),
		doublyLinkedList: list.New(),
		callback:         callback,
	}
}

// Get 从缓存获取对应key的value。
// ok 指明查询结果 false代表查无此key
func (c *LRUCache) Get(key string) (value cache.Lengthable, ok bool) {
	if elem, ok := c.hashmap[key]; ok {
		if checkExpirationTime(elem.Value.(*entry).expirationTime) {
			c.removeElement(elem)
			return nil, false
		}
		c.doublyLinkedList.MoveToFront(elem)
		kv := elem.Value.(*entry)
		return kv.value, true
	}
	return
}

// Add 向缓存中添加指定key的value
func (c *LRUCache) Add(key string, value cache.Lengthable, expirationTime int64) {
	kvSize := int64(len(key)) + int64(value.Len())
	// cache 容量检查
	for c.capacity != 0 && c.nowcap+kvSize > c.capacity {
		c.RemoveOldest()
	}
	if kvSize > c.capacity { //新加的这个比总容量都要大了
		return
	}
	if elem, ok := c.hashmap[key]; ok {
		// 更新缓存key值
		c.doublyLinkedList.MoveToFront(elem)
		oldEntry := elem.Value.(*entry)
		// 先更新写入字节 再更新
		c.nowcap += int64(value.Len()) - int64(oldEntry.value.Len())
		oldEntry.value = value
		oldEntry.expirationTime = expirationTime
	} else {
		// 新增缓存key
		elem := c.doublyLinkedList.PushFront(&entry{key: key, value: value, expirationTime: expirationTime})
		c.hashmap[key] = elem
		c.nowcap += kvSize
	}
}

// Remove 从缓存中移除提供的键
func (c *LRUCache) Remove(key string) (ok bool) {
	if e, exist := c.hashmap[key]; exist {
		c.removeElement(e)
		return exist
	}
	return false
}

// Contains 检查某个键是否在缓存中，但不更新缓存的状态
func (c *LRUCache) Contains(key string) (ok bool) {
	e, ok := c.hashmap[key]
	if ok {
		// 判断此值是否已经超时,如果超时则进行删除
		if checkExpirationTime(e.Value.(*entry).expirationTime) {
			c.removeElement(e)
			return !ok
		}
	}
	return ok
}

// Peek 在不更新的情况下返回键值(如果没有找到则返回false)，不更新缓存的状态
func (c *LRUCache) Peek(key string) (value cache.Lengthable, expirationTime int64, ok bool) {
	if e, ok := c.hashmap[key]; ok {
		kv := e.Value.(*entry)
		if checkExpirationTime(kv.expirationTime) {
			c.removeElement(e)
			return nil, 0, false
		}
		return kv.value, kv.expirationTime, true
	}
	return nil, 0, ok
}

// removeOldest 淘汰一枚最近最不常用缓存
func (c *LRUCache) RemoveOldest() {
	tailElem := c.doublyLinkedList.Back()
	if tailElem != nil {
		kv := tailElem.Value.(*entry)
		k, v := kv.key, kv.value
		delete(c.hashmap, k)                       // 移除映射
		c.doublyLinkedList.Remove(tailElem)        // 移除缓存
		c.nowcap -= int64(len(k)) + int64(v.Len()) // 更新占用内存情况
		// 移除后的善后处理
		if c.callback != nil {
			c.callback(k, v)
		}
	}
}

// removeElement 删除一枚指定的元素
func (c *LRUCache) removeElement(e *list.Element) {
	c.nowcap -= int64(len(e.Value.(*entry).key) + e.Value.(*entry).value.Len())
	c.doublyLinkedList.Remove(e)
	delete(c.hashmap, e.Value.(*entry).key)
	if c.callback != nil {
		c.callback(e.Value.(*entry).key, e.Value.(*entry).value)
	}
}

// Len 获取缓存的长度
func (c *LRUCache) Len() int {
	return len(c.hashmap)
}

func checkExpirationTime(expirationTime int64) (ok bool) {
	if 0 != expirationTime && expirationTime <= time.Now().UnixNano()/1e6 {
		return true
	}
	return false
}

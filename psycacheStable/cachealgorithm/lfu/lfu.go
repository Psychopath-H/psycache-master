package lfu

import (
	"container/list"
	cache "github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm"
	"time"
)

type LFUCache struct {
	capacity int64
	nowcap   int64
	minFre   int
	kItems   map[string]*list.Element
	fItems   map[int]*list.List //频率链表

	callback cache.OnEliminated
}

type entry struct {
	key            string
	value          cache.Lengthable
	freq           int
	expirationTime int64
}

func New(capacity int64, callback cache.OnEliminated) *LFUCache {
	lfu := &LFUCache{
		kItems:   make(map[string]*list.Element),
		fItems:   make(map[int]*list.List),
		capacity: capacity,
		nowcap:   0,
		minFre:   1,
		callback: callback,
	}
	lfu.fItems[lfu.minFre] = list.New() //这个频率是一定会用到的，提前申请好
	return lfu
}

func (c *LFUCache) Get(key string) (value cache.Lengthable, ok bool) {
	if len(c.kItems) == 0 {
		return
	}
	if node, exist := c.kItems[key]; exist {
		if checkExpirationTime(node.Value.(*entry).expirationTime) {
			c.removeElement(node)
			return nil, false
		}
		value = node.Value.(*entry).value
		c.nodeExec(node) //进行更新
		return value, true
	}
	return nil, false
}

func (c *LFUCache) Add(key string, value cache.Lengthable, expirationTime int64) {
	if c.capacity <= 0 {
		return
	}
	kvSize := int64(len(key)) + int64(value.Len()) //新增键值对的长度
	//该键值已经存在
	if node, ok := c.kItems[key]; ok {
		for c.nowcap+int64(value.Len())-int64(node.Value.(*entry).value.Len()) > c.capacity {
			c.RemoveOldest()
		}
		c.nowcap += int64(value.Len()) - int64(node.Value.(*entry).value.Len())
		node.Value.(*entry).value = value
		node.Value.(*entry).expirationTime = expirationTime
		c.nodeExec(node)
		return
	}
	//该键值不存在
	for c.nowcap+kvSize > c.capacity { //挤出空间
		c.RemoveOldest()
	}
	if kvSize > c.capacity { //新加的这个比总容量都要大了
		return
	}

	kv := &entry{key: key, value: value, freq: 1, expirationTime: expirationTime}
	node := c.fItems[kv.freq].PushFront(kv)
	c.nowcap += kvSize
	c.kItems[key] = node
	c.minFre = 1
	return
}

// Remove 删除指定键的缓存
func (c *LFUCache) Remove(key string) (ok bool) {
	if e, ok := c.kItems[key]; ok {
		c.removeElement(e)
		return ok
	}
	return false
}

// Contains 检查某个键是否在缓存中，但不更新缓存的状态
func (c *LFUCache) Contains(key string) (ok bool) {
	e, ok := c.kItems[key]
	if ok {
		if checkExpirationTime(e.Value.(*entry).expirationTime) {
			c.removeElement(e)
			return !ok
		}
	}
	return ok
}

// Peek 在不更新的情况下返回键值(如果没有找到则返回false)，不更新缓存的状态
func (c *LFUCache) Peek(key string) (value cache.Lengthable, expirationTime int64, ok bool) {
	if e, ok := c.kItems[key]; ok {
		kv := e.Value.(*entry)
		if checkExpirationTime(kv.expirationTime) {
			c.removeElement(e)
			return nil, 0, false
		}
		return kv.value, kv.expirationTime, true
	}
	return nil, 0, ok
}

// removeOldest 从缓存中移除最老的项
func (c *LFUCache) RemoveOldest() {
	l := c.fItems[c.minFre] //找到最小使用频率的双向链表
	tailnode := l.Back()    //找到双向链表的最后一个(最小频率使用且最长时间未使用)
	if tailnode != nil {
		kv := tailnode.Value.(*entry)                 //找到那条记录
		delete(c.kItems, tailnode.Value.(*entry).key) //移除映射
		l.Remove(tailnode)                            //在频率链表里也移除
		kvSize := int64(len(kv.key) + kv.value.Len())
		c.nowcap -= kvSize
		//移除后的善后处理
		if c.callback != nil {
			c.callback(kv.key, kv.value)
		}
	}
}

// removeElement 删除缓存中的指定元素
func (c *LFUCache) removeElement(node *list.Element) {
	kv := node.Value.(*entry)
	c.nowcap -= int64(len(kv.key) + kv.value.Len())
	l := c.fItems[kv.freq] //找到是哪条频率链表
	l.Remove(node)         //在频率链表里删除

	//看是否有必要更新一下频率链表
	i := 1
	for c.fItems[i] != nil && c.fItems[i].Len() == 0 {
		i++
	}
	switch {
	case c.fItems[i] == nil:
		c.minFre = 1
	case c.fItems[i].Len() >= 0:
		c.minFre = i
	}

	delete(c.kItems, kv.key) //删除映射
	//移除后的善后处理
	if c.callback != nil {
		c.callback(kv.key, kv.value)
	}
}

// nodeExec将频率链表进行更新，并对minfrep进行更新
func (c *LFUCache) nodeExec(node *list.Element) {
	//原频率中删除
	kv := node.Value.(*entry)
	oldList := c.fItems[kv.freq]
	oldList.Remove(node)

	//更新minfreq
	if oldList.Len() == 0 && c.minFre == kv.freq {
		c.minFre++
	}

	//放入新的频率链表
	kv.freq++
	if _, ok := c.fItems[kv.freq]; !ok { //新的频率链表没有的话，就新建一个
		c.fItems[kv.freq] = list.New()
	}
	newList := c.fItems[kv.freq]
	node = newList.PushFront(kv)
	c.kItems[kv.key] = node //建立entry的key到频率节点的映射
}

// Len the number of cache entries
func (c *LFUCache) Len() int {
	return len(c.kItems)
}

func checkExpirationTime(expirationTime int64) (ok bool) {
	if 0 != expirationTime && expirationTime <= time.Now().UnixNano()/1e6 {
		return true
	}
	return false
}

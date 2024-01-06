package psycache

import (
	"fmt"
	"github.com/Psychopath-H/psycache-master/psycacheStable/singlefilght"
	"log"
	"sync"
)

const (
	TYPE_FIFO = "FIFO"
	TYPE_LRU  = "lru"
	TYPE_LFU  = "lfu"
	TYPE_LRUK = "lruk"
	TYPE_2Q   = "twoQ"
)

// psycache 模块提供比cache模块更高一层抽象的能力
// 换句话说，实现了填充缓存/命名划分缓存的能力

var (
	mu     sync.RWMutex // 管理读写groups并发控制
	groups = make(map[string]*Group)
)

// Retriever 要求对象实现从数据源获取数据的能力
type Retriever interface {
	retrieve(string) ([]byte, error)
}

type RetrieverFunc func(key string) ([]byte, error)

// RetrieverFunc 通过实现retrieve方法，使得任意匿名函数func
// 通过被RetrieverFunc(func)类型强制转换后，实现了 Retriever 接口的能力
func (f RetrieverFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}

// Group 提供命名管理缓存/填充缓存的能力
type Group struct {
	name           string
	cache          *cache
	retriever      Retriever
	server         Picker
	flight         *singlefilght.Flight
	expirationTime int64
}

// NewGroup 创建一个新的缓存空间,如果tp不是LRUK的话，最后一个参数k无所谓填什么
func NewGroup(name string, maxBytes int64, retriever Retriever, tp string, expirationTime int64, k int) *Group {
	if retriever == nil {
		panic("Group retriever must be existed!")
	}
	g := &Group{
		name:           name,
		retriever:      retriever,
		flight:         &singlefilght.Flight{},
		expirationTime: expirationTime,
	}
	switch tp {
	case TYPE_FIFO:
		g.cache = newFIFOCache(maxBytes, nil)
	case TYPE_LRU:
		g.cache = newLRUCache(maxBytes, nil)
	case TYPE_LFU:
		g.cache = newLFUCache(maxBytes, nil)
	case TYPE_LRUK:
		g.cache = newLRUKCache(maxBytes, k, nil)
	case TYPE_2Q:
		g.cache = newtwoQCache(maxBytes, nil)
	}
	mu.Lock()
	groups[name] = g
	mu.Unlock()
	return g
}

// RegisterSvr 为 Group 注册 Server
func (g *Group) RegisterSvr(p Picker) {
	if g.server != nil {
		panic("group had been registered server")
	}
	g.server = p
}

// GetGroup 获取对应命名空间的缓存
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func DestroyGroup(name string) {
	g := GetGroup(name)
	if g != nil {
		svr := g.server.(*server)
		svr.Stop()
		delete(groups, name)
		log.Printf("Destroy cache [%s %s]", name, svr.addr)
	}
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key required")
	}
	if value, ok := g.cache.get(key); ok {
		log.Println("get cache hit")
		return value, nil
	}
	// cache missing, get it another way
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	view, err := g.flight.Fly(key, func() (interface{}, error) {
		if g.server != nil {
			if fetcher, ok := g.server.Pick(key); ok {
				bytes, err := fetcher.Fetch(g.name, key)
				if err == nil {
					return ByteView{b: cloneBytes(bytes)}, nil
				}
				log.Printf("fail to get *%s* from peer, %s.\n", key, err.Error())
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return view.(ByteView), err
	}
	return ByteView{}, err
}

// getLocally 本地向Retriever取回数据并填充缓存
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value, g.expirationTime)
	return value, nil
}

// populateCache 提供填充缓存的能力
func (g *Group) populateCache(key string, value ByteView, expirationTime int64) {
	g.cache.add(key, value, expirationTime)
}

// Remove 删除缓存中的数据
func (g *Group) Remove(key string) error {
	if key == "" {
		return fmt.Errorf("key required")
	}
	if ok := g.cache.remove(key); ok {
		log.Println("remove cache hit")
		return nil
	}
	// remove local cache missing, get it another way
	return g.loadremove(key)
}

// loadremove 删除远端节点的缓存
func (g *Group) loadremove(key string) error {
	_, err := g.flight.Fly(key, func() (interface{}, error) {
		if g.server != nil {
			if fetcher, ok := g.server.Pick(key); ok {
				err := fetcher.Remove(g.name, key)
				if err != nil {
					log.Printf("fail to remove *%s* from peer, %s.\n", key, err.Error())
				}
				return nil, nil
			}
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}

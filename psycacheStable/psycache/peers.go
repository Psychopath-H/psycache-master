package psycache

import (
	cacheALg "github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm"
)

// peers 模块

// Picker 定义了获取分布式节点的能力
type Picker interface {
	Pick(key string) (Fetcher, bool)
}

// Fetcher 定义了从远端获取缓存的能力
// 所以每个Peer应实现这个接口
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
	Remove(group string, key string) error
}

type Cache interface {
	Get(key string) (value cacheALg.Lengthable, ok bool)
	Add(key string, value cacheALg.Lengthable, expirationTime int64)
	Remove(key string) (ok bool)
	Contains(key string) (ok bool)
}

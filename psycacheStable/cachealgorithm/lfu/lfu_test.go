package lfu

import (
	cache "github.com/Psychopath-H/psycache-master/psycacheStable/cachealgorithm"
	"reflect"
	"testing"
	"time"
)

type String string

func (d String) Len() int {
	return len(d)
}

// 生成当前时间 + 2秒
func initTime() int64 {
	return time.Now().UnixNano()/1e6 + 2000
}

func TestGet(t *testing.T) {
	initTime := initTime()
	lfu := New(int64(100), nil)
	lfu.Add("key1", String("1"), initTime)
	if v, ok := lfu.Get("key1"); !ok || string(v.(String)) != "1" {
		t.Fatalf("cache hit key1=1 failed")
	}
	if _, ok := lfu.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
	lfu.Add("key2", String("2"), initTime)
	lfu.Add("key3", String("3"), initTime)
	lfu.Get("key1")
	lfu.Get("key2")
	lfu.Get("key2")
	lfu.Remove("key3")
	if lfu.minFre != 3 {
		t.Fatalf("func removeElement has something wrong")
	}
}

func TestRemoveoldest(t *testing.T) {
	initTime := initTime()
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lfu := New(int64(cap), nil)
	lfu.Add(k2, String(v1), initTime)
	lfu.Add(k1, String(v2), initTime)
	lfu.Get(k2)
	lfu.Add(k3, String(v3), initTime)

	if _, ok := lfu.Get("key1"); ok || lfu.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	initTime := initTime()
	keys := make([]string, 0)
	callback := cache.OnEliminated(func(key string, value cache.Lengthable) {
		keys = append(keys, key)
	})
	lfu := New(int64(10), callback)
	lfu.Add("key1", String("123456"), initTime)
	lfu.Add("k2", String("k2"), initTime)
	lfu.Add("k3", String("k3"), initTime)
	lfu.Add("k4", String("k4"), initTime)

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestAdd(t *testing.T) {
	initTime := initTime()
	lfu := New(int64(11), nil)
	lfu.Add("key", String("1"), initTime)
	lfu.Add("key1", String("111"), initTime)

	if lfu.nowcap != 11 {
		t.Fatal("expected 11 but got", lfu.nowcap)
	}
}

func TestRemove(t *testing.T) {
	initTime := initTime()
	lfu := New(int64(6), nil)
	lfu.Add("key", String("1"), initTime)
	lfu.Add("key2", String("2"), initTime)
	lfu.Remove("key2")

	if _, ok := lfu.Get("key2"); ok {
		t.Fatal("expected nonexist but got key2")
	}
}

func TestContains(t *testing.T) {
	initTime := initTime()
	lfu := New(int64(6), nil)
	lfu.Add("key", String("1"), initTime)

	if !lfu.Contains("key") {
		t.Fatal("expected got key but nonexist")
	}
}

func TestExpirationTime(t *testing.T) {
	initTime := initTime()
	lfu := New(int64(100), nil)
	lfu.Add("key1", String("1"), initTime)
	lfu.Add("key2", String("2"), time.Now().UnixNano()/1e6+4000)
	lfu.Get("key2")
	time.Sleep(2 * time.Second)
	if _, ok := lfu.Get("key1"); ok {
		t.Fatal("expected key1 expired but got key")
	}
	if lfu.nowcap != 5 || lfu.minFre != 2 {
		t.Fatal("func removeElement has something wrong")
	}
}

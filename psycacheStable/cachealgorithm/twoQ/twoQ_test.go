package twoQ

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
	twoQ := New(int64(10), nil)
	twoQ.Add("key1", String("1234"), initTime)
	if v, ok := twoQ.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := twoQ.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
	time.Sleep(3 * time.Second)
	if _, ok := twoQ.Get("key1"); ok {
		t.Fatalf("key1 should be expired")
	}
}

func TestOnEvicted(t *testing.T) {
	initTime := initTime()
	keys := make([]string, 0)
	callback := cache.OnEliminated(func(key string, value cache.Lengthable) {
		keys = append(keys, key)
	})

	twoQ := New(int64(10), callback)
	twoQ.Add("key1", String("123456"), initTime)
	twoQ.Add("k2", String("k2"), initTime)
	twoQ.Add("k3", String("k3"), initTime)
	twoQ.Add("k4", String("k4"), initTime)

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestAdd(t *testing.T) {
	initTime := initTime()
	twoQ := New(int64(7), nil)
	twoQ.Add("key1", String("1"), initTime)
	twoQ.Add("key1", String("111"), initTime)
	twoQ.Add("key2", String("2"), initTime)
	if twoQ.lru.Len() != 1 || twoQ.FIFO.Len() != 1 {
		t.Fatal("func Add has something wrong")
	}
}

func TestRemove(t *testing.T) {
	initTime := initTime()
	twoQ := New(int64(6), nil)
	twoQ.Add("key", String("1"), initTime)
	twoQ.Add("key2", String("2"), initTime)
	twoQ.Remove("key2")

	if _, ok := twoQ.Get("key2"); ok {
		t.Fatal("expected nonexist but got key2")
	}
}

func TestContains(t *testing.T) {
	initTime := initTime()
	twoQ := New(int64(6), nil)
	twoQ.Add("key", String("1"), initTime)

	if !twoQ.Contains("key") {
		t.Fatal("expected got key but nonexist")
	}
}

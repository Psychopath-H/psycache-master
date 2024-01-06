package lruk

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
	lruk := New(int64(10), 2, nil)
	lruk.Add("key1", String("1234"), initTime)
	if v, ok := lruk.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lruk.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
	time.Sleep(3 * time.Second)
	if _, ok := lruk.Get("key1"); ok {
		t.Fatalf("key1 should be expired")
	}
	if _, ok := lruk.historyVisited["key1"]; ok {
		t.Fatalf("key1 should be expired and should not be found in map")
	}

}

func TestOnEvicted(t *testing.T) {
	initTime := initTime()
	keys := make([]string, 0)
	callback := cache.OnEliminated(func(key string, value cache.Lengthable) {
		keys = append(keys, key)
	})

	lruk := New(int64(10), 2, callback)
	lruk.Add("key1", String("123456"), initTime)
	lruk.Add("k2", String("k2"), initTime)
	lruk.Add("k3", String("k3"), initTime)
	lruk.Add("k4", String("k4"), initTime)

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestAdd(t *testing.T) {
	initTime := initTime()
	lruk := New(int64(20), 3, nil)
	lruk.Add("key1", String("1"), initTime)
	lruk.Add("key1", String("111"), initTime)
	lruk.Add("key1", String("11"), initTime)
	lruk.Add("key2", String("2"), initTime)
	if lruk.datalru.Len() != 1 || lruk.historylru.Len() != 1 {
		t.Fatal("func Add has something wrong")
	}
}

func TestRemove(t *testing.T) {
	initTime := initTime()
	lruk := New(int64(10), 2, nil)
	lruk.Add("key1", String("1"), initTime)
	lruk.Add("key2", String("2"), initTime)
	lruk.Remove("key2")

	if _, ok := lruk.Get("key2"); ok {
		t.Fatal("expected nonexist but got key2")
	}
	if _, ok := lruk.historyVisited["key2"]; ok {
		t.Fatal("key2 should not exist but got")
	}
}

func TestContains(t *testing.T) {
	initTime := initTime()
	lruk := New(int64(6), 2, nil)
	lruk.Add("key", String("1"), initTime)

	if !lruk.Contains("key") || lruk.historyVisited["key"] != 1 {
		t.Fatal("expected got key but nonexist")
	}
}

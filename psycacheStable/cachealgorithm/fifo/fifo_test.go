package fifo

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
	FIFO := New(int64(10), nil)
	FIFO.Add("key1", String("1234"), initTime)
	if v, ok := FIFO.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := FIFO.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
	time.Sleep(3 * time.Second)
	if _, ok := FIFO.Get("key1"); ok {
		t.Fatalf("key1 should be expired")
	}
}

func TestRemoveoldest(t *testing.T) {
	initTime := initTime()
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	FIFO := New(int64(cap), nil)
	FIFO.Add(k1, String(v1), initTime)
	FIFO.Add(k2, String(v2), initTime)
	FIFO.Add(k3, String(v3), initTime)

	if _, ok := FIFO.Get("key1"); ok || FIFO.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	initTime := initTime()
	keys := make([]string, 0)

	callback := cache.OnEliminated(func(key string, value cache.Lengthable) {
		keys = append(keys, key)
	})
	FIFO := New(int64(10), callback)
	FIFO.Add("key1", String("123456"), initTime)
	FIFO.Add("k2", String("k2"), initTime)
	FIFO.Add("k3", String("k3"), initTime)
	FIFO.Add("k4", String("k4"), initTime)

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestAdd(t *testing.T) {
	initTime := initTime()
	FIFO := New(int64(6), nil)
	FIFO.Add("key", String("1"), initTime)
	FIFO.Add("key", String("111"), initTime)

	if FIFO.nowcap != int64(len("key")+len("111")) {
		t.Fatal("expected 6 but got", FIFO.nowcap)
	}
}

func TestRemove(t *testing.T) {
	initTime := initTime()
	FIFO := New(int64(6), nil)
	FIFO.Add("key", String("1"), initTime)
	FIFO.Add("key2", String("2"), initTime)
	FIFO.Remove("key2")

	if _, ok := FIFO.Get("key2"); ok {
		t.Fatal("expected nonexist but got key2")
	}
}

func TestContains(t *testing.T) {
	initTime := initTime()
	FIFO := New(int64(6), nil)
	FIFO.Add("key", String("1"), initTime)

	if !FIFO.Contains("key") {
		t.Fatal("expected got key but nonexist")
	}
}

func TestExpirationTime(t *testing.T) {
	initTime := initTime()
	FIFO := New(int64(6), nil)
	FIFO.Add("key", String("111"), initTime)
	time.Sleep(3 * time.Second)
	if _, ok := FIFO.Get("key"); ok {
		t.Fatal("expected key expired but got key")
	}
}

package psycache

import (
	"fmt"
	"log"
	"sync"
	"testing"
)

func TestGet(t *testing.T) {
	mysql := map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	loadCounts := make(map[string]int, len(mysql))

	g := NewGroup("scores", 2<<10, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), TYPE_LFU, initTime(), 2)

	for k, v := range mysql {
		if view, err := g.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %s", k)
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := g.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	} else {
		log.Println(err)
	}
}

func TestAll(t *testing.T) {
	// 模拟MySQL数据库 用于peanutcache从数据源获取值
	var mysql = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	// 新建cache实例
	group := NewGroup("scores", 2<<10, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), TYPE_LFU, initTime(), 2)
	// New一个服务实例
	var addr string = "localhost:9999"
	svr, err := NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	// 设置同伴节点IP(包括自己)
	// todo: 这里的peer地址从etcd获取(服务发现)
	svr.SetPeers(addr)
	// 将服务与cache绑定 因为cache和server是解耦合的
	group.RegisterSvr(svr)
	log.Println("psycache is running at", addr)
	// 启动服务(注册服务至etcd/计算一致性哈希...)
	go func() {
		// Start将不会return 除非服务stop或者抛出error
		err = svr.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()
	// 发出几个Get请求
	var wg sync.WaitGroup
	wg.Add(4)
	go GetTomScore(group, &wg)
	go GetTomScore(group, &wg)
	go GetTomScore(group, &wg)
	go GetTomScore(group, &wg)
	wg.Wait()
}

func GetTomScore(group *Group, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("get Tom...")
	view, err := group.Get("Tom")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(view.String())
}

package main

import (
	"flag"
	"fmt"
	"log"
	"psycache"
	"sync"
	"time"
)

const (
	TYPE_FIFO = "FIFO"
	TYPE_LRU  = "lru"
	TYPE_LFU  = "lfu"
	TYPE_LRUK = "lruk"
	TYPE_2Q   = "twoQ"
)

// 生成当前时间 + 20秒
func initTime() int64 {
	return time.Now().UnixNano()/1e6 + 20000
}

// startCacheServer 在一个具体节点开启一个缓存服务
func startCacheServer(addr string, addrs []string, group *psycache.Group, wg *sync.WaitGroup) {
	defer wg.Done()
	// New一个服务实例
	svr, err := psycache.NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	// 设置同伴节点IP(包括自己)
	// todo: 这里的peer地址从etcd获取(服务发现)
	svr.SetPeers(addrs...)
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
}

func main() {
	// 模拟MySQL数据库 用于psycache从数据源获取值
	var mysql = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	// 新建cache实例
	group := psycache.NewGroup("scores", 2<<10, psycache.RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), TYPE_LFU, initTime(), 2)

	addrMap := map[int]string{
		8001: "localhost:8001",
		8002: "localhost:8002",
		8003: "localhost:8003",
	}

	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "psycache server port")        //命令行参数解析 这里解析的是端口
	flag.BoolVar(&api, "visit", false, "make sure if we visit Tom") //命令行参数解析 这里解析的是是否访问Tom
	flag.Parse()

	var addrs []string //addrs == []string{"localhost:8001","localhost:8002","localhost:8003"}
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go startCacheServer(addrMap[port], addrs, group, &wg)
	wg.Wait()

	//测试能否成功并发读取缓存
	wg.Add(15)
	if api {
		for i := 0; i < 5; i++ {
			go GetTomScore(group, &wg)
		}
		for i := 0; i < 5; i++ {
			go GetJackScore(group, &wg)
		}
		for i := 0; i < 5; i++ {
			go GetSamScore(group, &wg)
		}
	}
	wg.Wait()

	//测试能否成功删除缓存
	//wg.Add(2)
	//if api {
	//	GetTomScore(group, &wg)
	//	RemoveTomScore(group, &wg)
	//}
	//wg.Wait()

}

func RemoveTomScore(group *psycache.Group, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("remove Tom...")
	err := group.Remove("Tom")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println("Remove Tom's score successfully")
}

func GetTomScore(group *psycache.Group, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("get Tom...")
	view, err := group.Get("Tom")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println(view.String())
}

func GetJackScore(group *psycache.Group, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("get Jack...")
	view, err := group.Get("Jack")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println(view.String())
}

func GetSamScore(group *psycache.Group, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("get Sam...")
	view, err := group.Get("Sam")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println(view.String())
}

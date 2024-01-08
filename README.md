# PsyCache
PsyCache是仿照groupcache实现的一个分布式缓存系统。
主要工作如下：
- 提供多种缓存策略可选(FIFO、LRU、LFU、LRU-K、twoQ)；缓存具有过期机制，超时自动清理缓存；
- 使用一致性哈希算法，实现分布式节点的动态扩缩容，规避数据倾斜问题；
- 高并发访问缓存时，使用singleflight机制防止缓存击穿；
- 使用GRPC进行节点间通信，可以进行远端增加和删除缓存，通信数据格式选用Protobuf，提高通信效率；
- 使用etcd作为节点的服务注册与发现，实现节点的动态管理
- 使用jaeger进行分布式节点的链路追踪，能够观测到具体节点间的调用过程

#  Prerequisites
- **Golang** 1.21 or later
- **Etcd** v3.4.27 or later
-  **gRPC-go** v1.38.0 or later
-  **protobuf** v1.26.0 or later
-  **jaeger** v2.30.0 or later

# Installation

借助于 [Go module] 的支持(Go 1.11+), 只需要添加如下引入

import "github.com/Psychopath-H/psycache-master"

接着只需 `go [build|run|test]` 将会自动导入依赖.

或者，你也可以直接安装 `psycache-master` 包, 运行一下命令

$ go get -u github.com/Psychopath-H/psycache-master

# Usage

```go
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
   flag.BoolVar(&api, "visit", false, "make sure if we visit cache") //命令行参数解析 这里解析的是是否访问缓存  
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
 //    GetTomScore(group, &wg)
 //    RemoveTomScore(group, &wg)
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
```

在运行之前，你应当在本地开启etcd服务器,并在本地192.168.100.100开启jaeger服务，并分别在三个客户端使用以下命令:

```go
$ go run main.go -port=8002
2024/01/08 12:16:37 psycache is running at localhost:8003
...
2024/01/08 12:16:37 [localhost:8002] register service ok
2024/01/08 12:17:15 [psycache_svr localhost:8002] Recv RPC Request - (scores)/(Jack)
2024/01/08 12:17:15 [psycache_svr localhost:8002] Recv RPC Request - (scores)/(Sam)
2024/01/08 12:17:15 ooh! pick myself, I am localhost:8002
2024/01/08 12:17:15 [Mysql] search key Jack
2024/01/08 12:17:15 ooh! pick myself, I am localhost:8002
2024/01/08 12:17:15 [Mysql] search key Sam
2024/01/08 12:17:15 Reporting span 5e61d3e45c8bb022:5e61d3e45c8bb022:0000000000000000:1
2024/01/08 12:17:15 Reporting span 2038cb7010de8908:2038cb7010de8908:0000000000000000:1

```

```go
$ go run main.go -port=8003
2024/01/08 12:17:02 psycache is running at localhost:8003
2024/01/08 12:17:02 debug logging disabled
2024/01/08 12:17:02 Initializing logging reporter
2024/01/08 12:17:02 debug logging disabled
2024/01/08 12:17:03 [localhost:8003] register service ok
```

```go
$ go run main.go -port=8001 -visit=true
2024/01/08 14:54:25 psycache is running at localhost:8001
2024/01/08 14:54:25 get Tom... 
2024/01/08 14:54:25 get Jack...
2024/01/08 14:54:25 get Tom... 
2024/01/08 14:54:25 get Jack...
...
2024/01/08 14:54:25 debug logging disabled
...
2024/01/08 14:54:25 ooh! pick myself, I am localhost:8001
2024/01/08 14:54:25 [Mysql] search key Tom
2024/01/08 14:54:25 630
2024/01/08 14:54:25 get packet succeed!
2024/01/08 14:54:25 630
...
2024/01/08 14:54:25 [cache localhost:8001] pick remote peer: localhost:8002
2024/01/08 14:54:25 630
2024/01/08 14:54:25 [cache localhost:8001] pick remote peer: localhost:8002
2024/01/08 14:54:25 [localhost:8001] register service ok
2024/01/08 14:54:25 589
2024/01/08 14:54:25 567
2024/01/08 14:54:25 get packet succeed!
2024/01/08 14:54:25 567
...
2024/01/08 14:54:25 589
2024/01/08 14:54:25 get packet succeed!
2024/01/08 14:54:25 589
...
```

package registry

// register模块提供服务Service注册至etcd的能力

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"log"
	"time"
)

var (
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// etcdAdd 在租赁模式添加一对kv至etcd
func etcdAdd(c *clientv3.Client, lid clientv3.LeaseID, service string, addr string) error {
	//该函数用于创建一个 Endpoints Manager（端点管理器）对象。这个管理器用于管理服务的端点（Endpoints）。
	//c 参数通常包含了一些配置信息，例如服务发现配置等，而 service 参数是服务的名称，表示你要管理的服务。
	em, err := endpoints.NewManager(c, service)
	if err != nil {
		return err
	}
	//将服务的键值对注册到etcd中
	//这里的service+"/"+addr是服务的键。endpoints.Endpoint{Addr: addr}一般为localhost:9999是服务的值，通常是服务地址和端口
	return em.AddEndpoint(c.Ctx(), service+"/"+addr, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(lid))
}

// Register 注册一个服务至etcd
// 注意 Register将不会return 如果没有error的话
func Register(service string, addr string, stop chan error) error {
	// 创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("create etcd client failed: %v", err)
	}
	defer cli.Close()
	// 创建一个租约 配置5秒过期
	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("create lease failed: %v", err)
	}
	leaseId := resp.ID
	// 注册服务
	err = etcdAdd(cli, leaseId, service, addr) //注册服务需要用含有Etcd服务器地址的客户端，租约时间，服务名称，服务地址
	if err != nil {
		return fmt.Errorf("add etcd record failed: %v", err)
	}
	// 设置服务心跳检测
	ch, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}

	log.Printf("[%s] register service ok\n", addr)
	for {
		select {
		case err := <-stop:
			if err != nil {
				log.Println(err)
			}
			return err
		case <-cli.Ctx().Done():
			log.Println("service closed")
			return nil
		case _, ok := <-ch:
			// 监听租约
			if !ok {
				log.Println("keep alive channel closed")
				_, err := cli.Revoke(context.Background(), leaseId)
				return err
			}
			//log.Printf("Recv reply from service: %s/%s, ttl:%d", service, addr, resp.TTL)
		}
	}
}

package psycache

import (
	"context"
	"fmt"
	"github.com/Psychopath-H/psycache-master/psycacheStable/consistenthash"
	"github.com/Psychopath-H/psycache-master/psycacheStable/registry"
	"github.com/Psychopath-H/psycache-master/psycacheStable/tracer"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"log"
	"net"
	pb "psycachepb"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// server 模块为psycache之间提供通信能力
// 这样部署在其他机器上的cache可以通过访问server获取缓存
// 至于找哪台主机 那是一致性哈希的工作了

const (
	defaultAddr     = "127.0.0.1:6324"
	defaultReplicas = 50
)

var (
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// server 和 Group 是解耦合的 所以server要自己实现并发控制
type server struct {
	pb.UnimplementedPsyCacheServer

	addr       string     // format: ip:port
	status     bool       // true: running false: stop
	stopSignal chan error // 通知registry revoke服务
	mu         sync.Mutex
	consHash   *consistenthash.Consistency
	clients    map[string]*client
}

// NewServer 创建cache的svr 若addr为空 则使用defaultAddr
func NewServer(addr string) (*server, error) {
	if addr == "" {
		addr = defaultAddr
	}
	if !validPeerAddr(addr) {
		return nil, fmt.Errorf("invalid addr %s, it should be x.x.x.x:port", addr)
	}
	return &server{addr: addr}, nil
}

// Get 实现PsyCache service的Get接口
func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	group, key := in.GetGroup(), in.GetKey()
	resp := &pb.GetResponse{}

	log.Printf("[psycache_svr %s] Recv RPC Request - (%s)/(%s)", s.addr, group, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("group not found")
	}
	view, err := g.Get(key)
	if err != nil {
		log.Printf("get %s succeed", key)
		return resp, err
	}
	resp.Value = view.ByteSlice()
	return resp, nil
}

// Remove 实现PsyCache service的Remove接口
func (s *server) Remove(ctx context.Context, in *pb.GetRequest) (*pb.RemoveResponse, error) {
	group, key := in.GetGroup(), in.GetKey()
	resp := &pb.RemoveResponse{}

	log.Printf("[psycache_svr %s] Recv RPC Request - (%s)/(%s)", s.addr, group, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("group not found")
	}
	err := g.Remove(key)
	if err != nil {
		log.Printf("remove %s succeed", key)
		return resp, err
	}
	return resp, nil
}

// Start 启动cache服务
func (s *server) Start() error {
	s.mu.Lock()
	if s.status == true {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}
	// -----------------启动服务----------------------
	// 1. 设置status为true 表示服务器已在运行
	// 2. 初始化stop channal,这用于通知registry stop keep alive
	// 3. 初始化tcp socket并开始监听
	// 4. 注册rpc服务至grpc 这样grpc收到request可以分发给server处理
	// 5. 将自己的服务名/Host地址注册至etcd 这样client可以通过etcd
	//    获取服务Host地址 从而进行通信。这样的好处是client只需知道服务名
	//    以及etcd的Host即可获取对应服务IP 无需写死至client代码中
	// ----------------------------------------------
	s.status = true
	s.stopSignal = make(chan error)

	port := strings.Split(s.addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	Tracer, closer, err := tracer.CreateTracer("ScoreTracer", &config.SamplerConfig{
		Type:  jaeger.SamplerTypeConst,
		Param: 1,
	}, &config.ReporterConfig{
		LogSpans:          true,
		CollectorEndpoint: "http://192.168.100.100:14268/api/traces",
	}, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(Tracer))))
	pb.RegisterPsyCacheServer(grpcServer, s) //注册RPC服务至GRPC

	// 注册服务至etcd
	go func() {
		// Register never return unless stop singnal received
		err := registry.Register("psycache", s.addr, s.stopSignal)
		if err != nil {
			log.Fatalf(err.Error())
		}
		// Close channel
		close(s.stopSignal)
		// Close tcp listen
		err = lis.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.Printf("[%s] Revoke service and close tcp socket ok.", s.addr)
	}()

	//log.Printf("[%s] register service ok\n", s.addr)
	s.mu.Unlock()

	if err := grpcServer.Serve(lis); s.status && err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

// SetPeers 将各个远端主机IP配置到Server里
// 这样Server就可以Pick他们了
// 注意: 此操作是*覆写*操作！
// 注意: peersIP必须满足 x.x.x.x:port的格式
func (s *server) SetPeers(peersAddr ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.consHash = consistenthash.New(defaultReplicas, nil)
	s.consHash.Register(peersAddr...)
	s.clients = make(map[string]*client)
	for _, peerAddr := range peersAddr {
		if !validPeerAddr(peerAddr) {
			panic(fmt.Sprintf("[peer %s] invalid address format, it should be x.x.x.x:port", peerAddr))
		}
		service := fmt.Sprintf("psycache/%s", peerAddr)
		s.clients[peerAddr] = NewClient(service)
	}
}

// Pick 根据一致性哈希选举出key应存放在的cache
// return false 代表从本地获取cache
func (s *server) Pick(key string) (Fetcher, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peerAddr := s.consHash.GetPeer(key)
	// Pick itself
	if peerAddr == s.addr {
		log.Printf("ooh! pick myself, I am %s\n", s.addr)
		return nil, false
	}
	log.Printf("[cache %s] pick remote peer: %s\n", s.addr, peerAddr)
	return s.clients[peerAddr], true
}

// Stop 停止server运行 如果server没有运行 这将是一个no-op
func (s *server) Stop() {
	s.mu.Lock()
	if s.status == false {
		s.mu.Unlock()
		return
	}
	s.stopSignal <- nil // 发送停止keepalive信号
	s.status = false    // 设置server运行状态为stop
	s.clients = nil     // 清空一致性哈希信息 有助于垃圾回收
	s.consHash = nil
	s.mu.Unlock()
}

// 测试Server是否实现了Picker接口
var _ Picker = (*server)(nil)

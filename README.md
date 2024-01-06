# PsyCache
PsyCache是仿照groupcache实现的一个分布式缓存系统，在此基础上使用GRPC进行节点间通信，使用etcd作为服务注册和服务发现，在服务端使用了限流和熔断等服务容错策略，并使用jaeger进行链路追踪看到不同节点间的调用关系。

#  Prerequisites
- **Golang** 1.21 or later
- **Etcd** v3.4.27 or later
-  **gRPC-go** v1.38.0 or later
-  **protobuf** v1.26.0 or later

# Installation

借助于 [Go module] 的支持(Go 1.11+), 只需要添加如下引入

import "github.com/Psychopath-H/psycache-master"

接着只需 `go [build|run|test]` 将会自动导入依赖.

或者，你也可以直接安装 `psycache-master` 包, 运行一下命令

$ go get -u github.com/Psychopath-H/psycache-master

# Usage

example/main.go 文件夹下给出了具体的使用例子

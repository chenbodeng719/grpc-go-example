package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"helloworld/helloworld"
	"helloworld/internal"
	"log"
	"testing"
	"time"
)

func TestResolver(t *testing.T) {
	internal.Start("127.0.0.1:1000")
	// etcd中注册5个服务
	internal.Start("127.0.0.1:1001")
	internal.Start("127.0.0.1:1002")

	cli, cerr := clientv3.NewFromURL("http://localhost:22379")
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		panic(fmt.Sprintf("etcd failed: %v", cerr))
		return
	}
	options := []grpc.DialOption{
		grpc.WithResolvers(etcdResolver),
		grpc.WithInsecure(),
	}
	conn, gerr := grpc.Dial("etcd:///hello", options...)

	if gerr != nil {
		panic(fmt.Sprintf("get server failed: %v", gerr))
		return
	}
	defer conn.Close()

	c := helloworld.NewGreeterClient(conn)

	// 进行十次数据请求
	for i := 0; i < 10; i++ {
		resp, err := c.SayHello(context.Background(), &helloworld.HelloRequest{Name: "abc"})
		if err != nil {
			t.Fatalf("say hello failed %v", err)
		}
		log.Println(resp.Message)
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(10 * time.Second)
}

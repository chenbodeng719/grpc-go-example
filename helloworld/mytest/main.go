package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"helloworld/endpoint"
	"helloworld/helloworld"
	"helloworld/internal"
	"log"
	"strconv"
	"sync"
	"time"
)

type MyItem struct {
	gserver  *grpc.Server
	register *endpoint.Register
}

func main() {

	// etcd中注册5个服务

	etcdAddrs := []string{"127.0.0.1:22379"}
	slist := []MyItem{}
	app := "hello"
	wg := new(sync.WaitGroup)
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(thei int) {
			defer wg.Done()
			tmp := fmt.Sprintf("127.0.0.1:%s", strconv.Itoa(10000+thei))
			//server, _ := internal.StartWithReg(etcdAddrs, app, tmp)
			log.Println(tmp)
			server, reg, _ := internal.StartWithReg(etcdAddrs, app, tmp)
			slist = append(slist, MyItem{server, reg})
			//defer server.Stop()
		}(i)
	}

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
			log.Fatalf("say hello failed %v", err)
		}
		log.Println(resp.Message)
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(10 * time.Second)
	for _, v := range slist {
		v.gserver.Stop()
		v.register.Stop()
	}
}

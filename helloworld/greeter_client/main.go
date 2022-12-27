/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	resolver "go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	pb "helloworld/helloworld"
	"log"
	"time"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

//func main() {
//	flag.Parse()
//	// Set up a connection to the server.
//	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		log.Fatalf("did not connect: %v", err)
//	}
//	defer conn.Close()
//	c := pb.NewGreeterClient(conn)
//
//	// Contact the server and print out its response.
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	defer cancel()
//	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
//	if err != nil {
//		log.Fatalf("could not greet: %v", err)
//	}
//	log.Printf("Greeting: %s", r.GetMessage())
//
//	r, err = c.SayHelloAgain(ctx, &pb.HelloRequest2{Name: *name})
//	if err != nil {
//		log.Fatalf("could not greet: %v", err)
//	}
//	log.Printf("Greeting: %s", r.GetMessage())
//}

var etcdAddrs = []string{"127.0.0.1:22379"}

func main() {

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
	c := pb.NewGreeterClient(conn)
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	//r, err = c.SayHelloAgain(ctx, &pb.HelloRequest2{Name: *name})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %s", r.GetMessage())
}

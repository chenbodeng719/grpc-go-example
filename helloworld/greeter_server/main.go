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

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"helloworld/internal"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	//pb "google.golang.org/grpc/examples/helloworld/helloworld"
	pb "helloworld/helloworld"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *server) SayHelloAgain(ctx context.Context, in *pb.HelloRequest2) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello again " + in.GetName()}, nil
}

//
//func main() {
//
//	flag.Parse()
//	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
//	if err != nil {
//		log.Fatalf("failed to listen: %v", err)
//	}
//	s := grpc.NewServer()
//	pb.RegisterGreeterServer(s, &server{})
//	log.Printf("server listening at %v", lis.Addr())
//	if err := s.Serve(lis); err != nil {
//		log.Fatalf("failed to serve: %v", err)
//	}
//}
const (
	app         = "hello"
	grpcAddress = "127.0.0.1:8083"
)

func main() {
	addrs := []string{"127.0.0.1:22379"}
	etcdRegister := internal.NewRegister(addrs, zap.NewNop())
	node := internal.Server{
		Name: app,
		Addr: grpcAddress,
	}

	server, err := Start()
	if err != nil {
		panic(fmt.Sprintf("start server failed : %v", err))
	}
	if _, err := etcdRegister.Register(node, 3600); err != nil {
		panic(fmt.Sprintf("server register failed: %v", err))
	}
	fmt.Println("service started listen on", grpcAddress)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			server.Stop()
			etcdRegister.Stop()
			time.Sleep(1 * time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func Start() (*grpc.Server, error) {
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return s, nil
}

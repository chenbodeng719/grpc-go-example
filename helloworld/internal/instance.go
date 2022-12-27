package internal

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"helloworld/endpoint"
	pb "helloworld/helloworld"
	"log"
	"net"
)

type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`    //服务地址
	Version string `json:"version"` //服务版本
	Weight  int64  `json:"weight"`  //服务权重
}

func BuildPrefix(info Server) string {
	return fmt.Sprintf("/%s/%s/", info.Name, info.Addr)
}

type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v ", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *server) SayHelloAgain(ctx context.Context, in *pb.HelloRequest2) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello again " + in.GetName()}, nil
}

func Start(grpcAddress string) (*grpc.Server, error) {
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

func StartWithReg(etcdAddrs []string, app string, grpcAddress string) (*grpc.Server, *endpoint.Register, error) {
	node := endpoint.Server{
		Name: app,
		Addr: grpcAddress,
	}
	etcdRegister := endpoint.NewRegister(etcdAddrs, zap.NewNop())
	s, err := Start(node.Addr)
	if err != nil {
		panic(fmt.Sprintf("start server failed : %v", err))
	}
	if _, err := etcdRegister.Register(node, 3600); err != nil {
		panic(fmt.Sprintf("server register failed: %v", err))
	}
	fmt.Println("service started listen on", node.Addr)
	return s, etcdRegister, nil
}

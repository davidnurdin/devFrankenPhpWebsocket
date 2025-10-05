package mygrpcserver

import (
	"context"
	"fmt"

	pb "davidnurdin.com/mygrpcserver/helloworld"
	"github.com/dunglas/frankenphp"
	phpGrpc "github.com/davidnurdin/frankenphp-websocket"
	"github.com/go-viper/mapstructure/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func init() {
	phpGrpc.RegisterGrpcServerFactory(func() *grpc.Server {
		s := grpc.NewServer()
		pb.RegisterGreeterServer(s, &server{})
		reflection.Register(s)

		return s
	})
}

type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("the Name field is required")
	}

    // Convert the request to a map[string]any
	var phpRequest map[string]any
	if err := mapstructure.Decode(in, &phpRequest); err != nil {
		return nil, err
	}

    // Call the PHP code, pass the map as a PHP associative array
	phpResponse := phpGrpc.HandleRequest(phpRequest)

    // Convert the PHP response (a map) back to a HelloReply struct
	var response pb.HelloReply
	if err := mapstructure.Decode(phpResponse.(frankenphp.AssociativeArray).Map, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

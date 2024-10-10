package main

import (
	"context"
	pb "github.com/rafapcarvalho/logtracer/examples/grpc-example1/proto"
	"github.com/rafapcarvalho/logtracer/pkg/logtracer"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedGreeterServer
	lt *logtracer.LogTracer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	s.lt.SrvcLog.Info(ctx, "Received: "+in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	cfg := logtracer.Config{
		ServiceName:   "grpc-server",
		LogFormat:     "json",
		EnableTracing: true,
		OTLPEndpoint:  "localhost:4318",
	}

	lt, err := logtracer.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create LogTracer: %v", err)
	}
	defer lt.Shutdown(context.Background())

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		lt.InitLog.Error(context.Background(), "Failed to listen", "error", err)
		return
	}

	unaryInterceptor, streamInterceptor := logtracer.OTELGRPCServerInterceptor()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			lt.UnaryServerInterceptor(),
			unaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			lt.StreamServerInterceptor(),
			streamInterceptor,
		),
	)

	pb.RegisterGreeterServer(s, &server{lt: lt})

	lt.SrvcLog.Info(context.Background(), "Server listening", "addres", ":50051")
	if err := s.Serve(lis); err != nil {
		lt.SrvcLog.Error(context.Background(), "Failed to serve", "error", err)
	}
}

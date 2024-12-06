package main

import (
	"context"
	pb "github.com/rafapcarvalho/logtracer/examples/grpc-example1/proto"
	logger "github.com/rafapcarvalho/logtracer/pkg/logtracer"
	"google.golang.org/grpc"
	"net"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	ctx = logger.StartSpan(ctx, "SayHello")
	defer logger.EndSpan(ctx)

	logger.SrvcLog.Info(ctx, "handling request", "name", in.Name)
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	cfg := logger.Config{
		ServiceName:   "grpc-server",
		LogFormat:     "json",
		EnableTracing: true,
		OTLPEndpoint:  "localhost:4318",
	}

	logger.InitLogger(cfg)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.InitLog.Error(context.Background(), "Failed to listen", "error", err)
		return
	}

	logTracer := &logger.LogTracer{}
	unaryInterceptor, _ := logger.OTELGRPCServerInterceptor()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logTracer.UnaryServerInterceptor(),
			unaryInterceptor,
		),
	)

	pb.RegisterGreeterServer(s, &server{})
	logger.SrvcLog.Info(context.Background(), "Server starting", "port", 50051)

	if err := s.Serve(lis); err != nil {
		logger.SrvcLog.Error(context.Background(), "Failed to serve", "error", err)
	}
}

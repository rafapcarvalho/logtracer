package main

import (
	"context"
	pb "github.com/rafapcarvalho/logtracer/examples/grpc-example1/proto"
	logger "github.com/rafapcarvalho/logtracer/pkg/logtracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"time"
)

func main() {
	cfg := logger.Config{
		ServiceName:   "grpc-client",
		LogFormat:     "json",
		EnableTracing: true,
		OTLPEndpoint:  "localhost:4318",
	}

	logger.InitLogger(cfg)

	logTracer := &logger.LogTracer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(logTracer.UnaryClientInterceptor()),
	)
	if err != nil {
		logger.SrvcLog.Error(context.Background(), "Failed to connect to server", "error", err)
		return
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.SrvcLog.Error(context.Background(), "Failed to close connection", "error", err)
		}
	}(conn)

	c := pb.NewGreeterClient(conn)

	name := "World"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		logger.SrvcLog.Error(ctx, "Failed to call SayHello", "error", err)
		return
	}

	logger.SrvcLog.Info(ctx, "Greeting received", "message", r.GetMessage())
}

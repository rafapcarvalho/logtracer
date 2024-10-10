package main

import (
	"context"
	pb "github.com/rafapcarvalho/logtracer/examples/grpc-example1/proto"
	"github.com/rafapcarvalho/logtracer/pkg/logtracer"
	"google.golang.org/grpc"
	"log"
	"os"
	"time"
)

func main() {
	cfg := logtracer.Config{
		ServiceName:   "grpc-client",
		LogFormat:     "json",
		EnableTracing: true,
		OTLPEndpoint:  "localhost:4318",
	}

	lt, err := logtracer.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create LogTracer: %v", err)
	}
	defer lt.Shutdown(context.Background())

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(lt.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(lt.StreamClientInterceptor()),
	)
	if err != nil {
		lt.SrvcLog.Error(context.Background(), "Failed to connect to server", "error", err)
		return
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	name := "World"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		lt.SrvcLog.Error(ctx, "Failed to call SayHello", "error", err)
		return
	}

	lt.SrvcLog.Info(ctx, "Greeting received", "message", r.GetMessage())
}
